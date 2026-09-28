package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"fishing-notes-admin-api/internal/model"

	"github.com/sh3lk/h3-go/v4"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FishingSpotListFilter struct {
	Keyword       string
	SourceType    string
	VisibleStatus *int64
	PublishStatus *int64
	Offset        int64
	Limit         int64
}

type ArticleListFilter struct {
	Keyword       string
	VisibleStatus *int64
	PublishStatus *int64
	Offset        int64
	Limit         int64
}

type BannerListFilter struct {
	Keyword       string
	VisibleStatus *int64
	Offset        int64
	Limit         int64
}

type SpeciesListFilter struct {
	Keyword       string
	Category      string
	VisibleStatus *int64
	IsFeatured    *int64
	Offset        int64
	Limit         int64
}

type UserListFilter struct {
	Keyword string
	City    string
	Status  *int64
	Offset  int64
	Limit   int64
}

type SpotCorrectionListFilter struct {
	Keyword string
	Status  string
	Offset  int64
	Limit   int64
}

type PublicSpotApplicationListFilter struct {
	Keyword string
	Status  string
	Offset  int64
	Limit   int64
}

var ErrPossibleDuplicatePublicSpot = errors.New("possible duplicate public spot")

type MediaAssetListFilter struct {
	Keyword string
	Status  string
	Offset  int64
	Limit   int64
}

type ReviewTaskListFilter struct {
	Keyword  string
	Status   string
	TaskType string
	Offset   int64
	Limit    int64
}

type BusinessRepository struct {
	db *gorm.DB
}

func NewBusinessRepository(db *gorm.DB) *BusinessRepository {
	return &BusinessRepository{db: db}
}

// DashboardOverview 查询运营后台首页所需的业务数量指标。
func (r *BusinessRepository) DashboardOverview(ctx context.Context) (model.DashboardOverview, error) {
	if r.db == nil {
		return model.DashboardOverview{}, errors.New("GORM database is not configured")
	}
	var result model.DashboardOverview
	counts := []struct {
		target *int64
		query  *gorm.DB
	}{
		{&result.Users, r.db.WithContext(ctx).Model(&businessUserRepositoryModel{})},
		{&result.FishingSpots, r.db.WithContext(ctx).Model(&spotCorrectionSpotRepositoryModel{})},
		{&result.PendingFishingSpots, r.db.WithContext(ctx).Model(&spotCorrectionSpotRepositoryModel{}).Where("publish_status NOT IN ?", []int64{1, 2})},
		{&result.Articles, r.db.WithContext(ctx).Model(&model.DiscoverArticle{})},
		{&result.PublishedArticles, r.db.WithContext(ctx).Model(&model.DiscoverArticle{}).Where("publish_status = ?", 1)},
		{&result.Species, r.db.WithContext(ctx).Model(&model.Species{})},
		{&result.VisibleSpecies, r.db.WithContext(ctx).Model(&model.Species{}).Where("visible_status = ?", 1)},
		{&result.FishingRecords, r.db.WithContext(ctx).Model(&reviewPublicReportTaskRow{})},
		{&result.PublicFishingRecords, r.db.WithContext(ctx).Model(&reviewPublicReportTaskRow{}).Where("visible_scope = ?", "public")},
		{&result.PendingCorrections, r.db.WithContext(ctx).Model(&reviewCorrectionTaskRow{}).Where("status = ?", "pending")},
		{&result.PendingMediaAssets, r.db.WithContext(ctx).Model(&mediaAssetRepositoryModel{}).Joins("Record").Where("Record.visible_scope = ?", "public").Where("fishing_record_images.moderation_status = ?", "pending")},
		{&result.PendingPublicReports, r.db.WithContext(ctx).Model(&reviewPublicReportTaskRow{}).Where("visible_scope = ? AND moderation_status = ?", "public", "pending")},
	}
	for _, item := range counts {
		if err := item.query.Count(item.target).Error; err != nil {
			return model.DashboardOverview{}, err
		}
	}
	result.PendingReviewTasks = result.PendingCorrections + result.PendingFishingSpots + result.PendingPublicReports
	return result, nil
}

func (r *BusinessRepository) ListSpotCorrections(ctx context.Context, filter SpotCorrectionListFilter) ([]model.SpotCorrection, int64, error) {
	if r.db == nil {
		return nil, 0, errors.New("GORM database is not configured")
	}
	query := r.db.WithContext(ctx).Model(&spotCorrectionRepositoryModel{})
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("(spot_name LIKE ? OR spot_code LIKE ? OR content LIKE ?)", keyword, keyword, keyword)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.SpotCorrection{}, 0, nil
	}
	var rows []spotCorrectionRepositoryModel
	if err := query.Preload("User").Order("created_at DESC").Order("id DESC").Limit(int(filter.Limit)).Offset(int(filter.Offset)).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]model.SpotCorrection, len(rows))
	for index, row := range rows {
		items[index] = spotCorrectionFromRepositoryModel(row)
	}
	return items, total, nil
}

func (r *BusinessRepository) FindSpotCorrectionByID(ctx context.Context, correctionID int64) (*model.SpotCorrection, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var row spotCorrectionRepositoryModel
	err := r.db.WithContext(ctx).Preload("User").Preload("Spot").Where("id = ?", correctionID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	item := spotCorrectionFromRepositoryModel(row)
	return &item, nil
}

// ListPublicSpotApplications 查询公共钓点申请审核列表。
func (r *BusinessRepository) ListPublicSpotApplications(ctx context.Context, filter PublicSpotApplicationListFilter) ([]model.PublicSpotApplication, int64, error) {
	if r.db == nil {
		return nil, 0, errors.New("GORM database is not configured")
	}
	query := r.db.WithContext(ctx).Model(&publicSpotApplicationRepositoryModel{}).Preload("User")
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("application_code LIKE ? OR name LIKE ? OR address LIKE ?", keyword, keyword, keyword)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.PublicSpotApplication{}, 0, nil
	}
	var rows []publicSpotApplicationRepositoryModel
	if err := query.Order("created_at DESC").Order("id DESC").Limit(int(filter.Limit)).Offset(int(filter.Offset)).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]model.PublicSpotApplication, len(rows))
	for index, row := range rows {
		item, err := publicSpotApplicationFromRepositoryModel(row)
		if err != nil {
			return nil, 0, err
		}
		items[index] = item
	}
	return items, total, nil
}

// FindPublicSpotApplicationByID 查询公共钓点申请详情。
func (r *BusinessRepository) FindPublicSpotApplicationByID(ctx context.Context, applicationID int64) (*model.PublicSpotApplication, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var row publicSpotApplicationRepositoryModel
	if err := r.db.WithContext(ctx).Preload("User").Where("id = ?", applicationID).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sqlx.ErrNotFound
		}
		return nil, err
	}
	item, err := publicSpotApplicationFromRepositoryModel(row)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ReviewPublicSpotApplication 审核申请；通过时合并到已有钓点或创建新的公共钓点。
func (r *BusinessRepository) ReviewPublicSpotApplication(ctx context.Context, applicationID, reviewerUserID int64, status, reviewNote string, mergeSpotID int64, reviewedAt time.Time) (int64, error) {
	if r.db == nil {
		return 0, errors.New("GORM database is not configured")
	}
	var affected int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current publicSpotApplicationRepositoryModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", applicationID).Take(&current).Error; err != nil {
			return err
		}
		if current.Status != "pending" {
			affected = 1
			return nil
		}
		updates := map[string]any{
			"status": status, "review_note": strings.TrimSpace(reviewNote), "reviewer_user_id": reviewerUserID,
			"reviewed_at": reviewedAt, "updated_at": reviewedAt,
		}
		if status == "approved" {
			mergedCode, err := ensureApprovedPublicSpot(tx, current, mergeSpotID)
			if err != nil {
				return err
			}
			updates["merged_spot_id"] = mergeSpotID
			updates["merged_spot_code"] = mergedCode
		}
		result := tx.Model(&publicSpotApplicationRepositoryModel{}).Where("id = ? AND status = ?", applicationID, "pending").Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		affected = result.RowsAffected
		return nil
	})
	return affected, err
}

type publicSpotApplicationRepositoryModel struct {
	ID              int64                          `gorm:"column:id;primaryKey"`
	ApplicationCode string                         `gorm:"column:application_code"`
	UserID          int64                          `gorm:"column:user_id"`
	SourceSpotCode  string                         `gorm:"column:source_spot_code"`
	Name            string                         `gorm:"column:name"`
	Address         string                         `gorm:"column:address"`
	Latitude        float64                        `gorm:"column:latitude"`
	Longitude       float64                        `gorm:"column:longitude"`
	Description     string                         `gorm:"column:description"`
	ChargeType      string                         `gorm:"column:charge_type"`
	TargetSpecies   string                         `gorm:"column:target_species_json"`
	ImageURLs       string                         `gorm:"column:image_urls_json"`
	Status          string                         `gorm:"column:status"`
	ReviewNote      string                         `gorm:"column:review_note"`
	ReviewerUserID  *int64                         `gorm:"column:reviewer_user_id"`
	ReviewedAt      *time.Time                     `gorm:"column:reviewed_at"`
	MergedSpotID    int64                          `gorm:"column:merged_spot_id"`
	MergedSpotCode  string                         `gorm:"column:merged_spot_code"`
	CreatedAt       time.Time                      `gorm:"column:created_at"`
	User            publicSpotApplicationUserModel `gorm:"foreignKey:UserID;references:ID"`
}

func (publicSpotApplicationRepositoryModel) TableName() string { return "public_spot_applications" }

type publicSpotApplicationUserModel struct {
	ID       int64  `gorm:"column:id;primaryKey"`
	Nickname string `gorm:"column:nickname"`
}

func (publicSpotApplicationUserModel) TableName() string { return "users" }

type publicSpotApplicationSpotWriteModel struct {
	ID              int64      `gorm:"column:id;primaryKey"`
	SpotCode        string     `gorm:"column:spot_code"`
	Name            string     `gorm:"column:name"`
	CoverImageURL   string     `gorm:"column:cover_image_url"`
	Address         string     `gorm:"column:address"`
	Latitude        float64    `gorm:"column:latitude"`
	Longitude       float64    `gorm:"column:longitude"`
	H3Cell          string     `gorm:"column:h3_cell"`
	H3DuplicateCell string     `gorm:"column:h3_duplicate_cell"`
	TagText         string     `gorm:"column:tag_text"`
	TagType         string     `gorm:"column:tag_type"`
	Scene           string     `gorm:"column:scene"`
	SourceType      string     `gorm:"column:source_type"`
	VisibleStatus   int64      `gorm:"column:visible_status"`
	PublishStatus   int64      `gorm:"column:publish_status"`
	PublisherUserID int64      `gorm:"column:publisher_user_id"`
	PublishedAt     *time.Time `gorm:"column:published_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (publicSpotApplicationSpotWriteModel) TableName() string { return "fishing_spots" }

type publicSpotApplicationSpeciesWriteModel struct {
	ID          int64  `gorm:"column:id;primaryKey"`
	SpotID      int64  `gorm:"column:spot_id"`
	SpeciesName string `gorm:"column:species_name"`
	SortOrder   int64  `gorm:"column:sort_order"`
}

func (publicSpotApplicationSpeciesWriteModel) TableName() string { return "fishing_spot_species" }

func publicSpotApplicationFromRepositoryModel(row publicSpotApplicationRepositoryModel) (model.PublicSpotApplication, error) {
	var species, images []string
	if row.TargetSpecies != "" {
		if err := json.Unmarshal([]byte(row.TargetSpecies), &species); err != nil {
			return model.PublicSpotApplication{}, err
		}
	}
	if row.ImageURLs != "" {
		if err := json.Unmarshal([]byte(row.ImageURLs), &images); err != nil {
			return model.PublicSpotApplication{}, err
		}
	}
	return model.PublicSpotApplication{
		ID: row.ID, ApplicationCode: row.ApplicationCode, UserID: row.UserID, UserNickname: row.User.Nickname,
		SourceSpotCode: row.SourceSpotCode, Name: row.Name, Address: row.Address, Latitude: row.Latitude,
		Longitude: row.Longitude, Description: row.Description, ChargeType: row.ChargeType,
		TargetSpecies: species, ImageURLs: images, Status: row.Status, ReviewNote: row.ReviewNote,
		ReviewerUserID: row.ReviewerUserID, ReviewedAt: row.ReviewedAt, MergedSpotID: row.MergedSpotID,
		MergedSpotCode: row.MergedSpotCode, CreatedAt: row.CreatedAt,
	}, nil
}

func ensureApprovedPublicSpot(tx *gorm.DB, application publicSpotApplicationRepositoryModel, mergeSpotID int64) (string, error) {
	if mergeSpotID > 0 {
		var existing publicSpotApplicationSpotWriteModel
		if err := tx.Where("id = ? AND publish_status = ? AND visible_status = ?", mergeSpotID, 1, 1).Take(&existing).Error; err != nil {
			return "", err
		}
		return existing.SpotCode, nil
	}
	cell, err := h3.LatLngToCell(h3.NewLatLng(application.Latitude, application.Longitude), 8)
	if err != nil {
		return "", err
	}
	duplicateCell, err := h3.LatLngToCell(h3.NewLatLng(application.Latitude, application.Longitude), 11)
	if err != nil {
		return "", err
	}
	var duplicate publicSpotApplicationSpotWriteModel
	err = tx.Where("h3_duplicate_cell = ? AND publish_status = ? AND visible_status = ?", duplicateCell.String(), 1, 1).Take(&duplicate).Error
	if err == nil {
		return "", ErrPossibleDuplicatePublicSpot
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	var images []string
	if application.ImageURLs != "" {
		if err := json.Unmarshal([]byte(application.ImageURLs), &images); err != nil {
			return "", err
		}
	}
	coverImageURL := ""
	if len(images) > 0 {
		coverImageURL = images[0]
	}
	now := time.Now()
	spot := publicSpotApplicationSpotWriteModel{
		SpotCode: "public_" + application.ApplicationCode, Name: application.Name,
		CoverImageURL: coverImageURL, Address: application.Address, Latitude: application.Latitude,
		Longitude: application.Longitude, H3Cell: cell.String(), H3DuplicateCell: duplicateCell.String(),
		TagType: "user_application", Scene: application.Description, SourceType: "official",
		VisibleStatus: 1, PublishStatus: 1, PublisherUserID: application.UserID, PublishedAt: &now,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := tx.Create(&spot).Error; err != nil {
		return "", err
	}
	var species []string
	if application.TargetSpecies != "" {
		if err := json.Unmarshal([]byte(application.TargetSpecies), &species); err != nil {
			return "", err
		}
	}
	if len(species) > 0 {
		rows := make([]publicSpotApplicationSpeciesWriteModel, len(species))
		for index, name := range species {
			rows[index] = publicSpotApplicationSpeciesWriteModel{SpotID: spot.ID, SpeciesName: name, SortOrder: int64(index + 1)}
		}
		if err := tx.Create(&rows).Error; err != nil {
			return "", err
		}
	}
	return spot.SpotCode, nil
}

type mediaAssetRepositoryModel struct {
	ID               int64                           `gorm:"column:id;primaryKey"`
	RecordID         int64                           `gorm:"column:record_id"`
	FileID           string                          `gorm:"column:file_id"`
	ImageURL         string                          `gorm:"column:image_url"`
	Width            int                             `gorm:"column:width"`
	Height           int                             `gorm:"column:height"`
	ModerationStatus string                          `gorm:"column:moderation_status"`
	ModerationNote   string                          `gorm:"column:moderation_note"`
	CreatedAt        time.Time                       `gorm:"column:created_at"`
	Record           mediaAssetRecordRepositoryModel `gorm:"foreignKey:RecordID;references:ID"`
}

func (mediaAssetRepositoryModel) TableName() string { return "fishing_record_images" }

type mediaAssetRecordRepositoryModel struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	RecordCode   string `gorm:"column:record_code"`
	UserID       int64  `gorm:"column:user_id"`
	SpotName     string `gorm:"column:spot_name"`
	VisibleScope string `gorm:"column:visible_scope"`
}

func (mediaAssetRecordRepositoryModel) TableName() string { return "fishing_records" }

func (r *BusinessRepository) ListMediaAssets(ctx context.Context, filter MediaAssetListFilter) ([]model.MediaAsset, int64, error) {
	if r.db == nil {
		return nil, 0, errors.New("GORM database is not configured")
	}
	query := r.db.WithContext(ctx).Model(&mediaAssetRepositoryModel{}).
		Joins("Record").Where("Record.visible_scope = ?", "public")
	if filter.Status != "" {
		query = query.Where("fishing_record_images.moderation_status = ?", filter.Status)
	}
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("Record.record_code LIKE ? OR Record.spot_name LIKE ?", keyword, keyword)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.MediaAsset{}, 0, nil
	}
	var rows []mediaAssetRepositoryModel
	if err := query.Order("fishing_record_images.created_at DESC").Order("fishing_record_images.id DESC").
		Limit(int(filter.Limit)).Offset(int(filter.Offset)).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]model.MediaAsset, len(rows))
	for index, row := range rows {
		items[index] = mediaAssetFromRepositoryModel(row)
	}
	return items, total, nil
}

func (r *BusinessRepository) FindMediaAssetByID(ctx context.Context, mediaID int64) (*model.MediaAsset, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var row mediaAssetRepositoryModel
	err := r.db.WithContext(ctx).Model(&mediaAssetRepositoryModel{}).Joins("Record").
		Where("fishing_record_images.id = ?", mediaID).Where("Record.visible_scope = ?", "public").Take(&row).Error
	if err != nil {
		return nil, err
	}
	item := mediaAssetFromRepositoryModel(row)
	return &item, nil
}

func (r *BusinessRepository) ReviewMediaAsset(ctx context.Context, mediaID, reviewerUserID int64, status, note string, reviewedAt time.Time) (int64, error) {
	if r.db == nil {
		return 0, errors.New("GORM database is not configured")
	}
	var rows int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current mediaAssetRepositoryModel
		if err := tx.Model(&mediaAssetRepositoryModel{}).Select("fishing_record_images.id, fishing_record_images.record_id").Joins("Record").
			Where("fishing_record_images.id = ?", mediaID).Where("Record.visible_scope = ?", "public").
			Where("fishing_record_images.moderation_status = ?", "pending").Take(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		result := tx.Model(&mediaAssetRepositoryModel{}).
			Where("id = ? AND moderation_status = ?", mediaID, "pending").
			Updates(map[string]any{
				"moderation_status": status, "moderation_note": note,
				"reviewer_user_id": reviewerUserID, "reviewed_at": reviewedAt,
			})
		if result.Error != nil {
			return result.Error
		}
		rows = result.RowsAffected
		return nil
	})
	return rows, err
}

func mediaAssetFromRepositoryModel(row mediaAssetRepositoryModel) model.MediaAsset {
	return model.MediaAsset{
		ID: row.ID, RecordID: row.RecordID, RecordCode: row.Record.RecordCode,
		UserID: row.Record.UserID, SpotName: row.Record.SpotName, ImageURL: row.ImageURL,
		Width: row.Width, Height: row.Height, ModerationStatus: row.ModerationStatus,
		ModerationNote: row.ModerationNote, CreatedAt: row.CreatedAt,
	}
}

type spotCorrectionRepositoryModel struct {
	ID             int64                             `gorm:"column:id;primaryKey"`
	CorrectionCode string                            `gorm:"column:correction_code"`
	UserID         int64                             `gorm:"column:user_id"`
	SpotID         int64                             `gorm:"column:spot_id"`
	SpotCode       string                            `gorm:"column:spot_code"`
	SpotName       string                            `gorm:"column:spot_name"`
	Content        string                            `gorm:"column:content"`
	Status         string                            `gorm:"column:status"`
	ReviewNote     string                            `gorm:"column:review_note"`
	ReviewerUserID sql.NullInt64                     `gorm:"column:reviewer_user_id"`
	ReviewedAt     sql.NullTime                      `gorm:"column:reviewed_at"`
	CreatedAt      time.Time                         `gorm:"column:created_at"`
	UpdatedAt      time.Time                         `gorm:"column:updated_at"`
	User           spotCorrectionUserRepositoryModel `gorm:"foreignKey:UserID;references:ID"`
	Spot           spotCorrectionSpotRepositoryModel `gorm:"foreignKey:SpotID;references:ID"`
}

func (spotCorrectionRepositoryModel) TableName() string { return "spot_corrections" }

type spotCorrectionUserRepositoryModel struct {
	ID       int64  `gorm:"column:id;primaryKey"`
	Nickname string `gorm:"column:nickname"`
}

func (spotCorrectionUserRepositoryModel) TableName() string { return "users" }

type spotCorrectionSpotRepositoryModel struct {
	ID                int64                             `gorm:"column:id;primaryKey"`
	SpotCode          string                            `gorm:"column:spot_code"`
	Name              string                            `gorm:"column:name"`
	CoverImageURL     string                            `gorm:"column:cover_image_url"`
	Province          string                            `gorm:"column:province"`
	City              string                            `gorm:"column:city"`
	District          string                            `gorm:"column:district"`
	Address           string                            `gorm:"column:address"`
	Latitude          float64                           `gorm:"column:latitude"`
	Longitude         float64                           `gorm:"column:longitude"`
	TagText           string                            `gorm:"column:tag_text"`
	TagType           string                            `gorm:"column:tag_type"`
	FishingIndex      int64                             `gorm:"column:fishing_index"`
	Scene             string                            `gorm:"column:scene"`
	SceneHint         string                            `gorm:"column:scene_hint"`
	SourceType        string                            `gorm:"column:source_type"`
	VisibleStatus     int64                             `gorm:"column:visible_status"`
	PublishStatus     int64                             `gorm:"column:publish_status"`
	PublisherUserID   int64                             `gorm:"column:publisher_user_id"`
	PublisherNickname string                            `gorm:"-"`
	PublishedAt       sql.NullTime                      `gorm:"column:published_at"`
	CreatedAt         time.Time                         `gorm:"column:created_at"`
	UpdatedAt         time.Time                         `gorm:"column:updated_at"`
	Publisher         spotCorrectionUserRepositoryModel `gorm:"foreignKey:PublisherUserID;references:ID"`
}

func (spotCorrectionSpotRepositoryModel) TableName() string { return "fishing_spots" }

type fishingSpotSpeciesRepositoryModel struct {
	ID          int64  `gorm:"column:id;primaryKey"`
	SpotID      int64  `gorm:"column:spot_id"`
	SpeciesName string `gorm:"column:species_name"`
	SortOrder   int64  `gorm:"column:sort_order"`
}

func (fishingSpotSpeciesRepositoryModel) TableName() string { return "fishing_spot_species" }

type fishingSpotTipRepositoryModel struct {
	ID            int64  `gorm:"column:id;primaryKey"`
	SpotID        int64  `gorm:"column:spot_id"`
	TipText       string `gorm:"column:tip_text"`
	SortOrder     int64  `gorm:"column:sort_order"`
	VisibleStatus int64  `gorm:"column:visible_status"`
}

func (fishingSpotTipRepositoryModel) TableName() string { return "fishing_spot_tips" }

type userProfileBusinessRepositoryModel struct {
	UserID               int64  `gorm:"column:user_id;primaryKey"`
	LevelText            string `gorm:"column:level_text"`
	ProfileDesc          string `gorm:"column:profile_desc"`
	LocationText         string `gorm:"column:location_text"`
	City                 string `gorm:"column:city"`
	District             string `gorm:"column:district"`
	StreakWeeks          int64  `gorm:"column:streak_weeks"`
	PreferredSpeciesText string `gorm:"column:preferred_species_text"`
}

func (userProfileBusinessRepositoryModel) TableName() string { return "user_profiles" }

type businessUserRepositoryModel struct {
	ID          int64                              `gorm:"column:id;primaryKey"`
	Nickname    string                             `gorm:"column:nickname"`
	AvatarURL   string                             `gorm:"column:avatar"`
	Status      int64                              `gorm:"column:status"`
	LastLoginAt sql.NullTime                       `gorm:"column:last_login_at"`
	CreatedAt   time.Time                          `gorm:"column:created_at"`
	UpdatedAt   time.Time                          `gorm:"column:updated_at"`
	Profile     userProfileBusinessRepositoryModel `gorm:"foreignKey:UserID;references:ID"`
}

func (businessUserRepositoryModel) TableName() string { return "users" }

func fishingSpotFromRepositoryModel(row spotCorrectionSpotRepositoryModel) model.FishingSpot {
	return model.FishingSpot{
		ID: row.ID, SpotCode: row.SpotCode, Name: row.Name, CoverImageURL: row.CoverImageURL,
		Province: row.Province, City: row.City, District: row.District, Address: row.Address,
		Latitude: row.Latitude, Longitude: row.Longitude, TagText: row.TagText, TagType: row.TagType,
		FishingIndex: row.FishingIndex, Scene: row.Scene, SceneHint: row.SceneHint, SourceType: row.SourceType,
		VisibleStatus: row.VisibleStatus, PublishStatus: row.PublishStatus, PublisherUserID: row.PublisherUserID,
		PublisherNickname: row.Publisher.Nickname, PublishedAt: row.PublishedAt, CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func userProfileFromRepositoryModel(row businessUserRepositoryModel) model.UserProfile {
	return model.UserProfile{
		UserID: row.ID, Nickname: row.Nickname, AvatarURL: row.AvatarURL, Status: row.Status,
		LastLoginAt: row.LastLoginAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		LevelText: row.Profile.LevelText, ProfileDesc: row.Profile.ProfileDesc,
		LocationText: row.Profile.LocationText, City: row.Profile.City, District: row.Profile.District,
		StreakWeeks: row.Profile.StreakWeeks, PreferredSpeciesText: row.Profile.PreferredSpeciesText,
	}
}

var spotCorrectionFieldColumns = map[string]string{
	"name":      "name",
	"province":  "province",
	"city":      "city",
	"district":  "district",
	"address":   "address",
	"tagText":   "tag_text",
	"tagType":   "tag_type",
	"scene":     "scene",
	"sceneHint": "scene_hint",
	"latitude":  "latitude",
	"longitude": "longitude",
}

func spotCorrectionFromRepositoryModel(row spotCorrectionRepositoryModel) model.SpotCorrection {
	result := model.SpotCorrection{
		ID: row.ID, CorrectionCode: row.CorrectionCode, UserID: row.UserID, UserNickname: row.User.Nickname,
		SpotID: row.SpotID, SpotCode: row.SpotCode, SpotName: row.SpotName, Content: row.Content,
		Status: row.Status, ReviewNote: row.ReviewNote, ReviewerUserID: row.ReviewerUserID,
		ReviewedAt: row.ReviewedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.Spot.ID > 0 {
		result.CurrentSpot = &model.FishingSpot{
			ID: row.Spot.ID, SpotCode: row.Spot.SpotCode, Name: row.Spot.Name, CoverImageURL: row.Spot.CoverImageURL,
			Province: row.Spot.Province, City: row.Spot.City, District: row.Spot.District, Address: row.Spot.Address,
			Latitude: row.Spot.Latitude, Longitude: row.Spot.Longitude, TagText: row.Spot.TagText, TagType: row.Spot.TagType,
			FishingIndex: row.Spot.FishingIndex, Scene: row.Spot.Scene, SceneHint: row.Spot.SceneHint,
			SourceType: row.Spot.SourceType, VisibleStatus: row.Spot.VisibleStatus, PublishStatus: row.Spot.PublishStatus,
			PublisherUserID: row.Spot.PublisherUserID, PublishedAt: row.Spot.PublishedAt,
			CreatedAt: row.Spot.CreatedAt, UpdatedAt: row.Spot.UpdatedAt,
		}
	}
	return result
}

func (r *BusinessRepository) ReviewSpotCorrection(ctx context.Context, correctionID, reviewerUserID int64, status, reviewNote string, reviewedAt time.Time) (int64, error) {
	return r.ReviewSpotCorrectionWithChanges(ctx, correctionID, reviewerUserID, status, reviewNote, nil, reviewedAt)
}

func (r *BusinessRepository) ReviewSpotCorrectionWithChanges(ctx context.Context, correctionID, reviewerUserID int64, status, reviewNote string, changes map[string]string, reviewedAt time.Time) (int64, error) {
	if r.db == nil {
		return 0, errors.New("GORM database is not configured")
	}
	if len(changes) == 0 {
		result := r.db.WithContext(ctx).Model(&spotCorrectionRepositoryModel{}).
			Where("id = ? AND status = ?", correctionID, "pending").
			Updates(map[string]any{
				"status": status, "review_note": reviewNote, "reviewer_user_id": reviewerUserID,
				"reviewed_at": reviewedAt, "updated_at": reviewedAt,
			})
		return result.RowsAffected, result.Error
	}
	var rows int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var correction spotCorrectionRepositoryModel
		if err := tx.Where("id = ? AND status = ?", correctionID, "pending").Take(&correction).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				rows = 0
				return nil
			}
			return err
		}
		if len(changes) > 0 && status == "approved" {
			spotChanges := make(map[string]any, len(changes))
			for field, value := range changes {
				column, ok := spotCorrectionFieldColumns[field]
				if !ok {
					return fmt.Errorf("unsupported spot correction field %q", field)
				}
				if field == "latitude" || field == "longitude" {
					coordinate, err := strconv.ParseFloat(value, 64)
					if err != nil {
						return fmt.Errorf("spot correction field %q must be a number", field)
					}
					spotChanges[column] = coordinate
					continue
				}
				spotChanges[column] = value
			}
			if err := tx.Model(&spotCorrectionSpotRepositoryModel{}).Where("id = ?", correction.SpotID).Updates(spotChanges).Error; err != nil {
				return err
			}
		}
		result := tx.Model(&spotCorrectionRepositoryModel{}).
			Where("id = ? AND status = ?", correctionID, "pending").
			Updates(map[string]any{
				"status": status, "review_note": reviewNote, "reviewer_user_id": reviewerUserID,
				"reviewed_at": reviewedAt, "updated_at": reviewedAt,
			})
		if result.Error != nil {
			return result.Error
		}
		rows = result.RowsAffected
		return nil
	})
	return rows, err
}

func (r *BusinessRepository) ListReviewTasks(ctx context.Context, filter ReviewTaskListFilter) ([]model.ReviewTask, int64, error) {
	if r.db == nil {
		return nil, 0, errors.New("GORM database is not configured")
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	if filter.Limit <= 0 {
		return []model.ReviewTask{}, 0, nil
	}
	maxInt := int64(int(^uint(0) >> 1))
	if filter.Limit > maxInt || offset > maxInt-filter.Limit {
		return nil, 0, errors.New("review task page is out of range")
	}
	take := int(offset + filter.Limit)
	tasks := make([]model.ReviewTask, 0)
	var total int64

	if filter.TaskType == "" || filter.TaskType == "spot_correction" {
		query := r.db.WithContext(ctx).Model(&reviewCorrectionTaskRow{})
		if filter.Keyword != "" {
			keyword := "%" + filter.Keyword + "%"
			query = query.Where("(spot_name LIKE ? OR correction_code LIKE ? OR content LIKE ?)", keyword, keyword, keyword)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return nil, 0, err
		}
		total += count
		if count > 0 {
			var rows []reviewCorrectionTaskRow
			if err := query.Order("created_at ASC").Order("id ASC").Limit(take).Find(&rows).Error; err != nil {
				return nil, 0, err
			}
			for _, row := range rows {
				tasks = append(tasks, model.ReviewTask{
					TaskID: "correction:" + strconv.FormatInt(row.ID, 10), TaskType: "spot_correction",
					ResourceID: row.ID, ResourceCode: row.CorrectionCode, Title: row.SpotName,
					Summary: row.Content, Status: row.Status, CreatedAt: row.CreatedAt,
				})
			}
		}
	}

	if filter.TaskType == "" || filter.TaskType == "spot" || filter.TaskType == "user_spot" {
		query, applicable := reviewSpotTaskQuery(r.db.WithContext(ctx), filter)
		if applicable {
			var count int64
			if err := query.Count(&count).Error; err != nil {
				return nil, 0, err
			}
			total += count
			if count > 0 {
				var rows []reviewSpotTaskRow
				if err := query.Order("created_at ASC").Order("id ASC").Limit(take).Find(&rows).Error; err != nil {
					return nil, 0, err
				}
				for _, row := range rows {
					taskType, status := "spot", "pending"
					if row.SourceType == "user" {
						taskType = "user_spot"
					}
					switch row.PublishStatus {
					case 1:
						status = "approved"
					case 2:
						status = "rejected"
					}
					tasks = append(tasks, model.ReviewTask{
						TaskID: "spot:" + strconv.FormatInt(row.ID, 10), TaskType: taskType,
						ResourceID: row.ID, ResourceCode: row.SpotCode, Title: row.Name,
						Summary: row.Address, Status: status, CreatedAt: row.CreatedAt,
					})
				}
			}
		}
	}

	if filter.TaskType == "" || filter.TaskType == "public_report" {
		query, applicable := reviewPublicReportTaskQuery(r.db.WithContext(ctx), filter)
		if applicable {
			var count int64
			if err := query.Count(&count).Error; err != nil {
				return nil, 0, err
			}
			total += count
			if count > 0 {
				var rows []reviewPublicReportTaskRow
				if err := query.Order("created_at ASC").Order("id ASC").Limit(take).Find(&rows).Error; err != nil {
					return nil, 0, err
				}
				for _, row := range rows {
					tasks = append(tasks, model.ReviewTask{
						TaskID: "public_report:" + strconv.FormatInt(row.ID, 10), TaskType: "public_report",
						ResourceID: row.ID, ResourceCode: row.RecordCode, Title: row.Title,
						Summary: publicReportTaskSummary(row), Status: row.ModerationStatus, CreatedAt: row.CreatedAt,
					})
				}
			}
		}
	}

	sort.Slice(tasks, func(left, right int) bool {
		if tasks[left].CreatedAt.Equal(tasks[right].CreatedAt) {
			if tasks[left].ResourceID == tasks[right].ResourceID {
				return tasks[left].TaskType < tasks[right].TaskType
			}
			return tasks[left].ResourceID < tasks[right].ResourceID
		}
		return tasks[left].CreatedAt.Before(tasks[right].CreatedAt)
	})
	start := int(offset)
	if start >= len(tasks) {
		return []model.ReviewTask{}, total, nil
	}
	end := start + int(filter.Limit)
	if end > len(tasks) {
		end = len(tasks)
	}
	return tasks[start:end], total, nil
}

func reviewSpotTaskQuery(db *gorm.DB, filter ReviewTaskListFilter) (*gorm.DB, bool) {
	query := db.Model(&reviewSpotTaskRow{})
	if filter.TaskType == "user_spot" {
		query = query.Where("source_type = ?", "user")
	} else if filter.TaskType == "spot" {
		query = query.Where("(source_type IS NULL OR source_type <> ?)", "user")
	}
	switch filter.Status {
	case "approved":
		query = query.Where("publish_status = ?", 1)
	case "rejected":
		query = query.Where("publish_status = ?", 2)
	case "pending":
		query = query.Where("publish_status NOT IN ?", []int64{1, 2})
	case "":
	default:
		return query, false
	}
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("(name LIKE ? OR spot_code LIKE ? OR address LIKE ?)", keyword, keyword, keyword)
	}
	return query, true
}

func reviewPublicReportTaskQuery(db *gorm.DB, filter ReviewTaskListFilter) (*gorm.DB, bool) {
	query := db.Model(&reviewPublicReportTaskRow{}).Where("visible_scope = ?", "public").Where("moderation_status IN ?", []string{"pending", "rejected"})
	if filter.Status != "" {
		if filter.Status != "pending" && filter.Status != "rejected" {
			return query, false
		}
		query = query.Where("moderation_status = ?", filter.Status)
	}
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("(title LIKE ? OR record_code LIKE ? OR spot_name LIKE ? OR fish_species_text LIKE ?)", keyword, keyword, keyword, keyword)
		if count, err := strconv.ParseInt(strings.TrimSpace(strings.TrimSuffix(filter.Keyword, "尾")), 10, 64); err == nil {
			query = query.Or("visible_scope = ? AND moderation_status IN ? AND fish_count = ?", "public", []string{"pending", "rejected"}, count)
		}
		if start, end, ok := reviewDateRange(filter.Keyword); ok {
			query = query.Or("visible_scope = ? AND moderation_status IN ? AND record_date >= ? AND record_date < ?", "public", []string{"pending", "rejected"}, start, end)
		}
	}
	return query, true
}

func reviewDateRange(value string) (time.Time, time.Time, bool) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{"2006-01-02", "2006-01", "2006"} {
		start, err := time.Parse(layout, value)
		if err != nil {
			continue
		}
		switch layout {
		case "2006-01-02":
			return start, start.AddDate(0, 0, 1), true
		case "2006-01":
			return start, start.AddDate(0, 1, 0), true
		default:
			return start, start.AddDate(1, 0, 0), true
		}
	}
	return time.Time{}, time.Time{}, false
}

func publicReportTaskSummary(row reviewPublicReportTaskRow) string {
	parts := make([]string, 0, 4)
	if value := strings.TrimSpace(row.SpotName); value != "" {
		parts = append(parts, value)
	}
	if !row.RecordDate.IsZero() {
		parts = append(parts, row.RecordDate.Format("2006-01-02"))
	}
	if value := strings.TrimSpace(row.FishSpeciesText); value != "" {
		parts = append(parts, value)
	}
	if row.FishCount > 0 {
		parts = append(parts, strconv.FormatInt(row.FishCount, 10)+" 尾")
	}
	return strings.Join(parts, " · ")
}

type reviewCorrectionTaskRow struct {
	ID             int64     `gorm:"column:id"`
	CorrectionCode string    `gorm:"column:correction_code"`
	SpotName       string    `gorm:"column:spot_name"`
	Content        string    `gorm:"column:content"`
	Status         string    `gorm:"column:status"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (reviewCorrectionTaskRow) TableName() string { return "spot_corrections" }

type reviewSpotTaskRow struct {
	ID            int64     `gorm:"column:id"`
	SpotCode      string    `gorm:"column:spot_code"`
	Name          string    `gorm:"column:name"`
	Address       string    `gorm:"column:address"`
	SourceType    string    `gorm:"column:source_type"`
	PublishStatus int64     `gorm:"column:publish_status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (reviewSpotTaskRow) TableName() string { return "fishing_spots" }

type reviewPublicReportTaskRow struct {
	ID               int64     `gorm:"column:id"`
	RecordCode       string    `gorm:"column:record_code"`
	Title            string    `gorm:"column:title"`
	SpotName         string    `gorm:"column:spot_name"`
	RecordDate       time.Time `gorm:"column:record_date"`
	FishSpeciesText  string    `gorm:"column:fish_species_text"`
	FishCount        int64     `gorm:"column:fish_count"`
	ModerationStatus string    `gorm:"column:moderation_status"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (reviewPublicReportTaskRow) TableName() string { return "fishing_records" }

func (r *BusinessRepository) ReviewPublicFishingReport(ctx context.Context, recordID int64, status string) (int64, error) {
	if r.db == nil {
		return 0, errors.New("GORM database is not configured")
	}
	result := r.db.WithContext(ctx).Model(&publicFishingReportReviewModel{}).
		Where("id = ? AND visible_scope = ? AND moderation_status IN ?", recordID, "public", []string{"pending", "rejected"}).
		Updates(map[string]any{"moderation_status": status, "updated_at": time.Now()})
	if result.Error != nil || result.RowsAffected > 0 {
		return result.RowsAffected, result.Error
	}
	var currentStatus string
	if err := r.db.WithContext(ctx).Model(&publicFishingReportReviewModel{}).
		Select("moderation_status").Where("id = ? AND visible_scope = ?", recordID, "public").Take(&currentStatus).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	if currentStatus == status {
		return 1, nil
	}
	return 0, nil
}

func (r *BusinessRepository) ListFishingSpots(ctx context.Context, filter FishingSpotListFilter) ([]model.FishingSpot, int64, error) {
	if r.db == nil {
		return nil, 0, errors.New("GORM database is not configured")
	}
	query := applyFishingSpotFilters(r.db.WithContext(ctx).Model(&spotCorrectionSpotRepositoryModel{}), filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.FishingSpot{}, 0, nil
	}
	var rows []spotCorrectionSpotRepositoryModel
	if err := query.Preload("Publisher").Order("updated_at DESC").Order("id DESC").
		Limit(int(filter.Limit)).Offset(int(filter.Offset)).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]model.FishingSpot, len(rows))
	for index, row := range rows {
		items[index] = fishingSpotFromRepositoryModel(row)
	}
	return items, total, nil
}

func (r *BusinessRepository) FindFishingSpotByID(ctx context.Context, spotID int64) (*model.FishingSpot, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var row spotCorrectionSpotRepositoryModel
	err := r.db.WithContext(ctx).Preload("Publisher").Where("id = ?", spotID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	item := fishingSpotFromRepositoryModel(row)
	return &item, nil
}

func (r *BusinessRepository) ListFishingSpotSpeciesBySpotIDs(ctx context.Context, spotIDs []int64) ([]model.FishingSpotSpecies, error) {
	if len(spotIDs) == 0 {
		return []model.FishingSpotSpecies{}, nil
	}
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var rows []fishingSpotSpeciesRepositoryModel
	if err := r.db.WithContext(ctx).Where("spot_id IN ?", spotIDs).
		Order("spot_id ASC").Order("sort_order ASC").Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]model.FishingSpotSpecies, len(rows))
	for index, row := range rows {
		items[index] = model.FishingSpotSpecies{SpotID: row.SpotID, SpeciesName: row.SpeciesName, SortOrder: row.SortOrder}
	}
	return items, nil
}

func (r *BusinessRepository) ListFishingSpotTipsBySpotIDs(ctx context.Context, spotIDs []int64) ([]model.FishingSpotTip, error) {
	if len(spotIDs) == 0 {
		return []model.FishingSpotTip{}, nil
	}

	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var rows []fishingSpotTipRepositoryModel
	if err := r.db.WithContext(ctx).Where("spot_id IN ? AND visible_status = ?", spotIDs, 1).
		Order("spot_id ASC").Order("sort_order ASC").Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]model.FishingSpotTip, len(rows))
	for index, row := range rows {
		items[index] = model.FishingSpotTip{SpotID: row.SpotID, TipText: row.TipText, SortOrder: row.SortOrder, VisibleStatus: row.VisibleStatus}
	}
	return items, nil
}

func (r *BusinessRepository) ApproveFishingSpot(ctx context.Context, spotID int64, publishStatus int64, publishedAt time.Time) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current model.FishingSpot
		if err := tx.Select("id", "published_at").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", spotID).Take(&current).Error; err != nil {
			return err
		}
		if current.PublishedAt.Valid {
			publishedAt = current.PublishedAt.Time
		}
		return tx.Model(&model.FishingSpot{}).Where("id = ?", spotID).Updates(map[string]any{
			"visible_status": 1, "publish_status": publishStatus, "published_at": publishedAt,
			"updated_at": time.Now(),
		}).Error
	})
}

func (r *BusinessRepository) RejectFishingSpot(ctx context.Context, spotID int64, publishStatus int64) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Model(&model.FishingSpot{}).Where("id = ?", spotID).Updates(map[string]any{
		"visible_status": 0, "publish_status": publishStatus, "published_at": nil, "updated_at": time.Now(),
	}).Error
}

type publicFishingReportReviewModel struct {
	ID               int64  `gorm:"column:id;primaryKey"`
	VisibleScope     string `gorm:"column:visible_scope"`
	ModerationStatus string `gorm:"column:moderation_status"`
}

func (publicFishingReportReviewModel) TableName() string { return "fishing_records" }

func (r *BusinessRepository) ListArticles(ctx context.Context, filter ArticleListFilter) ([]model.DiscoverArticle, int64, error) {
	if r.db == nil {
		return nil, 0, errors.New("GORM database is not configured")
	}
	query := applyArticleFilters(r.db.WithContext(ctx).Model(&model.DiscoverArticle{}), filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.DiscoverArticle{}, 0, nil
	}
	var items []model.DiscoverArticle
	if err := query.Order("sort_order ASC").Order("id DESC").Limit(int(filter.Limit)).Offset(int(filter.Offset)).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func applyBannerFilters(query *gorm.DB, filter BannerListFilter) *gorm.DB {
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("(banner_code LIKE ? OR tag LIKE ? OR title LIKE ? OR description LIKE ? OR location_name LIKE ?)", keyword, keyword, keyword, keyword, keyword)
	}
	if filter.VisibleStatus != nil {
		query = query.Where("visible_status = ?", *filter.VisibleStatus)
	}
	return query
}

func (r *BusinessRepository) ListBanners(ctx context.Context, filter BannerListFilter) ([]model.DiscoverEventBanner, int64, error) {
	if r.db == nil {
		return nil, 0, errors.New("GORM database is not configured")
	}
	query := applyBannerFilters(r.db.WithContext(ctx).Model(&model.DiscoverEventBanner{}), filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.DiscoverEventBanner{}, 0, nil
	}
	items := make([]model.DiscoverEventBanner, 0)
	if err := query.Order("sort_order ASC").Order("id DESC").Limit(int(filter.Limit)).Offset(int(filter.Offset)).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *BusinessRepository) FindBannerByID(ctx context.Context, bannerID int64) (*model.DiscoverEventBanner, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var item model.DiscoverEventBanner
	err := r.db.WithContext(ctx).Model(&model.DiscoverEventBanner{}).Where("id = ?", bannerID).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *BusinessRepository) CreateBanner(ctx context.Context, item *model.DiscoverEventBanner) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Model(&model.DiscoverEventBanner{}).Create(item).Error
}

func (r *BusinessRepository) UpdateBanner(ctx context.Context, bannerID int64, item model.DiscoverEventBanner) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Model(&model.DiscoverEventBanner{}).Where("id = ?", bannerID).Updates(map[string]any{
		"tag": item.Tag, "title": item.Title, "description": item.Description,
		"deadline_text": item.DeadlineText, "deadline_at": item.DeadlineAt,
		"location_name": item.LocationName, "sort_order": item.SortOrder,
		"visible_status": item.VisibleStatus, "updated_at": time.Now(),
	}).Error
}

func (r *BusinessRepository) UpdateBannerVisibility(ctx context.Context, bannerID, visibleStatus int64, publishedAt *time.Time) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Model(&model.DiscoverEventBanner{}).Where("id = ?", bannerID).Updates(map[string]any{
		"visible_status": visibleStatus, "published_at": publishedAt, "updated_at": time.Now(),
	}).Error
}

func (r *BusinessRepository) FindArticleByID(ctx context.Context, articleID int64) (*model.DiscoverArticle, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var item model.DiscoverArticle
	err := r.db.WithContext(ctx).Model(&model.DiscoverArticle{}).Where("id = ?", articleID).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *BusinessRepository) FindArticleDetailByCode(ctx context.Context, articleCode string) (*model.DiscoverContentDetail, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var item model.DiscoverContentDetail
	err := r.db.WithContext(ctx).Model(&model.DiscoverContentDetail{}).
		Where("content_type = ? AND content_code = ?", "article", articleCode).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *BusinessRepository) FindArticleDetailsByCodes(ctx context.Context, articleCodes []string) ([]model.DiscoverContentDetail, error) {
	if len(articleCodes) == 0 {
		return []model.DiscoverContentDetail{}, nil
	}

	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	items := make([]model.DiscoverContentDetail, 0)
	if err := r.db.WithContext(ctx).Model(&model.DiscoverContentDetail{}).
		Where("content_type = ? AND content_code IN ?", "article", articleCodes).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *BusinessRepository) CreateArticle(ctx context.Context, item model.DiscoverArticle, detail model.DiscoverContentDetail) (int64, error) {
	if r.db == nil {
		return 0, errors.New("GORM database is not configured")
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.DiscoverArticle{}).Create(&item).Error; err != nil {
			return err
		}
		detail.ContentType = "article"
		if err := tx.Model(&model.DiscoverContentDetail{}).Create(&detail).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (r *BusinessRepository) UpdateArticle(ctx context.Context, articleID int64, item model.DiscoverArticle, detail model.DiscoverContentDetail) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.DiscoverArticle{}).Where("id = ?", articleID).Updates(map[string]any{
			"short_label": item.ShortLabel, "title": item.Title, "description": item.Description,
			"theme": item.Theme, "sort_order": item.SortOrder, "visible_status": item.VisibleStatus,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}
		return upsertArticleDetail(tx, detail)
	})
}

func (r *BusinessRepository) PublishArticle(ctx context.Context, articleID int64, visibleStatus int64, publishedAt time.Time) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item model.DiscoverArticle
		err := tx.Model(&model.DiscoverArticle{}).Select("id", "published_at").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", articleID).Take(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if item.PublishedAt.Valid {
			publishedAt = item.PublishedAt.Time
		}
		return tx.Model(&model.DiscoverArticle{}).Where("id = ?", articleID).Updates(map[string]any{
			"publish_status": 1, "visible_status": visibleStatus, "published_at": publishedAt, "updated_at": time.Now(),
		}).Error
	})
}

func (r *BusinessRepository) OfflineArticle(ctx context.Context, articleID int64) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Model(&model.DiscoverArticle{}).Where("id = ?", articleID).Updates(map[string]any{
		"publish_status": 2, "visible_status": 0, "updated_at": time.Now(),
	}).Error
}

func (r *BusinessRepository) ListSpecies(ctx context.Context, filter SpeciesListFilter) ([]model.Species, int64, error) {
	if r.db == nil {
		return nil, 0, errors.New("GORM database is not configured")
	}
	query := applySpeciesFilters(r.db.WithContext(ctx).Model(&model.Species{}), filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.Species{}, 0, nil
	}
	var items []model.Species
	if err := query.Order("sort_order ASC").Order("id DESC").Limit(int(filter.Limit)).Offset(int(filter.Offset)).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *BusinessRepository) FindSpeciesByID(ctx context.Context, speciesID int64) (*model.Species, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var item model.Species
	err := r.db.WithContext(ctx).Model(&model.Species{}).Where("id = ?", speciesID).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *BusinessRepository) ListSpeciesDetailTipsBySpeciesIDs(ctx context.Context, speciesIDs []int64) ([]model.SpeciesDetailTip, error) {
	if len(speciesIDs) == 0 {
		return []model.SpeciesDetailTip{}, nil
	}

	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	items := make([]model.SpeciesDetailTip, 0)
	if err := r.db.WithContext(ctx).Model(&model.SpeciesDetailTip{}).
		Where("species_id IN ? AND visible_status = ?", speciesIDs, 1).
		Order("species_id ASC").Order("sort_order ASC").Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *BusinessRepository) CreateSpecies(ctx context.Context, item model.Species, tips []string) (int64, error) {
	if r.db == nil {
		return 0, errors.New("GORM database is not configured")
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Species{}).Create(&item).Error; err != nil {
			return err
		}
		return replaceSpeciesDetailTips(tx, item.ID, tips)
	})
	if err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (r *BusinessRepository) UpdateSpecies(ctx context.Context, speciesID int64, item model.Species, tips []string) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Species{}).Where("id = ?", speciesID).Updates(map[string]any{
			"name": item.Name, "alias": item.Alias, "category": item.Category, "water_layer": item.WaterLayer,
			"tag_text": item.TagText, "season_text": item.SeasonText, "best_window": item.BestWindow,
			"fishing_method": item.FishingMethod, "bait_text": item.BaitText, "description": item.Description,
			"highlight_text": item.HighlightText, "is_featured": item.IsFeatured, "sort_order": item.SortOrder,
			"visible_status": item.VisibleStatus, "updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}
		return replaceSpeciesDetailTips(tx, speciesID, tips)
	})
}

func (r *BusinessRepository) UpdateSpeciesVisibility(ctx context.Context, speciesID int64, visibleStatus int64) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Model(&model.Species{}).Where("id = ?", speciesID).Updates(map[string]any{
		"visible_status": visibleStatus, "updated_at": time.Now(),
	}).Error
}

func (r *BusinessRepository) ListUsers(ctx context.Context, filter UserListFilter) ([]model.UserProfile, int64, error) {
	if r.db == nil {
		return nil, 0, errors.New("GORM database is not configured")
	}
	query := applyUserFilters(r.db.WithContext(ctx).Model(&businessUserRepositoryModel{}).Joins("Profile"), filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.UserProfile{}, 0, nil
	}
	var rows []businessUserRepositoryModel
	if err := query.Order("users.updated_at DESC").Order("users.id DESC").
		Limit(int(filter.Limit)).Offset(int(filter.Offset)).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]model.UserProfile, len(rows))
	for index, row := range rows {
		items[index] = userProfileFromRepositoryModel(row)
	}
	return items, total, nil
}

func (r *BusinessRepository) FindUserByID(ctx context.Context, userID int64) (*model.UserProfile, error) {
	if r.db == nil {
		return nil, errors.New("GORM database is not configured")
	}
	var row businessUserRepositoryModel
	err := r.db.WithContext(ctx).Model(&businessUserRepositoryModel{}).Joins("Profile").Where("users.id = ?", userID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	item := userProfileFromRepositoryModel(row)
	return &item, nil
}

func (r *BusinessRepository) UpdateUserStatus(ctx context.Context, userID int64, status int64) error {
	if r.db == nil {
		return errors.New("GORM database is not configured")
	}
	return r.db.WithContext(ctx).Model(&businessUserRepositoryModel{}).Where("id = ?", userID).Updates(map[string]any{
		"status": status, "updated_at": time.Now(),
	}).Error
}

func applyFishingSpotFilters(query *gorm.DB, filter FishingSpotListFilter) *gorm.DB {
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where(clause.Or(
			clause.Like{Column: clause.Column{Name: "name"}, Value: keyword},
			clause.Like{Column: clause.Column{Name: "spot_code"}, Value: keyword},
			clause.Like{Column: clause.Column{Name: "address"}, Value: keyword},
		))
	}
	if filter.SourceType != "" {
		query = query.Where(clause.Eq{Column: clause.Column{Name: "source_type"}, Value: filter.SourceType})
	}
	if filter.VisibleStatus != nil {
		query = query.Where(clause.Eq{Column: clause.Column{Name: "visible_status"}, Value: *filter.VisibleStatus})
	}
	if filter.PublishStatus != nil {
		query = query.Where(clause.Eq{Column: clause.Column{Name: "publish_status"}, Value: *filter.PublishStatus})
	}
	return query
}

func applyArticleFilters(query *gorm.DB, filter ArticleListFilter) *gorm.DB {
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where(clause.Or(
			clause.Like{Column: clause.Column{Name: "title"}, Value: keyword},
			clause.Like{Column: clause.Column{Name: "article_code"}, Value: keyword},
			clause.Like{Column: clause.Column{Name: "description"}, Value: keyword},
		))
	}
	if filter.VisibleStatus != nil {
		query = query.Where(clause.Eq{Column: clause.Column{Name: "visible_status"}, Value: *filter.VisibleStatus})
	}
	if filter.PublishStatus != nil {
		query = query.Where(clause.Eq{Column: clause.Column{Name: "publish_status"}, Value: *filter.PublishStatus})
	}
	return query
}

func applySpeciesFilters(query *gorm.DB, filter SpeciesListFilter) *gorm.DB {
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where(clause.Or(
			clause.Like{Column: clause.Column{Name: "name"}, Value: keyword},
			clause.Like{Column: clause.Column{Name: "species_code"}, Value: keyword},
			clause.Like{Column: clause.Column{Name: "alias"}, Value: keyword},
		))
	}
	if filter.Category != "" {
		query = query.Where(clause.Eq{Column: clause.Column{Name: "category"}, Value: filter.Category})
	}
	if filter.VisibleStatus != nil {
		query = query.Where(clause.Eq{Column: clause.Column{Name: "visible_status"}, Value: *filter.VisibleStatus})
	}
	if filter.IsFeatured != nil {
		query = query.Where(clause.Eq{Column: clause.Column{Name: "is_featured"}, Value: *filter.IsFeatured})
	}
	return query
}

func applyUserFilters(query *gorm.DB, filter UserListFilter) *gorm.DB {
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where(clause.Or(
			clause.Like{Column: clause.Column{Table: "users", Name: "nickname"}, Value: keyword},
			clause.Like{Column: clause.Column{Table: "users", Name: "id"}, Value: keyword},
		))
	}
	if filter.City != "" {
		query = query.Where(clause.Eq{Column: clause.Column{Table: "Profile", Name: "city"}, Value: filter.City})
	}
	if filter.Status != nil {
		query = query.Where(clause.Eq{Column: clause.Column{Table: "users", Name: "status"}, Value: *filter.Status})
	}
	return query
}

func upsertArticleDetail(tx *gorm.DB, detail model.DiscoverContentDetail) error {
	detail.ContentType = "article"
	return tx.Model(&model.DiscoverContentDetail{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "content_type"}, {Name: "content_code"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"type_label", "summary_text", "takeaways_text", "visible_status", "updated_at",
		}),
	}).Create(&detail).Error
}

func replaceSpeciesDetailTips(tx *gorm.DB, speciesID int64, tips []string) error {
	if err := tx.Model(&model.SpeciesDetailTip{}).Where("species_id = ?", speciesID).Delete(&model.SpeciesDetailTip{}).Error; err != nil {
		return err
	}
	if len(tips) == 0 {
		return nil
	}
	rows := make([]model.SpeciesDetailTip, len(tips))
	for index, tip := range tips {
		rows[index] = model.SpeciesDetailTip{
			SpeciesID: speciesID, TipText: tip, SortOrder: int64(index + 1), VisibleStatus: 1,
		}
	}
	return tx.Model(&model.SpeciesDetailTip{}).Create(&rows).Error
}
