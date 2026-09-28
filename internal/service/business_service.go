package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"fishing-notes-admin-api/internal/assembler"
	"fishing-notes-admin-api/internal/audit"
	"fishing-notes-admin-api/internal/auth"
	"fishing-notes-admin-api/internal/errorsx"
	"fishing-notes-admin-api/internal/model"
	"fishing-notes-admin-api/internal/repository"
	"fishing-notes-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type BusinessService struct {
	repo  *repository.BusinessRepository
	audit *repository.AuditRepository
}

func NewBusinessService(repo *repository.BusinessRepository, auditRepos ...*repository.AuditRepository) *BusinessService {
	var auditRepo *repository.AuditRepository
	if len(auditRepos) > 0 {
		auditRepo = auditRepos[0]
	}
	return &BusinessService{repo: repo, audit: auditRepo}
}

func (s *BusinessService) DashboardOverview(ctx context.Context) (*types.DashboardOverviewResponse, error) {
	if s == nil || s.repo == nil {
		return nil, errorsx.Internal("business repository is not configured")
	}
	result, err := s.repo.DashboardOverview(ctx)
	if err != nil {
		return nil, errorsx.Internal("failed to query dashboard overview")
	}
	return &types.DashboardOverviewResponse{
		Users: result.Users, FishingSpots: result.FishingSpots, PendingFishingSpots: result.PendingFishingSpots,
		Articles: result.Articles, PublishedArticles: result.PublishedArticles, Species: result.Species,
		VisibleSpecies: result.VisibleSpecies, FishingRecords: result.FishingRecords,
		PublicFishingRecords: result.PublicFishingRecords, PendingCorrections: result.PendingCorrections,
		PendingMediaAssets: result.PendingMediaAssets, PendingPublicReports: result.PendingPublicReports,
		PendingReviewTasks: result.PendingReviewTasks,
	}, nil
}

func (s *BusinessService) ListAuditLogs(ctx context.Context, req *types.ListAuditLogsRequest) (*types.ListAuditLogsResponse, error) {
	if s == nil || s.audit == nil {
		return nil, errorsx.Internal("audit repository is not configured")
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	items, total, err := s.audit.List(ctx, repository.AuditLogFilter{
		ResourceType: strings.TrimSpace(req.ResourceType), ResourceID: req.ResourceID,
		Action: strings.TrimSpace(req.Action), Offset: int((page - 1) * pageSize), Limit: int(pageSize),
	})
	if err != nil {
		return nil, errorsx.Internal("failed to query audit logs")
	}
	list := make([]types.AuditLogItem, 0, len(items))
	for _, item := range items {
		detail := json.RawMessage(item.DetailJSON)
		if !json.Valid(detail) {
			detail = json.RawMessage(`{}`)
		}
		list = append(list, types.AuditLogItem{
			ID: item.ID, ActorUserID: item.ActorUserID, Action: item.Action, ResourceType: item.ResourceType,
			ResourceID: item.ResourceID, Detail: detail, CreatedAt: item.CreatedAt.Format(time.RFC3339),
		})
	}
	return &types.ListAuditLogsResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// ExportAuditLogs 生成当前筛选条件下的操作日志 CSV。同步导出限制为最近 10000 条，
// 超出限制时由调用方通过 truncated 标记提示用户改用更窄的筛选条件。
func (s *BusinessService) ExportAuditLogs(ctx context.Context, req *types.ListAuditLogsRequest) ([]byte, bool, error) {
	if s == nil || s.audit == nil {
		return nil, false, errorsx.Internal("audit repository is not configured")
	}
	const maxRows = 10000
	items, total, err := s.audit.List(ctx, repository.AuditLogFilter{
		ResourceType: strings.TrimSpace(req.ResourceType), ResourceID: req.ResourceID,
		Action: strings.TrimSpace(req.Action), Offset: 0, Limit: maxRows,
	})
	if err != nil {
		return nil, false, errorsx.Internal("failed to export audit logs")
	}
	var buffer bytes.Buffer
	buffer.WriteString("\xEF\xBB\xBF")
	writer := csv.NewWriter(&buffer)
	if err := writer.Write([]string{"id", "actor_user_id", "action", "resource_type", "resource_id", "detail", "created_at"}); err != nil {
		return nil, false, errorsx.Internal("failed to encode audit logs")
	}
	for _, item := range items {
		actorID, resourceID := "", ""
		if item.ActorUserID != nil {
			actorID = strconv.FormatInt(*item.ActorUserID, 10)
		}
		if item.ResourceID != nil {
			resourceID = strconv.FormatInt(*item.ResourceID, 10)
		}
		if err := writer.Write([]string{
			strconv.FormatInt(item.ID, 10), actorID, item.Action, item.ResourceType, resourceID,
			item.DetailJSON, item.CreatedAt.Format(time.RFC3339),
		}); err != nil {
			return nil, false, errorsx.Internal("failed to encode audit logs")
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, false, errorsx.Internal("failed to encode audit logs")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "audit.log.export", ResourceType: "audit_log", Detail: map[string]any{"total": total, "exported": len(items)}}); err != nil {
		return nil, false, errorsx.Internal("failed to write audit log")
	}
	return buffer.Bytes(), total > int64(len(items)), nil
}

func (s *BusinessService) ListSpotCorrections(ctx context.Context, req *types.ListSpotCorrectionsRequest) (*types.ListSpotCorrectionsResponse, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	status := strings.TrimSpace(req.Status)
	if status != "" && status != "pending" && status != "approved" && status != "rejected" {
		return nil, errorsx.BadRequest("invalid correction status")
	}
	items, total, err := s.repo.ListSpotCorrections(ctx, repository.SpotCorrectionListFilter{Keyword: strings.TrimSpace(req.Keyword), Status: status, Offset: (page - 1) * pageSize, Limit: pageSize})
	if err != nil {
		return nil, errorsx.Internal("failed to query spot corrections")
	}
	list := make([]types.SpotCorrectionItem, 0, len(items))
	for _, item := range items {
		list = append(list, spotCorrectionItem(item))
	}
	return &types.ListSpotCorrectionsResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *BusinessService) GetSpotCorrection(ctx context.Context, correctionID int64) (*types.SpotCorrectionItem, error) {
	if correctionID <= 0 {
		return nil, errorsx.BadRequest("invalid spot correction id")
	}
	item, err := s.repo.FindSpotCorrectionByID(ctx, correctionID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("spot correction not found")
		}
		return nil, errorsx.Internal("failed to query spot correction")
	}
	result := spotCorrectionItem(*item)
	return &result, nil
}

func (s *BusinessService) ListPublicSpotApplications(ctx context.Context, req *types.ListPublicSpotApplicationsRequest) (*types.ListPublicSpotApplicationsResponse, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	status := strings.TrimSpace(req.Status)
	if status != "" && status != "pending" && status != "approved" && status != "rejected" {
		return nil, errorsx.BadRequest("invalid public spot application status")
	}
	items, total, err := s.repo.ListPublicSpotApplications(ctx, repository.PublicSpotApplicationListFilter{
		Keyword: strings.TrimSpace(req.Keyword), Status: status, Offset: (page - 1) * pageSize, Limit: pageSize,
	})
	if err != nil {
		return nil, errorsx.Internal("failed to query public spot applications")
	}
	list := make([]types.PublicSpotApplicationItem, 0, len(items))
	for _, item := range items {
		list = append(list, publicSpotApplicationItem(item))
	}
	return &types.ListPublicSpotApplicationsResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *BusinessService) GetPublicSpotApplication(ctx context.Context, applicationID int64) (*types.PublicSpotApplicationItem, error) {
	if applicationID <= 0 {
		return nil, errorsx.BadRequest("invalid public spot application id")
	}
	item, err := s.repo.FindPublicSpotApplicationByID(ctx, applicationID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorsx.NotFound("public spot application not found")
		}
		return nil, errorsx.Internal("failed to query public spot application")
	}
	result := publicSpotApplicationItem(*item)
	return &result, nil
}

func (s *BusinessService) ReviewPublicSpotApplication(ctx context.Context, applicationID int64, status string, req *types.ReviewPublicSpotApplicationRequest) error {
	if applicationID <= 0 || (status != "approved" && status != "rejected") {
		return errorsx.BadRequest("invalid public spot application review")
	}
	if req == nil || strings.TrimSpace(req.ReviewNote) == "" || len([]rune(strings.TrimSpace(req.ReviewNote))) > 500 {
		return errorsx.BadRequest("review note is required and must not exceed 500 characters")
	}
	if status == "rejected" && req.MergeSpotID != nil {
		return errorsx.BadRequest("merge spot is only valid when approving")
	}
	reviewerUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || reviewerUserID <= 0 {
		return errorsx.Unauthorized("invalid login user")
	}
	before, err := s.repo.FindPublicSpotApplicationByID(ctx, applicationID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return errorsx.NotFound("public spot application not found")
		}
		return errorsx.Internal("failed to query public spot application")
	}
	mergeSpotID := int64(0)
	if req.MergeSpotID != nil {
		mergeSpotID = *req.MergeSpotID
		if mergeSpotID < 0 {
			return errorsx.BadRequest("invalid merge spot id")
		}
	}
	note := strings.TrimSpace(req.ReviewNote)
	rows, err := s.repo.ReviewPublicSpotApplication(ctx, applicationID, reviewerUserID, status, note, mergeSpotID, time.Now())
	if err != nil {
		if errors.Is(err, repository.ErrPossibleDuplicatePublicSpot) {
			return errorsx.BadRequest("possible duplicate public spot; select an existing spot to merge")
		}
		return errorsx.Internal("failed to review public spot application")
	}
	if rows == 0 {
		return errorsx.NotFound("pending public spot application not found")
	}
	if err := s.writeAudit(ctx, audit.Event{
		Action: "public_spot_application." + status, ResourceType: "public_spot_application", ResourceID: applicationID,
		Detail: map[string]any{"before": map[string]any{"status": before.Status}, "after": map[string]any{"status": status, "reviewNote": note, "mergeSpotId": mergeSpotID}},
	}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}
	return nil
}

func (s *BusinessService) ListMediaAssets(ctx context.Context, req *types.ListMediaAssetsRequest) (*types.ListMediaAssetsResponse, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "pending"
	}
	if status != "pending" && status != "approved" && status != "rejected" {
		return nil, errorsx.BadRequest("invalid media moderation status")
	}
	items, total, err := s.repo.ListMediaAssets(ctx, repository.MediaAssetListFilter{
		Keyword: strings.TrimSpace(req.Keyword), Status: status,
		Offset: (page - 1) * pageSize, Limit: pageSize,
	})
	if err != nil {
		return nil, errorsx.Internal("failed to query media assets")
	}
	list := make([]types.MediaAssetItem, 0, len(items))
	for _, item := range items {
		list = append(list, mediaAssetItem(item))
	}
	return &types.ListMediaAssetsResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *BusinessService) ReviewMediaAsset(ctx context.Context, mediaID int64, status string, req *types.ReviewMediaAssetRequest) error {
	if mediaID <= 0 || (status != "approved" && status != "rejected") {
		return errorsx.BadRequest("invalid media review")
	}
	if req == nil || strings.TrimSpace(req.ReviewNote) == "" || len([]rune(strings.TrimSpace(req.ReviewNote))) > 500 {
		return errorsx.BadRequest("review note is required and must not exceed 500 characters")
	}
	reviewerUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || reviewerUserID <= 0 {
		return errorsx.Unauthorized("invalid login user")
	}
	before, err := s.repo.FindMediaAssetByID(ctx, mediaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorsx.NotFound("pending media asset not found")
		}
		return errorsx.Internal("failed to query media asset")
	}
	if before.ModerationStatus != "pending" {
		return errorsx.NotFound("pending media asset not found")
	}
	note := strings.TrimSpace(req.ReviewNote)
	reviewedAt := time.Now()
	rows, err := s.repo.ReviewMediaAsset(ctx, mediaID, reviewerUserID, status, note, reviewedAt)
	if err != nil {
		return errorsx.Internal("failed to review media asset")
	}
	if rows == 0 {
		return errorsx.NotFound("pending media asset not found")
	}
	after := map[string]any{"status": status, "reviewNote": note}
	if err := s.writeAudit(ctx, audit.Event{
		Action: "media.asset.review", ResourceType: "fishing_record_image", ResourceID: mediaID,
		Detail: map[string]any{"before": map[string]any{"status": before.ModerationStatus}, "after": after},
	}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}
	return nil
}

func mediaAssetItem(item model.MediaAsset) types.MediaAssetItem {
	return types.MediaAssetItem{
		ID: item.ID, RecordCode: item.RecordCode, UserID: item.UserID, SpotName: item.SpotName,
		ImageURL: item.ImageURL, Width: item.Width, Height: item.Height,
		ModerationStatus: item.ModerationStatus, ModerationNote: item.ModerationNote,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
	}
}

func (s *BusinessService) ListReviewTasks(ctx context.Context, req *types.ListReviewTasksRequest) (*types.ListReviewTasksResponse, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	status := strings.TrimSpace(req.Status)
	if status != "" && status != "pending" && status != "approved" && status != "rejected" {
		return nil, errorsx.BadRequest("invalid review task status")
	}
	taskType := strings.TrimSpace(req.TaskType)
	if taskType != "" && taskType != "spot" && taskType != "user_spot" && taskType != "spot_correction" && taskType != "public_report" {
		return nil, errorsx.BadRequest("invalid review task type")
	}
	tasks, total, err := s.repo.ListReviewTasks(ctx, repository.ReviewTaskListFilter{
		Keyword: strings.TrimSpace(req.Keyword), Status: status, TaskType: taskType,
		Offset: (page - 1) * pageSize, Limit: pageSize,
	})
	if err != nil {
		return nil, errorsx.Internal("failed to query review tasks")
	}
	list := make([]types.ReviewTaskItem, 0, len(tasks))
	for _, task := range tasks {
		list = append(list, types.ReviewTaskItem{
			TaskID: task.TaskID, TaskType: task.TaskType, ResourceID: task.ResourceID,
			ResourceCode: task.ResourceCode, Title: task.Title, Summary: task.Summary,
			Status: task.Status, CreatedAt: task.CreatedAt.Format(time.RFC3339),
		})
	}
	return &types.ListReviewTasksResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *BusinessService) ReviewPublicFishingReport(ctx context.Context, recordID int64, status string, req *types.ReviewPublicFishingReportRequest) error {
	if recordID <= 0 || (status != "approved" && status != "rejected") {
		return errorsx.BadRequest("invalid public fishing report review")
	}
	if req == nil {
		return errorsx.BadRequest("review note is required and must not exceed 500 characters")
	}
	note := strings.TrimSpace(req.ReviewNote)
	if note == "" || len([]rune(note)) > 500 {
		return errorsx.BadRequest("review note is required and must not exceed 500 characters")
	}
	var before *repository.PublicReportReviewState
	if s.audit != nil {
		var err error
		before, err = s.audit.FindPublicReportReviewState(ctx, recordID)
		if err != nil {
			return errorsx.Internal("failed to query public fishing report review state")
		}
	}
	rows, err := s.repo.ReviewPublicFishingReport(ctx, recordID, status)
	if err != nil {
		return errorsx.Internal("failed to review public fishing report")
	}
	if rows == 0 {
		return errorsx.NotFound("public fishing report not found")
	}
	after := repository.PublicReportReviewState{ModerationStatus: status}
	if before != nil {
		after.VisibleScope = before.VisibleScope
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "public_fishing_report.review", ResourceType: "fishing_record", ResourceID: recordID, Detail: map[string]any{"before": before, "after": after, "reviewNote": note}}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}
	return nil
}

func (s *BusinessService) ReviewSpotCorrection(ctx context.Context, correctionID int64, status string, req *types.ReviewSpotCorrectionRequest) error {
	if correctionID <= 0 || (status != "approved" && status != "rejected") {
		return errorsx.BadRequest("invalid correction review")
	}
	if req == nil {
		return errorsx.BadRequest("review note is required and must not exceed 500 characters")
	}
	note := strings.TrimSpace(req.ReviewNote)
	if note == "" || len([]rune(note)) > 500 {
		return errorsx.BadRequest("review note is required and must not exceed 500 characters")
	}
	reviewerUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || reviewerUserID <= 0 {
		return errorsx.Unauthorized("invalid login user")
	}
	correction, err := s.repo.FindSpotCorrectionByID(ctx, correctionID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return errorsx.NotFound("spot correction not found")
		}
		return errorsx.Internal("failed to query spot correction")
	}
	var changes map[string]string
	if status == "approved" {
		changes, err = selectedSpotCorrectionChanges(correction, req.Fields)
		if err != nil {
			return errorsx.BadRequest(err.Error())
		}
	}
	var before *repository.CorrectionReviewState
	if s.audit != nil {
		var err error
		before, err = s.audit.FindCorrectionReviewState(ctx, correctionID)
		if err != nil {
			return errorsx.Internal("failed to query spot correction review state")
		}
	}
	rows, err := s.repo.ReviewSpotCorrectionWithChanges(ctx, correctionID, reviewerUserID, status, note, changes, time.Now())
	if err != nil {
		return errorsx.Internal("failed to review spot correction")
	}
	if rows == 0 {
		return errorsx.NotFound("pending spot correction not found")
	}
	after := repository.CorrectionReviewState{Status: status, ReviewNote: note}
	detail := map[string]any{"before": before, "after": after}
	if fieldChanges := spotCorrectionAuditChanges(correction, changes); len(fieldChanges) > 0 {
		detail["fieldChanges"] = fieldChanges
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "spot.correction.review", ResourceType: "spot_correction", ResourceID: correctionID, Detail: detail}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}
	return nil
}

func spotCorrectionItem(item model.SpotCorrection) types.SpotCorrectionItem {
	result := types.SpotCorrectionItem{
		ID: item.ID, CorrectionCode: item.CorrectionCode, UserID: item.UserID, UserNickname: item.UserNickname,
		SpotCode: item.SpotCode, SpotName: item.SpotName, Content: item.Content, Status: item.Status,
		ReviewNote: item.ReviewNote, CreatedAt: item.CreatedAt.Format(time.RFC3339), Diff: buildSpotCorrectionDiff(&item),
	}
	if item.ReviewerUserID.Valid {
		value := item.ReviewerUserID.Int64
		result.ReviewerUserID = &value
	}
	if item.ReviewedAt.Valid {
		value := item.ReviewedAt.Time.Format(time.RFC3339)
		result.ReviewedAt = &value
	}
	return result
}

func (s *BusinessService) ListFishingSpotReviews(ctx context.Context, req *types.ListFishingSpotReviewsRequest) (*types.ListFishingSpotReviewsResponse, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	filter := repository.FishingSpotListFilter{
		Keyword:    strings.TrimSpace(req.Keyword),
		SourceType: strings.TrimSpace(req.SourceType),
		Offset:     (page - 1) * pageSize,
		Limit:      pageSize,
	}

	var err error
	if filter.VisibleStatus, err = parseOptionalInt64(req.VisibleStatus, "invalid visibleStatus"); err != nil {
		return nil, err
	}
	if filter.PublishStatus, err = parseOptionalInt64(req.PublishStatus, "invalid publishStatus"); err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListFishingSpots(ctx, filter)
	if err != nil {
		return nil, errorsx.Internal("failed to query fishing spots")
	}

	speciesBySpot, tipsBySpot, err := s.loadFishingSpotRelations(ctx, items)
	if err != nil {
		return nil, err
	}

	list := make([]types.FishingSpotReviewItem, 0, len(items))
	for _, item := range items {
		list = append(list, assembler.FishingSpotReview(item, speciesBySpot[item.ID], tipsBySpot[item.ID]))
	}

	return &types.ListFishingSpotReviewsResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *BusinessService) FishingSpotReviewDetail(ctx context.Context, spotID int64) (*types.FishingSpotReviewItem, error) {
	item, err := s.repo.FindFishingSpotByID(ctx, spotID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("fishing spot not found")
		}
		return nil, errorsx.Internal("failed to query fishing spot")
	}

	speciesBySpot, tipsBySpot, err := s.loadFishingSpotRelations(ctx, []model.FishingSpot{*item})
	if err != nil {
		return nil, err
	}

	resp := assembler.FishingSpotReview(*item, speciesBySpot[item.ID], tipsBySpot[item.ID])
	return &resp, nil
}

func (s *BusinessService) ApproveFishingSpot(ctx context.Context, spotID int64, req *types.ApproveFishingSpotRequest) error {
	var before *repository.FishingSpotReviewState
	if s.audit != nil {
		var err error
		before, err = s.audit.FindFishingSpotReviewState(ctx, spotID)
		if err != nil {
			return errorsx.Internal("failed to query fishing spot review state")
		}
		if before == nil {
			return errorsx.NotFound("fishing spot not found")
		}
	} else if _, err := s.FishingSpotReviewDetail(ctx, spotID); err != nil {
		return err
	}

	publishStatus := int64(1)
	if req != nil && req.PublishStatus != nil {
		publishStatus = *req.PublishStatus
	}

	if err := s.repo.ApproveFishingSpot(ctx, spotID, publishStatus, time.Now()); err != nil {
		return errorsx.Internal("failed to approve fishing spot")
	}
	after := repository.FishingSpotReviewState{VisibleStatus: 1, PublishStatus: publishStatus}
	if err := s.writeAudit(ctx, audit.Event{Action: "spot.review.approve", ResourceType: "fishing_spot", ResourceID: spotID, Detail: map[string]any{"before": before, "after": after}}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}

	return nil
}

func (s *BusinessService) RejectFishingSpot(ctx context.Context, spotID int64, req *types.RejectFishingSpotRequest) error {
	var before *repository.FishingSpotReviewState
	if s.audit != nil {
		var err error
		before, err = s.audit.FindFishingSpotReviewState(ctx, spotID)
		if err != nil {
			return errorsx.Internal("failed to query fishing spot review state")
		}
		if before == nil {
			return errorsx.NotFound("fishing spot not found")
		}
	} else if _, err := s.FishingSpotReviewDetail(ctx, spotID); err != nil {
		return err
	}

	publishStatus := int64(0)
	if req != nil && req.PublishStatus != nil {
		publishStatus = *req.PublishStatus
	}

	if err := s.repo.RejectFishingSpot(ctx, spotID, publishStatus); err != nil {
		return errorsx.Internal("failed to reject fishing spot")
	}
	after := repository.FishingSpotReviewState{VisibleStatus: 0, PublishStatus: publishStatus}
	if err := s.writeAudit(ctx, audit.Event{Action: "spot.review.reject", ResourceType: "fishing_spot", ResourceID: spotID, Detail: map[string]any{"before": before, "after": after}}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}

	return nil
}

func (s *BusinessService) ListArticles(ctx context.Context, req *types.ListArticlesRequest) (*types.ListArticlesResponse, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	filter := repository.ArticleListFilter{
		Keyword: strings.TrimSpace(req.Keyword),
		Offset:  (page - 1) * pageSize,
		Limit:   pageSize,
	}

	var err error
	if filter.PublishStatus, err = parseOptionalInt64(req.PublishStatus, "invalid publishStatus"); err != nil {
		return nil, err
	}
	if filter.VisibleStatus, err = parseOptionalInt64(req.VisibleStatus, "invalid visibleStatus"); err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListArticles(ctx, filter)
	if err != nil {
		return nil, errorsx.Internal("failed to query articles")
	}

	detailMap, err := s.loadArticleDetails(ctx, items)
	if err != nil {
		return nil, err
	}

	list := make([]types.ArticleItem, 0, len(items))
	for _, item := range items {
		list = append(list, assembler.Article(item, detailMap[item.ArticleCode]))
	}

	return &types.ListArticlesResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *BusinessService) ArticleDetail(ctx context.Context, articleID int64) (*types.ArticleItem, error) {
	item, err := s.repo.FindArticleByID(ctx, articleID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("article not found")
		}
		return nil, errorsx.Internal("failed to query article")
	}

	detail, err := s.repo.FindArticleDetailByCode(ctx, item.ArticleCode)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		return nil, errorsx.Internal("failed to query article detail")
	}

	resp := assembler.Article(*item, detail)
	return &resp, nil
}

func (s *BusinessService) CreateArticle(ctx context.Context, req *types.ArticleUpsertRequest) (*types.ArticleItem, error) {
	item, detail, err := buildArticleModels(req, "")
	if err != nil {
		return nil, err
	}

	item.ArticleCode = generateBizCode("article")
	detail.ContentCode = item.ArticleCode
	item.PublishStatus = 0
	item.PublishedAt = toNullTime(time.Time{})

	articleID, err := s.repo.CreateArticle(ctx, item, detail)
	if err != nil {
		return nil, errorsx.Internal("failed to create article")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "article.create", ResourceType: "article", ResourceID: articleID, Detail: map[string]any{"title": item.Title}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}

	return s.ArticleDetail(ctx, articleID)
}

func (s *BusinessService) UpdateArticle(ctx context.Context, articleID int64, req *types.ArticleUpsertRequest) (*types.ArticleItem, error) {
	current, err := s.repo.FindArticleByID(ctx, articleID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("article not found")
		}
		return nil, errorsx.Internal("failed to query article")
	}

	item, detail, err := buildArticleModels(req, current.ArticleCode)
	if err != nil {
		return nil, err
	}
	item.PublishStatus = current.PublishStatus
	item.PublishedAt = current.PublishedAt

	if err := s.repo.UpdateArticle(ctx, articleID, item, detail); err != nil {
		return nil, errorsx.Internal("failed to update article")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "article.update", ResourceType: "article", ResourceID: articleID, Detail: map[string]any{"title": item.Title}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}

	return s.ArticleDetail(ctx, articleID)
}

func (s *BusinessService) PublishArticle(ctx context.Context, articleID int64, req *types.ArticlePublishRequest) error {
	if _, err := s.ArticleDetail(ctx, articleID); err != nil {
		return err
	}

	visibleStatus := int64(1)
	if req != nil && req.VisibleStatus != nil {
		visibleStatus = *req.VisibleStatus
	}

	if err := s.repo.PublishArticle(ctx, articleID, visibleStatus, time.Now()); err != nil {
		return errorsx.Internal("failed to publish article")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "article.publish", ResourceType: "article", ResourceID: articleID, Detail: map[string]any{"visibleStatus": visibleStatus}}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}

	return nil
}

func (s *BusinessService) OfflineArticle(ctx context.Context, articleID int64) error {
	if _, err := s.ArticleDetail(ctx, articleID); err != nil {
		return err
	}

	if err := s.repo.OfflineArticle(ctx, articleID); err != nil {
		return errorsx.Internal("failed to offline article")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "article.offline", ResourceType: "article", ResourceID: articleID}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}

	return nil
}

func (s *BusinessService) ListBanners(ctx context.Context, req *types.ListBannersRequest) (*types.ListBannersResponse, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	visibleStatus, err := parseOptionalInt64(req.VisibleStatus, "invalid visibleStatus")
	if err != nil {
		return nil, err
	}
	if visibleStatus != nil && (*visibleStatus < 0 || *visibleStatus > 1) {
		return nil, errorsx.BadRequest("invalid visibleStatus")
	}
	items, total, err := s.repo.ListBanners(ctx, repository.BannerListFilter{
		Keyword: strings.TrimSpace(req.Keyword), VisibleStatus: visibleStatus,
		Offset: (page - 1) * pageSize, Limit: pageSize,
	})
	if err != nil {
		return nil, errorsx.Internal("failed to query banners")
	}
	list := make([]types.BannerItem, 0, len(items))
	for _, item := range items {
		list = append(list, bannerItem(item))
	}
	return &types.ListBannersResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *BusinessService) BannerDetail(ctx context.Context, bannerID int64) (*types.BannerItem, error) {
	if bannerID <= 0 {
		return nil, errorsx.BadRequest("invalid banner id")
	}
	item, err := s.repo.FindBannerByID(ctx, bannerID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("banner not found")
		}
		return nil, errorsx.Internal("failed to query banner")
	}
	result := bannerItem(*item)
	return &result, nil
}

func (s *BusinessService) CreateBanner(ctx context.Context, req *types.BannerUpsertRequest) (*types.BannerItem, error) {
	item, err := buildBannerModel(req)
	if err != nil {
		return nil, err
	}
	item.BannerCode = generateBizCode("banner")
	if item.VisibleStatus == 1 {
		now := time.Now()
		item.PublishedAt = &now
	}
	if err := s.repo.CreateBanner(ctx, &item); err != nil {
		return nil, errorsx.Internal("failed to create banner")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "banner.create", ResourceType: "discover_event_banner", ResourceID: item.ID, Detail: map[string]any{"title": item.Title}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}
	result := bannerItem(item)
	return &result, nil
}

func (s *BusinessService) UpdateBanner(ctx context.Context, bannerID int64, req *types.BannerUpsertRequest) (*types.BannerItem, error) {
	current, err := s.repo.FindBannerByID(ctx, bannerID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("banner not found")
		}
		return nil, errorsx.Internal("failed to query banner")
	}
	item, err := buildBannerModel(req)
	if err != nil {
		return nil, err
	}
	item.ID = current.ID
	item.BannerCode = current.BannerCode
	if item.VisibleStatus == 1 {
		item.PublishedAt = current.PublishedAt
		if item.PublishedAt == nil {
			now := time.Now()
			item.PublishedAt = &now
		}
	}
	if err := s.repo.UpdateBanner(ctx, bannerID, item); err != nil {
		return nil, errorsx.Internal("failed to update banner")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "banner.update", ResourceType: "discover_event_banner", ResourceID: bannerID, Detail: map[string]any{"title": item.Title}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}
	return s.BannerDetail(ctx, bannerID)
}

func (s *BusinessService) UpdateBannerVisibility(ctx context.Context, bannerID int64, req *types.UpdateBannerVisibilityRequest) error {
	if bannerID <= 0 || req == nil || (req.VisibleStatus != 0 && req.VisibleStatus != 1) {
		return errorsx.BadRequest("invalid banner visibility")
	}
	if _, err := s.BannerDetail(ctx, bannerID); err != nil {
		return err
	}
	var publishedAt *time.Time
	if req.VisibleStatus == 1 {
		now := time.Now()
		publishedAt = &now
	}
	if err := s.repo.UpdateBannerVisibility(ctx, bannerID, req.VisibleStatus, publishedAt); err != nil {
		return errorsx.Internal("failed to update banner visibility")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "banner.visibility.update", ResourceType: "discover_event_banner", ResourceID: bannerID, Detail: map[string]any{"visibleStatus": req.VisibleStatus}}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}
	return nil
}

func (s *BusinessService) ListSpecies(ctx context.Context, req *types.ListSpeciesRequest) (*types.ListSpeciesResponse, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	filter := repository.SpeciesListFilter{
		Keyword:  strings.TrimSpace(req.Keyword),
		Category: strings.TrimSpace(req.Category),
		Offset:   (page - 1) * pageSize,
		Limit:    pageSize,
	}

	var err error
	if filter.VisibleStatus, err = parseOptionalInt64(req.VisibleStatus, "invalid visibleStatus"); err != nil {
		return nil, err
	}
	if filter.IsFeatured, err = parseOptionalInt64(req.IsFeatured, "invalid isFeatured"); err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListSpecies(ctx, filter)
	if err != nil {
		return nil, errorsx.Internal("failed to query species")
	}

	tipsBySpecies, err := s.loadSpeciesDetailTips(ctx, items)
	if err != nil {
		return nil, err
	}

	list := make([]types.SpeciesItem, 0, len(items))
	for _, item := range items {
		list = append(list, assembler.Species(item, tipsBySpecies[item.ID]))
	}

	return &types.ListSpeciesResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *BusinessService) SpeciesDetail(ctx context.Context, speciesID int64) (*types.SpeciesItem, error) {
	item, err := s.repo.FindSpeciesByID(ctx, speciesID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("species not found")
		}
		return nil, errorsx.Internal("failed to query species")
	}

	tipsBySpecies, err := s.loadSpeciesDetailTips(ctx, []model.Species{*item})
	if err != nil {
		return nil, err
	}

	resp := assembler.Species(*item, tipsBySpecies[item.ID])
	return &resp, nil
}

func (s *BusinessService) CreateSpecies(ctx context.Context, req *types.SpeciesUpsertRequest) (*types.SpeciesItem, error) {
	item, tips, err := buildSpeciesModel(req)
	if err != nil {
		return nil, err
	}
	item.SpeciesCode = generateBizCode("species")

	speciesID, err := s.repo.CreateSpecies(ctx, item, tips)
	if err != nil {
		return nil, errorsx.Internal("failed to create species")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "species.create", ResourceType: "species", ResourceID: speciesID, Detail: map[string]any{"name": item.Name}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}

	return s.SpeciesDetail(ctx, speciesID)
}

func (s *BusinessService) UpdateSpecies(ctx context.Context, speciesID int64, req *types.SpeciesUpsertRequest) (*types.SpeciesItem, error) {
	current, err := s.repo.FindSpeciesByID(ctx, speciesID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("species not found")
		}
		return nil, errorsx.Internal("failed to query species")
	}

	item, tips, err := buildSpeciesModel(req)
	if err != nil {
		return nil, err
	}
	item.SpeciesCode = current.SpeciesCode

	if err := s.repo.UpdateSpecies(ctx, speciesID, item, tips); err != nil {
		return nil, errorsx.Internal("failed to update species")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "species.update", ResourceType: "species", ResourceID: speciesID, Detail: map[string]any{"name": item.Name}}); err != nil {
		return nil, errorsx.Internal("failed to write audit log")
	}

	return s.SpeciesDetail(ctx, speciesID)
}

func (s *BusinessService) UpdateSpeciesVisibility(ctx context.Context, speciesID int64, visibleStatus int64) error {
	if visibleStatus != 0 && visibleStatus != 1 {
		return errorsx.BadRequest("invalid visibleStatus")
	}

	if _, err := s.SpeciesDetail(ctx, speciesID); err != nil {
		return err
	}

	if err := s.repo.UpdateSpeciesVisibility(ctx, speciesID, visibleStatus); err != nil {
		return errorsx.Internal("failed to update species visibility")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "species.visibility.update", ResourceType: "species", ResourceID: speciesID, Detail: map[string]any{"visibleStatus": visibleStatus}}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}

	return nil
}

func (s *BusinessService) ListUsers(ctx context.Context, req *types.ListUsersRequest) (*types.ListUsersResponse, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	filter := repository.UserListFilter{
		Keyword: strings.TrimSpace(req.Keyword),
		City:    strings.TrimSpace(req.City),
		Offset:  (page - 1) * pageSize,
		Limit:   pageSize,
	}

	var err error
	if filter.Status, err = parseOptionalInt64(req.Status, "invalid status"); err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListUsers(ctx, filter)
	if err != nil {
		return nil, errorsx.Internal("failed to query users")
	}

	list := make([]types.UserItem, 0, len(items))
	for _, item := range items {
		list = append(list, assembler.User(item))
	}

	return &types.ListUsersResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *BusinessService) UserDetail(ctx context.Context, userID int64) (*types.UserItem, error) {
	item, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorsx.NotFound("user not found")
		}
		return nil, errorsx.Internal("failed to query user")
	}

	resp := assembler.User(*item)
	return &resp, nil
}

func (s *BusinessService) UpdateUserStatus(ctx context.Context, userID int64, status int64) error {
	if status != 0 && status != 1 {
		return errorsx.BadRequest("invalid status")
	}

	if _, err := s.UserDetail(ctx, userID); err != nil {
		return err
	}

	if err := s.repo.UpdateUserStatus(ctx, userID, status); err != nil {
		return errorsx.Internal("failed to update user status")
	}
	if err := s.writeAudit(ctx, audit.Event{Action: "user.status.update", ResourceType: "user", ResourceID: userID, Detail: map[string]any{"status": status}}); err != nil {
		return errorsx.Internal("failed to write audit log")
	}

	return nil
}

func (s *BusinessService) loadFishingSpotRelations(ctx context.Context, items []model.FishingSpot) (map[int64][]string, map[int64][]string, error) {
	spotIDs := make([]int64, 0, len(items))
	for _, item := range items {
		spotIDs = append(spotIDs, item.ID)
	}

	speciesRows, err := s.repo.ListFishingSpotSpeciesBySpotIDs(ctx, spotIDs)
	if err != nil {
		return nil, nil, errorsx.Internal("failed to query fishing spot species")
	}
	tipRows, err := s.repo.ListFishingSpotTipsBySpotIDs(ctx, spotIDs)
	if err != nil {
		return nil, nil, errorsx.Internal("failed to query fishing spot tips")
	}

	speciesBySpot := make(map[int64][]string)
	for _, row := range speciesRows {
		speciesBySpot[row.SpotID] = append(speciesBySpot[row.SpotID], row.SpeciesName)
	}

	tipsBySpot := make(map[int64][]string)
	for _, row := range tipRows {
		tipsBySpot[row.SpotID] = append(tipsBySpot[row.SpotID], row.TipText)
	}

	return speciesBySpot, tipsBySpot, nil
}

func (s *BusinessService) loadArticleDetails(ctx context.Context, items []model.DiscoverArticle) (map[string]*model.DiscoverContentDetail, error) {
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.ArticleCode)
	}

	rows, err := s.repo.FindArticleDetailsByCodes(ctx, codes)
	if err != nil {
		return nil, errorsx.Internal("failed to query article details")
	}

	detailMap := make(map[string]*model.DiscoverContentDetail, len(rows))
	for index := range rows {
		row := rows[index]
		detailMap[row.ContentCode] = &row
	}

	return detailMap, nil
}

func (s *BusinessService) loadSpeciesDetailTips(ctx context.Context, items []model.Species) (map[int64][]string, error) {
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}

	rows, err := s.repo.ListSpeciesDetailTipsBySpeciesIDs(ctx, ids)
	if err != nil {
		return nil, errorsx.Internal("failed to query species detail tips")
	}

	tipsBySpecies := make(map[int64][]string)
	for _, row := range rows {
		tipsBySpecies[row.SpeciesID] = append(tipsBySpecies[row.SpeciesID], row.TipText)
	}

	return tipsBySpecies, nil
}

func buildArticleModels(req *types.ArticleUpsertRequest, articleCode string) (model.DiscoverArticle, model.DiscoverContentDetail, error) {
	if req == nil {
		return model.DiscoverArticle{}, model.DiscoverContentDetail{}, errorsx.BadRequest("missing article payload")
	}
	if strings.TrimSpace(req.Title) == "" {
		return model.DiscoverArticle{}, model.DiscoverContentDetail{}, errorsx.BadRequest("title is required")
	}

	visibleStatus := req.VisibleStatus
	if visibleStatus != 0 && visibleStatus != 1 {
		return model.DiscoverArticle{}, model.DiscoverContentDetail{}, errorsx.BadRequest("invalid visibleStatus")
	}

	item := model.DiscoverArticle{
		ArticleCode:   articleCode,
		ShortLabel:    strings.TrimSpace(req.ShortLabel),
		Title:         strings.TrimSpace(req.Title),
		Description:   strings.TrimSpace(req.Description),
		Theme:         strings.TrimSpace(req.Theme),
		SortOrder:     req.SortOrder,
		VisibleStatus: visibleStatus,
	}

	detail := model.DiscoverContentDetail{
		ContentCode:   articleCode,
		TypeLabel:     strings.TrimSpace(req.TypeLabel),
		SummaryText:   joinLines(req.SummaryLines),
		TakeawaysText: joinLines(req.TakeawayLines),
		VisibleStatus: visibleStatus,
	}

	return item, detail, nil
}

func buildSpeciesModel(req *types.SpeciesUpsertRequest) (model.Species, []string, error) {
	if req == nil {
		return model.Species{}, nil, errorsx.BadRequest("missing species payload")
	}
	if strings.TrimSpace(req.Name) == "" {
		return model.Species{}, nil, errorsx.BadRequest("name is required")
	}
	if req.VisibleStatus != 0 && req.VisibleStatus != 1 {
		return model.Species{}, nil, errorsx.BadRequest("invalid visibleStatus")
	}
	if req.IsFeatured != 0 && req.IsFeatured != 1 {
		return model.Species{}, nil, errorsx.BadRequest("invalid isFeatured")
	}

	item := model.Species{
		Name:          strings.TrimSpace(req.Name),
		Alias:         strings.TrimSpace(req.Alias),
		Category:      strings.TrimSpace(req.Category),
		WaterLayer:    strings.TrimSpace(req.WaterLayer),
		TagText:       strings.TrimSpace(req.TagText),
		SeasonText:    strings.TrimSpace(req.SeasonText),
		BestWindow:    strings.TrimSpace(req.BestWindow),
		FishingMethod: strings.TrimSpace(req.FishingMethod),
		BaitText:      strings.TrimSpace(req.BaitText),
		Description:   strings.TrimSpace(req.Description),
		HighlightText: strings.TrimSpace(req.HighlightText),
		IsFeatured:    req.IsFeatured,
		SortOrder:     req.SortOrder,
		VisibleStatus: req.VisibleStatus,
	}

	return item, cleanLines(req.DetailTips), nil
}

func normalizePage(page int64, pageSize int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}

func parseOptionalInt64(raw string, invalidMessage string) (*int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, errorsx.BadRequest(invalidMessage)
	}

	return &parsed, nil
}

func joinLines(items []string) string {
	return strings.Join(cleanLines(items), "\n")
}

func cleanLines(items []string) []string {
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}

	return cleaned
}

func publicSpotApplicationItem(item model.PublicSpotApplication) types.PublicSpotApplicationItem {
	result := types.PublicSpotApplicationItem{
		ID: item.ID, ApplicationCode: item.ApplicationCode, UserID: item.UserID, UserNickname: item.UserNickname,
		SourceSpotCode: item.SourceSpotCode, Name: item.Name, Address: item.Address, Latitude: item.Latitude,
		Longitude: item.Longitude, Description: item.Description, ChargeType: item.ChargeType,
		TargetSpecies: append([]string{}, item.TargetSpecies...), ImageURLs: append([]string{}, item.ImageURLs...),
		Status: item.Status, ReviewNote: item.ReviewNote, ReviewerUserID: item.ReviewerUserID,
		MergedSpotID: item.MergedSpotID, MergedSpotCode: item.MergedSpotCode, CreatedAt: item.CreatedAt.Format(time.RFC3339),
	}
	if item.ReviewedAt != nil {
		value := item.ReviewedAt.Format(time.RFC3339)
		result.ReviewedAt = &value
	}
	return result
}

func buildBannerModel(req *types.BannerUpsertRequest) (model.DiscoverEventBanner, error) {
	if req == nil {
		return model.DiscoverEventBanner{}, errorsx.BadRequest("banner payload is required")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" || len([]rune(title)) > 128 {
		return model.DiscoverEventBanner{}, errorsx.BadRequest("banner title is required and must not exceed 128 characters")
	}
	if len([]rune(strings.TrimSpace(req.Tag))) > 32 || len([]rune(strings.TrimSpace(req.Description))) > 255 ||
		len([]rune(strings.TrimSpace(req.DeadlineText))) > 64 || len([]rune(strings.TrimSpace(req.LocationName))) > 128 {
		return model.DiscoverEventBanner{}, errorsx.BadRequest("banner field exceeds maximum length")
	}
	if req.VisibleStatus != 0 && req.VisibleStatus != 1 {
		return model.DiscoverEventBanner{}, errorsx.BadRequest("invalid visibleStatus")
	}
	deadlineAt, err := parseBannerDate(req.DeadlineAt)
	if err != nil {
		return model.DiscoverEventBanner{}, errorsx.BadRequest("invalid deadlineAt")
	}
	return model.DiscoverEventBanner{
		Tag: strings.TrimSpace(req.Tag), Title: title, Description: strings.TrimSpace(req.Description),
		DeadlineText: strings.TrimSpace(req.DeadlineText), DeadlineAt: deadlineAt,
		LocationName: strings.TrimSpace(req.LocationName), SortOrder: req.SortOrder,
		VisibleStatus: req.VisibleStatus,
	}, nil
}

func parseBannerDate(raw string) (*time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func bannerItem(item model.DiscoverEventBanner) types.BannerItem {
	result := types.BannerItem{
		ID: item.ID, BannerCode: item.BannerCode, Tag: item.Tag, Title: item.Title,
		Description: item.Description, DeadlineText: item.DeadlineText, LocationName: item.LocationName,
		SortOrder: item.SortOrder, VisibleStatus: item.VisibleStatus, CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
	}
	if item.DeadlineAt != nil {
		result.DeadlineAt = item.DeadlineAt.Format("2006-01-02")
	}
	if item.PublishedAt != nil {
		result.PublishedAt = item.PublishedAt.Format(time.RFC3339)
	}
	return result
}

func generateBizCode(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func (s *BusinessService) writeAudit(ctx context.Context, event audit.Event) error {
	if s.audit == nil {
		return nil
	}
	actorUserID := event.ActorUserID
	if actorUserID <= 0 {
		actorUserID, _ = auth.UserIDFromContext(ctx)
	}
	detail, err := audit.MarshalDetail(event.Detail)
	if err != nil {
		slog.WarnContext(ctx, "failed to serialize admin audit detail", "action", event.Action, "resource_type", event.ResourceType, "resource_id", event.ResourceID, "error", err)
		return nil
	}
	if err := s.audit.Append(ctx, actorUserID, event.Action, event.ResourceType, event.ResourceID, detail); err != nil {
		slog.WarnContext(ctx, "failed to append admin audit log", "action", event.Action, "resource_type", event.ResourceType, "resource_id", event.ResourceID, "error", err)
	}
	return nil
}

func toNullTime(value time.Time) sql.NullTime {
	if value.IsZero() {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: value, Valid: true}
}
