package types

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
	ReviewNote string `json:"reviewNote"`
}

type SpotCorrectionItem struct {
	ID             int64   `json:"id"`
	CorrectionCode string  `json:"correctionCode"`
	SpotCode       string  `json:"spotCode"`
	SpotName       string  `json:"spotName"`
	Content        string  `json:"content"`
	Status         string  `json:"status"`
	ReviewNote     string  `json:"reviewNote"`
	ReviewerUserID *int64  `json:"reviewerUserId,omitempty"`
	ReviewedAt     *string `json:"reviewedAt,omitempty"`
	CreatedAt      string  `json:"createdAt"`
}

type ListSpotCorrectionsResponse struct {
	List     []SpotCorrectionItem `json:"list"`
	Total    int64                `json:"total"`
	Page     int64                `json:"page"`
	PageSize int64                `json:"pageSize"`
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
