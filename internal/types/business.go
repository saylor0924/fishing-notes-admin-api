package types

import "encoding/json"

type DashboardOverviewResponse struct {
	Users                int64 `json:"users"`
	FishingSpots         int64 `json:"fishingSpots"`
	PendingFishingSpots  int64 `json:"pendingFishingSpots"`
	Articles             int64 `json:"articles"`
	PublishedArticles    int64 `json:"publishedArticles"`
	Species              int64 `json:"species"`
	VisibleSpecies       int64 `json:"visibleSpecies"`
	FishingRecords       int64 `json:"fishingRecords"`
	PublicFishingRecords int64 `json:"publicFishingRecords"`
	PendingCorrections   int64 `json:"pendingCorrections"`
	PendingMediaAssets   int64 `json:"pendingMediaAssets"`
	PendingPublicReports int64 `json:"pendingPublicReports"`
	PendingReviewTasks   int64 `json:"pendingReviewTasks"`
}

type ListAuditLogsRequest struct {
	Page         int64  `form:"page,optional"`
	PageSize     int64  `form:"pageSize,optional"`
	ResourceType string `form:"resourceType,optional"`
	ResourceID   int64  `form:"resourceId,optional"`
	Action       string `form:"action,optional"`
}

type AuditLogItem struct {
	ID           int64           `json:"id"`
	ActorUserID  *int64          `json:"actorUserId,omitempty"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resourceType"`
	ResourceID   *int64          `json:"resourceId,omitempty"`
	Detail       json.RawMessage `json:"detail"`
	CreatedAt    string          `json:"createdAt"`
}

type ListAuditLogsResponse struct {
	List     []AuditLogItem `json:"list"`
	Total    int64          `json:"total"`
	Page     int64          `json:"page"`
	PageSize int64          `json:"pageSize"`
}

type ListFishingSpotReviewsRequest struct {
	Page          int64  `form:"page,optional"`
	PageSize      int64  `form:"pageSize,optional"`
	Keyword       string `form:"keyword,optional"`
	SourceType    string `form:"sourceType,optional"`
	VisibleStatus string `form:"visibleStatus,optional"`
	PublishStatus string `form:"publishStatus,optional"`
}

type SpotPathRequest struct {
	SpotID int64 `path:"spotId"`
}

type FishingSpotReviewItem struct {
	ID                int64    `json:"id"`
	SpotCode          string   `json:"spotCode"`
	Name              string   `json:"name"`
	CoverImageURL     string   `json:"coverImageUrl"`
	Province          string   `json:"province"`
	City              string   `json:"city"`
	District          string   `json:"district"`
	Address           string   `json:"address"`
	Latitude          float64  `json:"latitude"`
	Longitude         float64  `json:"longitude"`
	TagText           string   `json:"tagText"`
	TagType           string   `json:"tagType"`
	FishingIndex      int64    `json:"fishingIndex"`
	Scene             string   `json:"scene"`
	SceneHint         string   `json:"sceneHint"`
	SourceType        string   `json:"sourceType"`
	VisibleStatus     int64    `json:"visibleStatus"`
	PublishStatus     int64    `json:"publishStatus"`
	PublisherUserID   int64    `json:"publisherUserId"`
	PublisherNickname string   `json:"publisherNickname"`
	PublishedAt       *string  `json:"publishedAt,omitempty"`
	CreatedAt         string   `json:"createdAt"`
	UpdatedAt         string   `json:"updatedAt"`
	Species           []string `json:"species"`
	Tips              []string `json:"tips"`
}

type ListFishingSpotReviewsResponse struct {
	List     []FishingSpotReviewItem `json:"list"`
	Total    int64                   `json:"total"`
	Page     int64                   `json:"page"`
	PageSize int64                   `json:"pageSize"`
}

type ListPublicSpotApplicationsRequest struct {
	Page     int64  `form:"page,optional"`
	PageSize int64  `form:"pageSize,optional"`
	Keyword  string `form:"keyword,optional"`
	Status   string `form:"status,optional"`
}

type PublicSpotApplicationPathRequest struct {
	ApplicationID int64 `path:"applicationId"`
}

type ReviewPublicSpotApplicationRequest struct {
	ReviewNote  string `json:"reviewNote"`
	MergeSpotID *int64 `json:"mergeSpotId,optional,omitempty"`
}

type PublicSpotApplicationItem struct {
	ID              int64    `json:"id"`
	ApplicationCode string   `json:"applicationCode"`
	UserID          int64    `json:"userId"`
	UserNickname    string   `json:"userNickname"`
	SourceSpotCode  string   `json:"sourceSpotCode"`
	Name            string   `json:"name"`
	Address         string   `json:"address"`
	Latitude        float64  `json:"latitude"`
	Longitude       float64  `json:"longitude"`
	Description     string   `json:"description"`
	ChargeType      string   `json:"chargeType"`
	TargetSpecies   []string `json:"targetSpecies"`
	ImageURLs       []string `json:"imageUrls"`
	Status          string   `json:"status"`
	ReviewNote      string   `json:"reviewNote"`
	ReviewerUserID  *int64   `json:"reviewerUserId,omitempty"`
	ReviewedAt      *string  `json:"reviewedAt,omitempty"`
	MergedSpotID    int64    `json:"mergedSpotId"`
	MergedSpotCode  string   `json:"mergedSpotCode"`
	CreatedAt       string   `json:"createdAt"`
}

type ListPublicSpotApplicationsResponse struct {
	List     []PublicSpotApplicationItem `json:"list"`
	Total    int64                       `json:"total"`
	Page     int64                       `json:"page"`
	PageSize int64                       `json:"pageSize"`
}

type ApproveFishingSpotRequest struct {
	PublishStatus *int64 `json:"publishStatus,omitempty"`
}

type RejectFishingSpotRequest struct {
	PublishStatus *int64 `json:"publishStatus,omitempty"`
}

type ListSpotCorrectionsRequest struct {
	Page     int64  `form:"page,optional"`
	PageSize int64  `form:"pageSize,optional"`
	Keyword  string `form:"keyword,optional"`
	Status   string `form:"status,optional"`
}

type SpotCorrectionPathRequest struct {
	CorrectionID int64 `path:"correctionId"`
}

type ReviewSpotCorrectionRequest struct {
	ReviewNote string   `json:"reviewNote"`
	Fields     []string `json:"fields,optional,omitempty"`
}

type ListMediaAssetsRequest struct {
	Page     int64  `form:"page,optional"`
	PageSize int64  `form:"pageSize,optional"`
	Keyword  string `form:"keyword,optional"`
	Status   string `form:"status,optional"`
}

type MediaAssetPathRequest struct {
	MediaID int64 `path:"mediaId"`
}

type ReviewMediaAssetRequest struct {
	ReviewNote string `json:"reviewNote"`
}

type MediaAssetItem struct {
	ID               int64  `json:"id"`
	RecordCode       string `json:"recordCode"`
	UserID           int64  `json:"userId"`
	SpotName         string `json:"spotName"`
	ImageURL         string `json:"imageUrl"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	ModerationStatus string `json:"moderationStatus"`
	ModerationNote   string `json:"moderationNote"`
	CreatedAt        string `json:"createdAt"`
}

type ListMediaAssetsResponse struct {
	List     []MediaAssetItem `json:"list"`
	Total    int64            `json:"total"`
	Page     int64            `json:"page"`
	PageSize int64            `json:"pageSize"`
}

type SpotCorrectionFieldChange struct {
	Field    string `json:"field"`
	Label    string `json:"label"`
	Current  string `json:"current"`
	Proposed string `json:"proposed"`
}

type SpotCorrectionDiff struct {
	Structured bool                        `json:"structured"`
	Message    string                      `json:"message"`
	Fields     []SpotCorrectionFieldChange `json:"fields"`
}

type SpotCorrectionItem struct {
	ID             int64              `json:"id"`
	CorrectionCode string             `json:"correctionCode"`
	UserID         int64              `json:"userId"`
	UserNickname   string             `json:"userNickname"`
	SpotCode       string             `json:"spotCode"`
	SpotName       string             `json:"spotName"`
	Content        string             `json:"content"`
	Status         string             `json:"status"`
	ReviewNote     string             `json:"reviewNote"`
	ReviewerUserID *int64             `json:"reviewerUserId,omitempty"`
	ReviewedAt     *string            `json:"reviewedAt,omitempty"`
	CreatedAt      string             `json:"createdAt"`
	Diff           SpotCorrectionDiff `json:"diff"`
}

type ListSpotCorrectionsResponse struct {
	List     []SpotCorrectionItem `json:"list"`
	Total    int64                `json:"total"`
	Page     int64                `json:"page"`
	PageSize int64                `json:"pageSize"`
}

type ListReviewTasksRequest struct {
	Page     int64  `form:"page,optional"`
	PageSize int64  `form:"pageSize,optional"`
	Keyword  string `form:"keyword,optional"`
	Status   string `form:"status,optional"`
	TaskType string `form:"taskType,optional"`
}

type PublicFishingReportPathRequest struct {
	RecordID int64 `path:"recordId"`
}

type ReviewPublicFishingReportRequest struct {
	ReviewNote string `json:"reviewNote"`
}

type ReviewTaskItem struct {
	TaskID       string `json:"taskId"`
	TaskType     string `json:"taskType"`
	ResourceID   int64  `json:"resourceId"`
	ResourceCode string `json:"resourceCode"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	Status       string `json:"status"`
	CreatedAt    string `json:"createdAt"`
}

type ListReviewTasksResponse struct {
	List     []ReviewTaskItem `json:"list"`
	Total    int64            `json:"total"`
	Page     int64            `json:"page"`
	PageSize int64            `json:"pageSize"`
}

type ArticlePathRequest struct {
	ArticleID int64 `path:"articleId"`
}

type ListArticlesRequest struct {
	Page          int64  `form:"page,optional"`
	PageSize      int64  `form:"pageSize,optional"`
	Keyword       string `form:"keyword,optional"`
	PublishStatus string `form:"publishStatus,optional"`
	VisibleStatus string `form:"visibleStatus,optional"`
}

type ArticleUpsertRequest struct {
	ShortLabel    string   `json:"shortLabel"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Theme         string   `json:"theme"`
	SortOrder     int64    `json:"sortOrder"`
	VisibleStatus int64    `json:"visibleStatus"`
	SummaryLines  []string `json:"summaryLines"`
	TakeawayLines []string `json:"takeawayLines"`
	TypeLabel     string   `json:"typeLabel"`
}

type ArticlePublishRequest struct {
	VisibleStatus *int64 `json:"visibleStatus,omitempty"`
}

type ArticleItem struct {
	ID            int64    `json:"id"`
	ArticleCode   string   `json:"articleCode"`
	ShortLabel    string   `json:"shortLabel"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Theme         string   `json:"theme"`
	SortOrder     int64    `json:"sortOrder"`
	PublishStatus int64    `json:"publishStatus"`
	VisibleStatus int64    `json:"visibleStatus"`
	PublishedAt   *string  `json:"publishedAt,omitempty"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
	TypeLabel     string   `json:"typeLabel"`
	SummaryLines  []string `json:"summaryLines"`
	TakeawayLines []string `json:"takeawayLines"`
}

type ListArticlesResponse struct {
	List     []ArticleItem `json:"list"`
	Total    int64         `json:"total"`
	Page     int64         `json:"page"`
	PageSize int64         `json:"pageSize"`
}

type BannerPathRequest struct {
	BannerID int64 `path:"bannerId"`
}

type ListBannersRequest struct {
	Page          int64  `form:"page,optional"`
	PageSize      int64  `form:"pageSize,optional"`
	Keyword       string `form:"keyword,optional"`
	VisibleStatus string `form:"visibleStatus,optional"`
}

type BannerUpsertRequest struct {
	Tag           string `json:"tag"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	DeadlineText  string `json:"deadlineText"`
	DeadlineAt    string `json:"deadlineAt,optional"`
	LocationName  string `json:"locationName"`
	SortOrder     int64  `json:"sortOrder"`
	VisibleStatus int64  `json:"visibleStatus"`
}

type UpdateBannerVisibilityRequest struct {
	VisibleStatus int64 `json:"visibleStatus"`
}

type BannerItem struct {
	ID            int64  `json:"id"`
	BannerCode    string `json:"bannerCode"`
	Tag           string `json:"tag"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	DeadlineText  string `json:"deadlineText"`
	DeadlineAt    string `json:"deadlineAt,omitempty"`
	LocationName  string `json:"locationName"`
	SortOrder     int64  `json:"sortOrder"`
	VisibleStatus int64  `json:"visibleStatus"`
	PublishedAt   string `json:"publishedAt,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type ListBannersResponse struct {
	List     []BannerItem `json:"list"`
	Total    int64        `json:"total"`
	Page     int64        `json:"page"`
	PageSize int64        `json:"pageSize"`
}

type SpeciesPathRequest struct {
	SpeciesID int64 `path:"speciesId"`
}

type ListSpeciesRequest struct {
	Page          int64  `form:"page,optional"`
	PageSize      int64  `form:"pageSize,optional"`
	Keyword       string `form:"keyword,optional"`
	Category      string `form:"category,optional"`
	VisibleStatus string `form:"visibleStatus,optional"`
	IsFeatured    string `form:"isFeatured,optional"`
}

type SpeciesUpsertRequest struct {
	Name          string   `json:"name"`
	Alias         string   `json:"alias"`
	Category      string   `json:"category"`
	WaterLayer    string   `json:"waterLayer"`
	TagText       string   `json:"tagText"`
	SeasonText    string   `json:"seasonText"`
	BestWindow    string   `json:"bestWindow"`
	FishingMethod string   `json:"fishingMethod"`
	BaitText      string   `json:"baitText"`
	Description   string   `json:"description"`
	HighlightText string   `json:"highlightText"`
	IsFeatured    int64    `json:"isFeatured"`
	SortOrder     int64    `json:"sortOrder"`
	VisibleStatus int64    `json:"visibleStatus"`
	DetailTips    []string `json:"detailTips"`
}

type UpdateSpeciesVisibilityRequest struct {
	VisibleStatus int64 `json:"visibleStatus"`
}

type SpeciesItem struct {
	ID            int64    `json:"id"`
	SpeciesCode   string   `json:"speciesCode"`
	Name          string   `json:"name"`
	Alias         string   `json:"alias"`
	Category      string   `json:"category"`
	WaterLayer    string   `json:"waterLayer"`
	TagText       string   `json:"tagText"`
	SeasonText    string   `json:"seasonText"`
	BestWindow    string   `json:"bestWindow"`
	FishingMethod string   `json:"fishingMethod"`
	BaitText      string   `json:"baitText"`
	Description   string   `json:"description"`
	HighlightText string   `json:"highlightText"`
	IsFeatured    int64    `json:"isFeatured"`
	SortOrder     int64    `json:"sortOrder"`
	VisibleStatus int64    `json:"visibleStatus"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
	DetailTips    []string `json:"detailTips"`
}

type ListSpeciesResponse struct {
	List     []SpeciesItem `json:"list"`
	Total    int64         `json:"total"`
	Page     int64         `json:"page"`
	PageSize int64         `json:"pageSize"`
}

type UserPathRequest struct {
	UserID int64 `path:"userId"`
}

type ListUsersRequest struct {
	Page     int64  `form:"page,optional"`
	PageSize int64  `form:"pageSize,optional"`
	Keyword  string `form:"keyword,optional"`
	Status   string `form:"status,optional"`
	City     string `form:"city,optional"`
}

type UpdateUserStatusRequest struct {
	Status int64 `json:"status"`
}

type UserItem struct {
	ID                   int64   `json:"id"`
	Nickname             string  `json:"nickname"`
	AvatarURL            string  `json:"avatarUrl"`
	Status               int64   `json:"status"`
	LastLoginAt          *string `json:"lastLoginAt,omitempty"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
	LevelText            string  `json:"levelText"`
	ProfileDesc          string  `json:"profileDesc"`
	LocationText         string  `json:"locationText"`
	City                 string  `json:"city"`
	District             string  `json:"district"`
	StreakWeeks          int64   `json:"streakWeeks"`
	PreferredSpeciesText string  `json:"preferredSpeciesText"`
}

type ListUsersResponse struct {
	List     []UserItem `json:"list"`
	Total    int64      `json:"total"`
	Page     int64      `json:"page"`
	PageSize int64      `json:"pageSize"`
}
