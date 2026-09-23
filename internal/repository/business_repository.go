package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"fishing-notes-admin-api/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

type BusinessRepository struct {
	conn sqlx.SqlConn
}

func NewBusinessRepository(conn sqlx.SqlConn) *BusinessRepository {
	return &BusinessRepository{conn: conn}
}

func (r *BusinessRepository) ListSpotCorrections(ctx context.Context, filter SpotCorrectionListFilter) ([]model.SpotCorrection, int64, error) {
	where := "1 = 1"
	args := make([]any, 0, 2)
	if filter.Keyword != "" {
		where += " AND (spot_name LIKE ? OR spot_code LIKE ? OR content LIKE ?)"
		keyword := "%" + filter.Keyword + "%"
		args = append(args, keyword, keyword, keyword)
	}
	if filter.Status != "" {
		where += " AND status = ?"
		args = append(args, filter.Status)
	}
	total, err := countRows(ctx, r.conn, "SELECT COUNT(1) FROM spot_corrections WHERE "+where, args...)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.SpotCorrection{}, 0, nil
	}
	queryArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	query := `SELECT id, correction_code, user_id, spot_id, spot_code, spot_name, content, status, review_note, reviewer_user_id, reviewed_at, created_at, updated_at
FROM spot_corrections WHERE ` + where + ` ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?`
	var items []model.SpotCorrection
	if err := r.conn.QueryRowsCtx(ctx, &items, query, queryArgs...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []model.SpotCorrection{}, total, nil
		}
		return nil, 0, err
	}
	return items, total, nil
}

func (r *BusinessRepository) ReviewSpotCorrection(ctx context.Context, correctionID, reviewerUserID int64, status, reviewNote string, reviewedAt time.Time) (int64, error) {
	result, err := r.conn.ExecCtx(ctx, `UPDATE spot_corrections SET status = ?, review_note = ?, reviewer_user_id = ?, reviewed_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'pending'`, status, reviewNote, reviewerUserID, reviewedAt, correctionID)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	return rows, err
}

func (r *BusinessRepository) ListFishingSpots(ctx context.Context, filter FishingSpotListFilter) ([]model.FishingSpot, int64, error) {
	whereClause, args := buildFishingSpotWhere(filter)

	total, err := countRows(ctx, r.conn, "SELECT COUNT(1) FROM fishing_spots fs WHERE "+whereClause, args...)
	if err != nil {
		return nil, 0, err
	}

	listArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	query := `
SELECT
	fs.id, fs.spot_code, fs.name, fs.cover_image_url, fs.province, fs.city, fs.district, fs.address,
	fs.latitude, fs.longitude, fs.tag_text, fs.tag_type, fs.fishing_index, fs.scene, fs.scene_hint,
	fs.source_type, fs.visible_status, fs.publish_status, fs.publisher_user_id, fs.published_at,
	fs.created_at, fs.updated_at, COALESCE(u.nickname, '') AS publisher_nickname
FROM fishing_spots fs
LEFT JOIN users u ON u.id = fs.publisher_user_id
WHERE ` + whereClause + `
ORDER BY fs.updated_at DESC, fs.id DESC
LIMIT ? OFFSET ?`

	var items []model.FishingSpot
	if err := r.conn.QueryRowsCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *BusinessRepository) FindFishingSpotByID(ctx context.Context, spotID int64) (*model.FishingSpot, error) {
	query := `
SELECT
	fs.id, fs.spot_code, fs.name, fs.cover_image_url, fs.province, fs.city, fs.district, fs.address,
	fs.latitude, fs.longitude, fs.tag_text, fs.tag_type, fs.fishing_index, fs.scene, fs.scene_hint,
	fs.source_type, fs.visible_status, fs.publish_status, fs.publisher_user_id, fs.published_at,
	fs.created_at, fs.updated_at, COALESCE(u.nickname, '') AS publisher_nickname
FROM fishing_spots fs
LEFT JOIN users u ON u.id = fs.publisher_user_id
WHERE fs.id = ?
LIMIT 1`

	var item model.FishingSpot
	if err := r.conn.QueryRowCtx(ctx, &item, query, spotID); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *BusinessRepository) ListFishingSpotSpeciesBySpotIDs(ctx context.Context, spotIDs []int64) ([]model.FishingSpotSpecies, error) {
	if len(spotIDs) == 0 {
		return []model.FishingSpotSpecies{}, nil
	}

	query, args := inQuery(`
SELECT spot_id, species_name, sort_order
FROM fishing_spot_species
WHERE spot_id IN (%s)
ORDER BY spot_id ASC, sort_order ASC, id ASC`, spotIDs)

	var items []model.FishingSpotSpecies
	if err := r.conn.QueryRowsCtx(ctx, &items, query, args...); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *BusinessRepository) ListFishingSpotTipsBySpotIDs(ctx context.Context, spotIDs []int64) ([]model.FishingSpotTip, error) {
	if len(spotIDs) == 0 {
		return []model.FishingSpotTip{}, nil
	}

	query, args := inQuery(`
SELECT spot_id, tip_text, sort_order, visible_status
FROM fishing_spot_tips
WHERE spot_id IN (%s) AND visible_status = 1
ORDER BY spot_id ASC, sort_order ASC, id ASC`, spotIDs)

	var items []model.FishingSpotTip
	if err := r.conn.QueryRowsCtx(ctx, &items, query, args...); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *BusinessRepository) ApproveFishingSpot(ctx context.Context, spotID int64, publishStatus int64, publishedAt time.Time) error {
	const query = `
UPDATE fishing_spots
SET visible_status = 1,
    publish_status = ?,
    published_at = COALESCE(published_at, ?),
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?`
	_, err := r.conn.ExecCtx(ctx, query, publishStatus, publishedAt, spotID)
	return err
}

func (r *BusinessRepository) RejectFishingSpot(ctx context.Context, spotID int64, publishStatus int64) error {
	const query = `
UPDATE fishing_spots
SET visible_status = 0,
    publish_status = ?,
    published_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?`
	_, err := r.conn.ExecCtx(ctx, query, publishStatus, spotID)
	return err
}

func (r *BusinessRepository) ListArticles(ctx context.Context, filter ArticleListFilter) ([]model.DiscoverArticle, int64, error) {
	whereClause, args := buildArticleWhere(filter)

	total, err := countRows(ctx, r.conn, "SELECT COUNT(1) FROM discover_articles da WHERE "+whereClause, args...)
	if err != nil {
		return nil, 0, err
	}

	listArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	query := `
SELECT id, article_code, short_label, title, description, theme, sort_order, publish_status,
       visible_status, published_at, created_at, updated_at
FROM discover_articles da
WHERE ` + whereClause + `
ORDER BY sort_order ASC, id DESC
LIMIT ? OFFSET ?`

	var items []model.DiscoverArticle
	if err := r.conn.QueryRowsCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *BusinessRepository) FindArticleByID(ctx context.Context, articleID int64) (*model.DiscoverArticle, error) {
	const query = `
SELECT id, article_code, short_label, title, description, theme, sort_order, publish_status,
       visible_status, published_at, created_at, updated_at
FROM discover_articles
WHERE id = ?
LIMIT 1`

	var item model.DiscoverArticle
	if err := r.conn.QueryRowCtx(ctx, &item, query, articleID); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *BusinessRepository) FindArticleDetailByCode(ctx context.Context, articleCode string) (*model.DiscoverContentDetail, error) {
	const query = `
SELECT id, content_type, content_code, type_label, summary_text, takeaways_text, visible_status, created_at, updated_at
FROM discover_content_details
WHERE content_type = 'article' AND content_code = ?
LIMIT 1`

	var item model.DiscoverContentDetail
	if err := r.conn.QueryRowCtx(ctx, &item, query, articleCode); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *BusinessRepository) FindArticleDetailsByCodes(ctx context.Context, articleCodes []string) ([]model.DiscoverContentDetail, error) {
	if len(articleCodes) == 0 {
		return []model.DiscoverContentDetail{}, nil
	}

	query, args := inStringQuery(`
SELECT id, content_type, content_code, type_label, summary_text, takeaways_text, visible_status, created_at, updated_at
FROM discover_content_details
WHERE content_type = 'article' AND content_code IN (%s)`, articleCodes)

	var items []model.DiscoverContentDetail
	if err := r.conn.QueryRowsCtx(ctx, &items, query, args...); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *BusinessRepository) CreateArticle(ctx context.Context, item model.DiscoverArticle, detail model.DiscoverContentDetail) (int64, error) {
	var articleID int64
	err := r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		result, err := conn.ExecCtx(ctx, `
INSERT INTO discover_articles (article_code, short_label, title, description, theme, sort_order, publish_status, visible_status, published_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ArticleCode, item.ShortLabel, item.Title, item.Description, item.Theme, item.SortOrder, item.PublishStatus, item.VisibleStatus, nullableTime(item.PublishedAt))
		if err != nil {
			return err
		}

		insertID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		articleID = insertID

		_, err = conn.ExecCtx(ctx, `
INSERT INTO discover_content_details (content_type, content_code, type_label, summary_text, takeaways_text, visible_status)
VALUES ('article', ?, ?, ?, ?, ?)`,
			detail.ContentCode, detail.TypeLabel, detail.SummaryText, detail.TakeawaysText, detail.VisibleStatus)
		return err
	})
	if err != nil {
		return 0, err
	}

	return articleID, nil
}

func (r *BusinessRepository) UpdateArticle(ctx context.Context, articleID int64, item model.DiscoverArticle, detail model.DiscoverContentDetail) error {
	return r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		if _, err := conn.ExecCtx(ctx, `
UPDATE discover_articles
SET short_label = ?, title = ?, description = ?, theme = ?, sort_order = ?, visible_status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?`,
			item.ShortLabel, item.Title, item.Description, item.Theme, item.SortOrder, item.VisibleStatus, articleID); err != nil {
			return err
		}

		return upsertArticleDetail(ctx, conn, detail)
	})
}

func (r *BusinessRepository) PublishArticle(ctx context.Context, articleID int64, visibleStatus int64, publishedAt time.Time) error {
	const query = `
UPDATE discover_articles
SET publish_status = 1,
    visible_status = ?,
    published_at = COALESCE(published_at, ?),
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?`
	_, err := r.conn.ExecCtx(ctx, query, visibleStatus, publishedAt, articleID)
	return err
}

func (r *BusinessRepository) OfflineArticle(ctx context.Context, articleID int64) error {
	const query = `
UPDATE discover_articles
SET publish_status = 2,
    visible_status = 0,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?`
	_, err := r.conn.ExecCtx(ctx, query, articleID)
	return err
}

func (r *BusinessRepository) ListSpecies(ctx context.Context, filter SpeciesListFilter) ([]model.Species, int64, error) {
	whereClause, args := buildSpeciesWhere(filter)

	total, err := countRows(ctx, r.conn, "SELECT COUNT(1) FROM species s WHERE "+whereClause, args...)
	if err != nil {
		return nil, 0, err
	}

	listArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	query := `
SELECT id, species_code, name, alias, category, water_layer, tag_text, season_text, best_window,
       fishing_method, bait_text, description, highlight_text, is_featured, sort_order, visible_status,
       created_at, updated_at
FROM species s
WHERE ` + whereClause + `
ORDER BY sort_order ASC, id DESC
LIMIT ? OFFSET ?`

	var items []model.Species
	if err := r.conn.QueryRowsCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *BusinessRepository) FindSpeciesByID(ctx context.Context, speciesID int64) (*model.Species, error) {
	const query = `
SELECT id, species_code, name, alias, category, water_layer, tag_text, season_text, best_window,
       fishing_method, bait_text, description, highlight_text, is_featured, sort_order, visible_status,
       created_at, updated_at
FROM species
WHERE id = ?
LIMIT 1`

	var item model.Species
	if err := r.conn.QueryRowCtx(ctx, &item, query, speciesID); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *BusinessRepository) ListSpeciesDetailTipsBySpeciesIDs(ctx context.Context, speciesIDs []int64) ([]model.SpeciesDetailTip, error) {
	if len(speciesIDs) == 0 {
		return []model.SpeciesDetailTip{}, nil
	}

	query, args := inQuery(`
SELECT id, species_id, tip_text, sort_order, visible_status, created_at, updated_at
FROM species_detail_tips
WHERE species_id IN (%s) AND visible_status = 1
ORDER BY species_id ASC, sort_order ASC, id ASC`, speciesIDs)

	var items []model.SpeciesDetailTip
	if err := r.conn.QueryRowsCtx(ctx, &items, query, args...); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *BusinessRepository) CreateSpecies(ctx context.Context, item model.Species, tips []string) (int64, error) {
	var speciesID int64
	err := r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		result, err := conn.ExecCtx(ctx, `
INSERT INTO species (species_code, name, alias, category, water_layer, tag_text, season_text, best_window,
                     fishing_method, bait_text, description, highlight_text, is_featured, sort_order, visible_status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.SpeciesCode, item.Name, item.Alias, item.Category, item.WaterLayer, item.TagText, item.SeasonText,
			item.BestWindow, item.FishingMethod, item.BaitText, item.Description, item.HighlightText, item.IsFeatured,
			item.SortOrder, item.VisibleStatus)
		if err != nil {
			return err
		}

		insertID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		speciesID = insertID

		return replaceSpeciesDetailTips(ctx, conn, speciesID, tips)
	})
	if err != nil {
		return 0, err
	}

	return speciesID, nil
}

func (r *BusinessRepository) UpdateSpecies(ctx context.Context, speciesID int64, item model.Species, tips []string) error {
	return r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		if _, err := conn.ExecCtx(ctx, `
UPDATE species
SET name = ?, alias = ?, category = ?, water_layer = ?, tag_text = ?, season_text = ?, best_window = ?,
    fishing_method = ?, bait_text = ?, description = ?, highlight_text = ?, is_featured = ?, sort_order = ?,
    visible_status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?`,
			item.Name, item.Alias, item.Category, item.WaterLayer, item.TagText, item.SeasonText, item.BestWindow,
			item.FishingMethod, item.BaitText, item.Description, item.HighlightText, item.IsFeatured, item.SortOrder,
			item.VisibleStatus, speciesID); err != nil {
			return err
		}

		return replaceSpeciesDetailTips(ctx, conn, speciesID, tips)
	})
}

func (r *BusinessRepository) UpdateSpeciesVisibility(ctx context.Context, speciesID int64, visibleStatus int64) error {
	const query = `UPDATE species SET visible_status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.conn.ExecCtx(ctx, query, visibleStatus, speciesID)
	return err
}

func (r *BusinessRepository) ListUsers(ctx context.Context, filter UserListFilter) ([]model.UserProfile, int64, error) {
	whereClause, args := buildUserWhere(filter)

	total, err := countRows(ctx, r.conn, `
SELECT COUNT(1)
FROM users u
LEFT JOIN user_profiles up ON up.user_id = u.id
WHERE `+whereClause, args...)
	if err != nil {
		return nil, 0, err
	}

	listArgs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	query := `
SELECT
	u.id, u.nickname, u.avatar AS avatar_url, u.status, u.last_login_at, u.created_at, u.updated_at,
	COALESCE(up.level_text, '') AS level_text,
	COALESCE(up.profile_desc, '') AS profile_desc,
	COALESCE(up.location_text, '') AS location_text,
	COALESCE(up.city, '') AS city,
	COALESCE(up.district, '') AS district,
	COALESCE(up.streak_weeks, 0) AS streak_weeks,
	COALESCE(up.preferred_species_text, '') AS preferred_species_text
FROM users u
LEFT JOIN user_profiles up ON up.user_id = u.id
WHERE ` + whereClause + `
ORDER BY u.updated_at DESC, u.id DESC
LIMIT ? OFFSET ?`

	var items []model.UserProfile
	if err := r.conn.QueryRowsCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *BusinessRepository) FindUserByID(ctx context.Context, userID int64) (*model.UserProfile, error) {
	query := `
SELECT
	u.id, u.nickname, u.avatar AS avatar_url, u.status, u.last_login_at, u.created_at, u.updated_at,
	COALESCE(up.level_text, '') AS level_text,
	COALESCE(up.profile_desc, '') AS profile_desc,
	COALESCE(up.location_text, '') AS location_text,
	COALESCE(up.city, '') AS city,
	COALESCE(up.district, '') AS district,
	COALESCE(up.streak_weeks, 0) AS streak_weeks,
	COALESCE(up.preferred_species_text, '') AS preferred_species_text
FROM users u
LEFT JOIN user_profiles up ON up.user_id = u.id
WHERE u.id = ?
LIMIT 1`

	var item model.UserProfile
	if err := r.conn.QueryRowCtx(ctx, &item, query, userID); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *BusinessRepository) UpdateUserStatus(ctx context.Context, userID int64, status int64) error {
	const query = `UPDATE users SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.conn.ExecCtx(ctx, query, status, userID)
	return err
}

func buildFishingSpotWhere(filter FishingSpotListFilter) (string, []any) {
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 4)

	if filter.Keyword != "" {
		conditions = append(conditions, "(fs.name LIKE ? OR fs.spot_code LIKE ? OR fs.address LIKE ?)")
		keyword := "%" + filter.Keyword + "%"
		args = append(args, keyword, keyword, keyword)
	}
	if filter.SourceType != "" {
		conditions = append(conditions, "fs.source_type = ?")
		args = append(args, filter.SourceType)
	}
	if filter.VisibleStatus != nil {
		conditions = append(conditions, "fs.visible_status = ?")
		args = append(args, *filter.VisibleStatus)
	}
	if filter.PublishStatus != nil {
		conditions = append(conditions, "fs.publish_status = ?")
		args = append(args, *filter.PublishStatus)
	}

	return strings.Join(conditions, " AND "), args
}

func buildArticleWhere(filter ArticleListFilter) (string, []any) {
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 3)

	if filter.Keyword != "" {
		conditions = append(conditions, "(da.title LIKE ? OR da.article_code LIKE ? OR da.description LIKE ?)")
		keyword := "%" + filter.Keyword + "%"
		args = append(args, keyword, keyword, keyword)
	}
	if filter.VisibleStatus != nil {
		conditions = append(conditions, "da.visible_status = ?")
		args = append(args, *filter.VisibleStatus)
	}
	if filter.PublishStatus != nil {
		conditions = append(conditions, "da.publish_status = ?")
		args = append(args, *filter.PublishStatus)
	}

	return strings.Join(conditions, " AND "), args
}

func buildSpeciesWhere(filter SpeciesListFilter) (string, []any) {
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 4)

	if filter.Keyword != "" {
		conditions = append(conditions, "(s.name LIKE ? OR s.species_code LIKE ? OR s.alias LIKE ?)")
		keyword := "%" + filter.Keyword + "%"
		args = append(args, keyword, keyword, keyword)
	}
	if filter.Category != "" {
		conditions = append(conditions, "s.category = ?")
		args = append(args, filter.Category)
	}
	if filter.VisibleStatus != nil {
		conditions = append(conditions, "s.visible_status = ?")
		args = append(args, *filter.VisibleStatus)
	}
	if filter.IsFeatured != nil {
		conditions = append(conditions, "s.is_featured = ?")
		args = append(args, *filter.IsFeatured)
	}

	return strings.Join(conditions, " AND "), args
}

func buildUserWhere(filter UserListFilter) (string, []any) {
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 4)

	if filter.Keyword != "" {
		conditions = append(conditions, "(u.nickname LIKE ? OR CAST(u.id AS CHAR) LIKE ?)")
		keyword := "%" + filter.Keyword + "%"
		args = append(args, keyword, keyword)
	}
	if filter.City != "" {
		conditions = append(conditions, "up.city = ?")
		args = append(args, filter.City)
	}
	if filter.Status != nil {
		conditions = append(conditions, "u.status = ?")
		args = append(args, *filter.Status)
	}

	return strings.Join(conditions, " AND "), args
}

func countRows(ctx context.Context, conn sqlx.SqlConn, query string, args ...any) (int64, error) {
	var total int64
	if err := conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}

	return total, nil
}

func inQuery(pattern string, ids []int64) (string, []any) {
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	return fmt.Sprintf(pattern, strings.Join(placeholders, ", ")), args
}

func inStringQuery(pattern string, values []string) (string, []any) {
	placeholders := make([]string, 0, len(values))
	args := make([]any, 0, len(values))
	for _, value := range values {
		placeholders = append(placeholders, "?")
		args = append(args, value)
	}

	return fmt.Sprintf(pattern, strings.Join(placeholders, ", ")), args
}

func upsertArticleDetail(ctx context.Context, conn sqlx.SqlConn, detail model.DiscoverContentDetail) error {
	const query = `
INSERT INTO discover_content_details (content_type, content_code, type_label, summary_text, takeaways_text, visible_status)
VALUES ('article', ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	type_label = VALUES(type_label),
	summary_text = VALUES(summary_text),
	takeaways_text = VALUES(takeaways_text),
	visible_status = VALUES(visible_status),
	updated_at = CURRENT_TIMESTAMP`

	_, err := conn.ExecCtx(ctx, query, detail.ContentCode, detail.TypeLabel, detail.SummaryText, detail.TakeawaysText, detail.VisibleStatus)
	return err
}

func replaceSpeciesDetailTips(ctx context.Context, conn sqlx.SqlConn, speciesID int64, tips []string) error {
	if _, err := conn.ExecCtx(ctx, "DELETE FROM species_detail_tips WHERE species_id = ?", speciesID); err != nil {
		return err
	}

	for index, tip := range tips {
		if _, err := conn.ExecCtx(ctx, `
INSERT INTO species_detail_tips (species_id, tip_text, sort_order, visible_status)
VALUES (?, ?, ?, 1)`, speciesID, tip, index+1); err != nil {
			return err
		}
	}

	return nil
}

func nullableTime(value sql.NullTime) any {
	if value.Valid {
		return value.Time
	}

	return nil
}
