package model

import (
	"database/sql"
	"time"
)

type FishingSpot struct {
	ID                int64        `db:"id"`
	SpotCode          string       `db:"spot_code"`
	Name              string       `db:"name"`
	CoverImageURL     string       `db:"cover_image_url"`
	Province          string       `db:"province"`
	City              string       `db:"city"`
	District          string       `db:"district"`
	Address           string       `db:"address"`
	Latitude          float64      `db:"latitude"`
	Longitude         float64      `db:"longitude"`
	TagText           string       `db:"tag_text"`
	TagType           string       `db:"tag_type"`
	FishingIndex      int64        `db:"fishing_index"`
	Scene             string       `db:"scene"`
	SceneHint         string       `db:"scene_hint"`
	SourceType        string       `db:"source_type"`
	VisibleStatus     int64        `db:"visible_status"`
	PublishStatus     int64        `db:"publish_status"`
	PublisherUserID   int64        `db:"publisher_user_id"`
	PublisherNickname string       `db:"publisher_nickname"`
	PublishedAt       sql.NullTime `db:"published_at"`
	CreatedAt         time.Time    `db:"created_at"`
	UpdatedAt         time.Time    `db:"updated_at"`
}

type FishingSpotSpecies struct {
	SpotID      int64  `db:"spot_id"`
	SpeciesName string `db:"species_name"`
	SortOrder   int64  `db:"sort_order"`
}

type FishingSpotTip struct {
	SpotID        int64  `db:"spot_id"`
	TipText       string `db:"tip_text"`
	SortOrder     int64  `db:"sort_order"`
	VisibleStatus int64  `db:"visible_status"`
}

// SpotCorrection 映射小程序用户提交的钓点纠错记录。
type SpotCorrection struct {
	ID             int64         `db:"id"`
	CorrectionCode string        `db:"correction_code"`
	UserID         int64         `db:"user_id"`
	UserNickname   string        `db:"user_nickname"`
	SpotID         int64         `db:"spot_id"`
	SpotCode       string        `db:"spot_code"`
	SpotName       string        `db:"spot_name"`
	Content        string        `db:"content"`
	Status         string        `db:"status"`
	ReviewNote     string        `db:"review_note"`
	ReviewerUserID sql.NullInt64 `db:"reviewer_user_id"`
	ReviewedAt     sql.NullTime  `db:"reviewed_at"`
	CreatedAt      time.Time     `db:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at"`
	CurrentSpot    *FishingSpot  `db:"-"`
}

// PublicSpotApplication 描述用户提交的公共钓点申请。
type PublicSpotApplication struct {
	ID              int64
	ApplicationCode string
	UserID          int64
	UserNickname    string
	SourceSpotCode  string
	Name            string
	Address         string
	Latitude        float64
	Longitude       float64
	Description     string
	ChargeType      string
	TargetSpecies   []string
	ImageURLs       []string
	Status          string
	ReviewNote      string
	ReviewerUserID  *int64
	ReviewedAt      *time.Time
	MergedSpotID    int64
	MergedSpotCode  string
	CreatedAt       time.Time
}

// MediaAsset 描述关联到公开渔获记录的图片审核项。
type MediaAsset struct {
	ID               int64
	RecordID         int64
	RecordCode       string
	UserID           int64
	SpotName         string
	ImageURL         string
	Width            int
	Height           int
	ModerationStatus string
	ModerationNote   string
	CreatedAt        time.Time
}

type ReviewTask struct {
	TaskID       string    `db:"task_id"`
	TaskType     string    `db:"task_type"`
	ResourceID   int64     `db:"resource_id"`
	ResourceCode string    `db:"resource_code"`
	Title        string    `db:"title"`
	Summary      string    `db:"summary"`
	Status       string    `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
}

type DiscoverArticle struct {
	ID            int64        `db:"id"`
	ArticleCode   string       `db:"article_code"`
	ShortLabel    string       `db:"short_label"`
	Title         string       `db:"title"`
	Description   string       `db:"description"`
	Theme         string       `db:"theme"`
	SortOrder     int64        `db:"sort_order"`
	PublishStatus int64        `db:"publish_status"`
	VisibleStatus int64        `db:"visible_status"`
	PublishedAt   sql.NullTime `db:"published_at"`
	CreatedAt     time.Time    `db:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at"`
}

type DiscoverContentDetail struct {
	ID            int64     `db:"id"`
	ContentType   string    `db:"content_type"`
	ContentCode   string    `db:"content_code"`
	TypeLabel     string    `db:"type_label"`
	SummaryText   string    `db:"summary_text"`
	TakeawaysText string    `db:"takeaways_text"`
	VisibleStatus int64     `db:"visible_status"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// DiscoverEventBanner 描述发现页顶部活动横幅。
type DiscoverEventBanner struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	BannerCode    string     `gorm:"column:banner_code"`
	Tag           string     `gorm:"column:tag"`
	Title         string     `gorm:"column:title"`
	Description   string     `gorm:"column:description"`
	DeadlineText  string     `gorm:"column:deadline_text"`
	DeadlineAt    *time.Time `gorm:"column:deadline_at"`
	LocationName  string     `gorm:"column:location_name"`
	SortOrder     int64      `gorm:"column:sort_order"`
	VisibleStatus int64      `gorm:"column:visible_status"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (DiscoverEventBanner) TableName() string { return "discover_event_banners" }

type Species struct {
	ID            int64     `db:"id"`
	SpeciesCode   string    `db:"species_code"`
	Name          string    `db:"name"`
	Alias         string    `db:"alias"`
	Category      string    `db:"category"`
	WaterLayer    string    `db:"water_layer"`
	TagText       string    `db:"tag_text"`
	SeasonText    string    `db:"season_text"`
	BestWindow    string    `db:"best_window"`
	FishingMethod string    `db:"fishing_method"`
	BaitText      string    `db:"bait_text"`
	Description   string    `db:"description"`
	HighlightText string    `db:"highlight_text"`
	IsFeatured    int64     `db:"is_featured"`
	SortOrder     int64     `db:"sort_order"`
	VisibleStatus int64     `db:"visible_status"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type SpeciesDetailTip struct {
	ID            int64     `db:"id"`
	SpeciesID     int64     `db:"species_id"`
	TipText       string    `db:"tip_text"`
	SortOrder     int64     `db:"sort_order"`
	VisibleStatus int64     `db:"visible_status"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type UserProfile struct {
	UserID               int64        `db:"id"`
	Nickname             string       `db:"nickname"`
	AvatarURL            string       `db:"avatar_url"`
	Status               int64        `db:"status"`
	LastLoginAt          sql.NullTime `db:"last_login_at"`
	CreatedAt            time.Time    `db:"created_at"`
	UpdatedAt            time.Time    `db:"updated_at"`
	LevelText            string       `db:"level_text"`
	ProfileDesc          string       `db:"profile_desc"`
	LocationText         string       `db:"location_text"`
	City                 string       `db:"city"`
	District             string       `db:"district"`
	StreakWeeks          int64        `db:"streak_weeks"`
	PreferredSpeciesText string       `db:"preferred_species_text"`
}

// DashboardOverview 汇总运营后台首页使用的业务数量指标。
type DashboardOverview struct {
	Users                int64
	FishingSpots         int64
	PendingFishingSpots  int64
	Articles             int64
	PublishedArticles    int64
	Species              int64
	VisibleSpecies       int64
	FishingRecords       int64
	PublicFishingRecords int64
	PendingCorrections   int64
	PendingMediaAssets   int64
	PendingPublicReports int64
	PendingReviewTasks   int64
}
