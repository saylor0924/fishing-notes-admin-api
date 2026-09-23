package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fishing-notes-admin-api/internal/assembler"
	"fishing-notes-admin-api/internal/auth"
	"fishing-notes-admin-api/internal/errorsx"
	"fishing-notes-admin-api/internal/model"
	"fishing-notes-admin-api/internal/repository"
	"fishing-notes-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type BusinessService struct {
	repo *repository.BusinessRepository
}

func NewBusinessService(repo *repository.BusinessRepository) *BusinessService {
	return &BusinessService{repo: repo}
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
		result := types.SpotCorrectionItem{ID: item.ID, CorrectionCode: item.CorrectionCode, SpotCode: item.SpotCode, SpotName: item.SpotName, Content: item.Content, Status: item.Status, ReviewNote: item.ReviewNote, CreatedAt: item.CreatedAt.Format(time.RFC3339)}
		if item.ReviewerUserID.Valid {
			value := item.ReviewerUserID.Int64
			result.ReviewerUserID = &value
		}
		if item.ReviewedAt.Valid {
			value := item.ReviewedAt.Time.Format(time.RFC3339)
			result.ReviewedAt = &value
		}
		list = append(list, result)
	}
	return &types.ListSpotCorrectionsResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *BusinessService) ReviewSpotCorrection(ctx context.Context, correctionID int64, status string, req *types.ReviewSpotCorrectionRequest) error {
	if correctionID <= 0 || (status != "approved" && status != "rejected") {
		return errorsx.BadRequest("invalid correction review")
	}
	note := strings.TrimSpace(req.ReviewNote)
	if note == "" || len([]rune(note)) > 500 {
		return errorsx.BadRequest("review note is required and must not exceed 500 characters")
	}
	reviewerUserID, ok := auth.UserIDFromContext(ctx)
	if !ok || reviewerUserID <= 0 {
		return errorsx.Unauthorized("invalid login user")
	}
	rows, err := s.repo.ReviewSpotCorrection(ctx, correctionID, reviewerUserID, status, note, time.Now())
	if err != nil {
		return errorsx.Internal("failed to review spot correction")
	}
	if rows == 0 {
		return errorsx.NotFound("pending spot correction not found")
	}
	return nil
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
	if _, err := s.FishingSpotReviewDetail(ctx, spotID); err != nil {
		return err
	}

	publishStatus := int64(1)
	if req != nil && req.PublishStatus != nil {
		publishStatus = *req.PublishStatus
	}

	if err := s.repo.ApproveFishingSpot(ctx, spotID, publishStatus, time.Now()); err != nil {
		return errorsx.Internal("failed to approve fishing spot")
	}

	return nil
}

func (s *BusinessService) RejectFishingSpot(ctx context.Context, spotID int64, req *types.RejectFishingSpotRequest) error {
	if _, err := s.FishingSpotReviewDetail(ctx, spotID); err != nil {
		return err
	}

	publishStatus := int64(0)
	if req != nil && req.PublishStatus != nil {
		publishStatus = *req.PublishStatus
	}

	if err := s.repo.RejectFishingSpot(ctx, spotID, publishStatus); err != nil {
		return errorsx.Internal("failed to reject fishing spot")
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

	return nil
}

func (s *BusinessService) OfflineArticle(ctx context.Context, articleID int64) error {
	if _, err := s.ArticleDetail(ctx, articleID); err != nil {
		return err
	}

	if err := s.repo.OfflineArticle(ctx, articleID); err != nil {
		return errorsx.Internal("failed to offline article")
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

func generateBizCode(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func toNullTime(value time.Time) sql.NullTime {
	if value.IsZero() {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: value, Valid: true}
}
