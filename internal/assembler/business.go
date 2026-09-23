package assembler

import (
	"strings"

	"fishing-notes-admin-api/internal/model"
	"fishing-notes-admin-api/internal/types"
)

func FishingSpotReview(item model.FishingSpot, species []string, tips []string) types.FishingSpotReviewItem {
	return types.FishingSpotReviewItem{
		ID:                item.ID,
		SpotCode:          item.SpotCode,
		Name:              item.Name,
		CoverImageURL:     item.CoverImageURL,
		Province:          item.Province,
		City:              item.City,
		District:          item.District,
		Address:           item.Address,
		Latitude:          item.Latitude,
		Longitude:         item.Longitude,
		TagText:           item.TagText,
		TagType:           item.TagType,
		FishingIndex:      item.FishingIndex,
		Scene:             item.Scene,
		SceneHint:         item.SceneHint,
		SourceType:        item.SourceType,
		VisibleStatus:     item.VisibleStatus,
		PublishStatus:     item.PublishStatus,
		PublisherUserID:   item.PublisherUserID,
		PublisherNickname: item.PublisherNickname,
		PublishedAt:       NullTimePtr(item.PublishedAt),
		CreatedAt:         Time(item.CreatedAt),
		UpdatedAt:         Time(item.UpdatedAt),
		Species:           species,
		Tips:              tips,
	}
}

func Article(item model.DiscoverArticle, detail *model.DiscoverContentDetail) types.ArticleItem {
	resp := types.ArticleItem{
		ID:            item.ID,
		ArticleCode:   item.ArticleCode,
		ShortLabel:    item.ShortLabel,
		Title:         item.Title,
		Description:   item.Description,
		Theme:         item.Theme,
		SortOrder:     item.SortOrder,
		PublishStatus: item.PublishStatus,
		VisibleStatus: item.VisibleStatus,
		PublishedAt:   NullTimePtr(item.PublishedAt),
		CreatedAt:     Time(item.CreatedAt),
		UpdatedAt:     Time(item.UpdatedAt),
	}

	if detail != nil {
		resp.TypeLabel = detail.TypeLabel
		resp.SummaryLines = splitLines(detail.SummaryText)
		resp.TakeawayLines = splitLines(detail.TakeawaysText)
	}

	return resp
}

func Species(item model.Species, detailTips []string) types.SpeciesItem {
	return types.SpeciesItem{
		ID:            item.ID,
		SpeciesCode:   item.SpeciesCode,
		Name:          item.Name,
		Alias:         item.Alias,
		Category:      item.Category,
		WaterLayer:    item.WaterLayer,
		TagText:       item.TagText,
		SeasonText:    item.SeasonText,
		BestWindow:    item.BestWindow,
		FishingMethod: item.FishingMethod,
		BaitText:      item.BaitText,
		Description:   item.Description,
		HighlightText: item.HighlightText,
		IsFeatured:    item.IsFeatured,
		SortOrder:     item.SortOrder,
		VisibleStatus: item.VisibleStatus,
		CreatedAt:     Time(item.CreatedAt),
		UpdatedAt:     Time(item.UpdatedAt),
		DetailTips:    detailTips,
	}
}

func User(item model.UserProfile) types.UserItem {
	return types.UserItem{
		ID:                   item.UserID,
		Nickname:             item.Nickname,
		AvatarURL:            item.AvatarURL,
		Status:               item.Status,
		LastLoginAt:          NullTimePtr(item.LastLoginAt),
		CreatedAt:            Time(item.CreatedAt),
		UpdatedAt:            Time(item.UpdatedAt),
		LevelText:            item.LevelText,
		ProfileDesc:          item.ProfileDesc,
		LocationText:         item.LocationText,
		City:                 item.City,
		District:             item.District,
		StreakWeeks:          item.StreakWeeks,
		PreferredSpeciesText: item.PreferredSpeciesText,
	}
}

func splitLines(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}

	parts := strings.Split(value, "\n")
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		line := strings.TrimSpace(part)
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}
