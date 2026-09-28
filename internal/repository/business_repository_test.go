package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func openBusinessTestGORM(t *testing.T, db *sql.DB) *gorm.DB {
	t.Helper()
	gormDB, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("create GORM DB: %v", err)
	}
	return gormDB
}

func ptrInt64(value int64) *int64 {
	return &value
}

func TestDashboardOverviewPropagatesFirstCountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").WillReturnError(wantErr)

	_, err = NewBusinessRepository(openBusinessTestGORM(t, db)).DashboardOverview(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("DashboardOverview() error = %v, want %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestDashboardOverviewCountsAllMetrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	values := []int64{10, 20, 3, 8, 6, 12, 9, 100, 40, 2, 4, 1}
	for _, value := range values {
		mock.ExpectQuery("SELECT count\\(\\*\\) FROM").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(value))
	}

	result, err := NewBusinessRepository(openBusinessTestGORM(t, db)).DashboardOverview(context.Background())
	if err != nil {
		t.Fatalf("DashboardOverview() error = %v", err)
	}
	if result.Users != 10 || result.FishingSpots != 20 || result.PendingFishingSpots != 3 ||
		result.Articles != 8 || result.PublishedArticles != 6 || result.Species != 12 ||
		result.VisibleSpecies != 9 || result.FishingRecords != 100 || result.PublicFishingRecords != 40 ||
		result.PendingCorrections != 2 || result.PendingMediaAssets != 4 || result.PendingPublicReports != 1 ||
		result.PendingReviewTasks != 6 {
		t.Fatalf("DashboardOverview() = %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListSpotCorrectionsPropagatesCountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `spot_corrections`")).WillReturnError(wantErr)

	_, _, err = NewBusinessRepository(openBusinessTestGORM(t, db)).ListSpotCorrections(context.Background(), SpotCorrectionListFilter{Limit: 20})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected count error to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListSpotCorrectionsPreloadsUserNickname(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	createdAt := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `spot_corrections` WHERE status = \\?").
		WithArgs("pending").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT \\* FROM `spot_corrections` WHERE status = \\? ORDER BY created_at DESC,id DESC LIMIT \\?").
		WithArgs("pending", 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "correction_code", "user_id", "spot_id", "spot_code", "spot_name", "content", "status", "review_note", "reviewer_user_id", "reviewed_at", "created_at", "updated_at"}).
			AddRow(int64(9), "correction_9", int64(7), int64(3), "spot_3", "河湾", "入口位置不对", "pending", "", nil, nil, createdAt, createdAt))
	mock.ExpectQuery("SELECT \\* FROM `users` WHERE `users`.`id` = \\?").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nickname"}).AddRow(int64(7), "钓鱼人"))

	items, total, err := NewBusinessRepository(openBusinessTestGORM(t, db)).ListSpotCorrections(
		context.Background(), SpotCorrectionListFilter{Status: "pending", Limit: 20},
	)
	if err != nil {
		t.Fatalf("ListSpotCorrections() error = %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].UserNickname != "钓鱼人" {
		t.Fatalf("ListSpotCorrections() = total %d items %+v", total, items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestReviewSpotCorrectionPropagatesUpdateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `spot_corrections` SET").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(wantErr)
	mock.ExpectRollback()

	_, err = NewBusinessRepository(openBusinessTestGORM(t, db)).ReviewSpotCorrection(context.Background(), 10, 7, "approved", "looks good", time.Now())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected update error to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListReviewTasksPropagatesCountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `spot_corrections`").WillReturnError(wantErr)

	_, _, err = NewBusinessRepository(openBusinessTestGORM(t, db)).ListReviewTasks(context.Background(), ReviewTaskListFilter{
		TaskType: "spot_correction", Limit: 20,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected review task count error to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListReviewTasksMapsUserSpotTaskType(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `fishing_spots` WHERE source_type = \\?").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT \\* FROM `fishing_spots` WHERE source_type = \\? ORDER BY created_at ASC,id ASC LIMIT \\?").
		WithArgs("user", 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "spot_code", "name", "address", "source_type", "publish_status", "created_at"}).
			AddRow(int64(12), "user_spot_1", "河湾", "用户发布钓点", "user", int64(1), time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)))

	items, total, err := NewBusinessRepository(openBusinessTestGORM(t, db)).ListReviewTasks(context.Background(), ReviewTaskListFilter{
		TaskType: "user_spot", Limit: 20,
	})
	if err != nil {
		t.Fatalf("ListReviewTasks() error = %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].TaskType != "user_spot" || items[0].Status != "approved" {
		t.Fatalf("ListReviewTasks() = total %d items %+v", total, items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestFindSpotCorrectionByIDPropagatesError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `spot_corrections` WHERE id = ? LIMIT ?")).
		WithArgs(int64(8), 1).
		WillReturnError(wantErr)

	_, err = NewBusinessRepository(openBusinessTestGORM(t, db)).FindSpotCorrectionByID(context.Background(), 8)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected detail query error to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListMediaAssetsReturnsEmptyListForNoPendingImages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `fishing_record_images`.*`Record`.*visible_scope.*moderation_status").
		WithArgs("public", "pending").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	items, total, err := NewBusinessRepository(openBusinessTestGORM(t, db)).ListMediaAssets(
		context.Background(), MediaAssetListFilter{Status: "pending", Limit: 20},
	)
	if err != nil || items == nil || len(items) != 0 || total != 0 {
		t.Fatalf("ListMediaAssets() = items %+v, total %d, err %v", items, total, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListMediaAssetsPropagatesCountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `fishing_record_images`").WillReturnError(wantErr)
	_, _, err = NewBusinessRepository(openBusinessTestGORM(t, db)).ListMediaAssets(
		context.Background(), MediaAssetListFilter{Status: "pending", Limit: 20},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected count error to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListPublicSpotApplicationsReturnsEmptyList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `public_spot_applications` WHERE status = ?")).
		WithArgs("pending").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	items, total, err := NewBusinessRepository(openBusinessTestGORM(t, db)).ListPublicSpotApplications(
		context.Background(), PublicSpotApplicationListFilter{Status: "pending", Limit: 20},
	)
	if err != nil || items == nil || len(items) != 0 || total != 0 {
		t.Fatalf("ListPublicSpotApplications() = items %+v, total %d, err %v", items, total, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListPublicSpotApplicationsPropagatesCountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `public_spot_applications`").WillReturnError(wantErr)
	_, _, err = NewBusinessRepository(openBusinessTestGORM(t, db)).ListPublicSpotApplications(
		context.Background(), PublicSpotApplicationListFilter{Limit: 20},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected count error to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListBannersReturnsEmptyList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `discover_event_banners` WHERE visible_status = ?")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	items, total, err := NewBusinessRepository(openBusinessTestGORM(t, db)).ListBanners(
		context.Background(), BannerListFilter{VisibleStatus: ptrInt64(1), Limit: 20},
	)
	if err != nil || items == nil || len(items) != 0 || total != 0 {
		t.Fatalf("ListBanners() = items %+v, total %d, err %v", items, total, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListBannersPropagatesCountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `discover_event_banners`")).
		WillReturnError(wantErr)
	_, _, err = NewBusinessRepository(openBusinessTestGORM(t, db)).ListBanners(
		context.Background(), BannerListFilter{Limit: 20},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected count error to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestReviewMediaAssetUsesTransactionAndPendingState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	reviewedAt := time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT fishing_record_images.id, fishing_record_images.record_id,.*FROM `fishing_record_images`.*JOIN `fishing_records`.*Record.visible_scope = \\?.*moderation_status = \\?.*LIMIT \\?").
		WithArgs(int64(15), "public", "pending", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "record_id"}).AddRow(int64(15), int64(42)))
	mock.ExpectExec("UPDATE `fishing_record_images` SET").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int64(15), "pending").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	rows, err := NewBusinessRepository(openBusinessTestGORM(t, db)).ReviewMediaAsset(
		context.Background(), 15, 7, "approved", "内容正常", reviewedAt,
	)
	if err != nil || rows != 1 {
		t.Fatalf("ReviewMediaAsset() = rows %d, err %v", rows, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestReviewPublicFishingReportPropagatesUpdateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `fishing_records` SET").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(wantErr)
	mock.ExpectRollback()

	_, err = NewBusinessRepository(openBusinessTestGORM(t, db)).ReviewPublicFishingReport(context.Background(), 12, "approved")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected review update error to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestReviewPublicFishingReportReturnsExistingNoopAsSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `fishing_records` SET").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT `moderation_status` FROM `fishing_records`").
		WithArgs(int64(12), "public", 1).
		WillReturnRows(sqlmock.NewRows([]string{"moderation_status"}).AddRow("approved"))

	rows, err := NewBusinessRepository(openBusinessTestGORM(t, db)).ReviewPublicFishingReport(context.Background(), 12, "approved")
	if err != nil || rows != 1 {
		t.Fatalf("ReviewPublicFishingReport() = rows %d, err %v", rows, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListFishingSpotsReturnsNonNilEmptyList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `fishing_spots` WHERE `source_type` = \\?").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	items, total, err := NewBusinessRepository(openBusinessTestGORM(t, db)).ListFishingSpots(context.Background(), FishingSpotListFilter{
		SourceType: "user", Limit: 20,
	})
	if err != nil || items == nil || len(items) != 0 || total != 0 {
		t.Fatalf("ListFishingSpots() = items %+v, total %d, err %v", items, total, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListFishingSpotsPreloadsPublisherNickname(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `fishing_spots`").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT \\* FROM `fishing_spots` ORDER BY updated_at DESC,id DESC LIMIT \\?").
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "spot_code", "name", "cover_image_url", "province", "city", "district", "address",
			"latitude", "longitude", "tag_text", "tag_type", "fishing_index", "scene", "scene_hint",
			"source_type", "visible_status", "publish_status", "publisher_user_id", "published_at", "created_at", "updated_at",
		}).AddRow(int64(12), "spot_12", "河湾", "", "江苏", "南京", "", "江边", 31.0, 118.0,
			"", "", int64(0), "", "", "user", int64(1), int64(1), int64(7), nil, time.Now(), time.Now()))
	mock.ExpectQuery("SELECT \\* FROM `users` WHERE `users`.`id` = \\?").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nickname"}).AddRow(int64(7), "钓友七号"))

	items, total, err := NewBusinessRepository(openBusinessTestGORM(t, db)).ListFishingSpots(context.Background(), FishingSpotListFilter{Limit: 20})
	if err != nil {
		t.Fatalf("ListFishingSpots() error = %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].PublisherNickname != "钓友七号" {
		t.Fatalf("ListFishingSpots() = total %d items %+v", total, items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListUsersFiltersByProfileCityWithGORM(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`.*`Profile`.`city` = \\?").
		WithArgs("南京").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	items, total, err := NewBusinessRepository(openBusinessTestGORM(t, db)).ListUsers(context.Background(), UserListFilter{
		City: "南京", Limit: 20,
	})
	if err != nil || items == nil || len(items) != 0 || total != 0 {
		t.Fatalf("ListUsers() = items %+v, total %d, err %v", items, total, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestFindArticleByIDMapsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT \\* FROM `discover_articles` WHERE id = \\? LIMIT \\?").
		WithArgs(int64(31), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err = NewBusinessRepository(openBusinessTestGORM(t, db)).FindArticleByID(context.Background(), 31)
	if !errors.Is(err, sqlx.ErrNotFound) {
		t.Fatalf("FindArticleByID() error = %v, want not-found sentinel", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestUpdateUserStatusPropagatesGORMError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `users` SET").
		WithArgs(int64(0), sqlmock.AnyArg(), int64(42)).
		WillReturnError(wantErr)
	mock.ExpectRollback()

	err = NewBusinessRepository(openBusinessTestGORM(t, db)).UpdateUserStatus(context.Background(), 42, 0)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected update error to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestApproveFishingSpotUsesORMAndPreservesPublishedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	previousPublishTime := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `fishing_spots` WHERE id = \\? LIMIT \\? FOR UPDATE").
		WithArgs(int64(18), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "published_at"}).AddRow(18, previousPublishTime))
	mock.ExpectExec("UPDATE `fishing_spots` SET").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int64(18)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = NewBusinessRepository(openBusinessTestGORM(t, db)).ApproveFishingSpot(
		context.Background(), 18, 1, time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("ApproveFishingSpot() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
