package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Append(ctx context.Context, actorUserID int64, action, resourceType string, resourceID int64, detailJSON string) error {
	var actorID sql.NullInt64
	if actorUserID > 0 {
		actorID = sql.NullInt64{Int64: actorUserID, Valid: true}
	}
	var targetID sql.NullInt64
	if resourceID > 0 {
		targetID = sql.NullInt64{Int64: resourceID, Valid: true}
	}
	if detailJSON == "" {
		detailJSON = "{}"
	}
	return r.db.WithContext(ctx).Create(&adminAuditLogModel{
		ActorUserID: actorID, Action: action, ResourceType: resourceType, ResourceID: targetID, DetailJSON: detailJSON,
	}).Error
}

type AuditLogFilter struct {
	ResourceType string
	ResourceID   int64
	Action       string
	Offset       int
	Limit        int
}

type AuditLog struct {
	ID           int64
	ActorUserID  *int64
	Action       string
	ResourceType string
	ResourceID   *int64
	DetailJSON   string
	CreatedAt    time.Time
}

func (r *AuditRepository) List(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&adminAuditLogModel{})
	if filter.ResourceType != "" {
		query = query.Where("resource_type = ?", filter.ResourceType)
	}
	if filter.ResourceID > 0 {
		query = query.Where("resource_id = ?", filter.ResourceID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []AuditLog{}, 0, nil
	}
	var rows []adminAuditLogModel
	if err := query.Order("created_at DESC").Order("id DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	result := make([]AuditLog, len(rows))
	for index, row := range rows {
		result[index] = AuditLog{
			ID: row.ID, ActorUserID: nullableInt64(row.ActorUserID), Action: row.Action, ResourceType: row.ResourceType,
			ResourceID: nullableInt64(row.ResourceID), DetailJSON: row.DetailJSON, CreatedAt: row.CreatedAt,
		}
	}
	return result, total, nil
}

type CorrectionReviewState struct {
	Status     string `json:"status"`
	ReviewNote string `json:"reviewNote"`
}

func (r *AuditRepository) FindCorrectionReviewState(ctx context.Context, correctionID int64) (*CorrectionReviewState, error) {
	var state CorrectionReviewState
	err := r.db.WithContext(ctx).Table("spot_corrections").Select("status, review_note").Where("id = ?", correctionID).Take(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

type FishingSpotReviewState struct {
	VisibleStatus int64 `json:"visibleStatus"`
	PublishStatus int64 `json:"publishStatus"`
}

func (r *AuditRepository) FindFishingSpotReviewState(ctx context.Context, spotID int64) (*FishingSpotReviewState, error) {
	var state FishingSpotReviewState
	err := r.db.WithContext(ctx).Table("fishing_spots").Select("visible_status, publish_status").Where("id = ?", spotID).Take(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

type PublicReportReviewState struct {
	ModerationStatus string `json:"moderationStatus"`
	VisibleScope     string `json:"visibleScope"`
}

func (r *AuditRepository) FindPublicReportReviewState(ctx context.Context, recordID int64) (*PublicReportReviewState, error) {
	var state PublicReportReviewState
	err := r.db.WithContext(ctx).Table("fishing_records").Select("moderation_status, visible_scope").Where("id = ?", recordID).Take(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

type adminAuditLogModel struct {
	ID           int64         `gorm:"column:id;primaryKey"`
	ActorUserID  sql.NullInt64 `gorm:"column:actor_user_id"`
	Action       string        `gorm:"column:action"`
	ResourceType string        `gorm:"column:resource_type"`
	ResourceID   sql.NullInt64 `gorm:"column:resource_id"`
	DetailJSON   string        `gorm:"column:detail_json"`
	CreatedAt    time.Time     `gorm:"column:created_at"`
}

func (adminAuditLogModel) TableName() string { return "admin_audit_logs" }

func nullableInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}
