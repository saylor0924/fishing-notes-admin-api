package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestAuditRepositoryAppendPropagatesInsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("audit table unavailable")
	mock.ExpectBegin()
	expectation := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `admin_audit_logs` (`actor_user_id`,`action`,`resource_type`,`resource_id`,`detail_json`,`created_at`) VALUES (?,?,?,?,?,?)"))
	expectation.WithArgs(int64(7), "user.status.update", "user", int64(8), "{\"status\":1}", sqlmock.AnyArg()).WillReturnError(wantErr)
	mock.ExpectRollback()

	gormDB, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("create GORM DB: %v", err)
	}
	err = NewAuditRepository(gormDB).Append(context.Background(), 7, "user.status.update", "user", 8, `{"status":1}`)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected audit insert error to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestAuditRepositoryListUsesORMAndReturnsEmptyList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	gormDB, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("create GORM DB: %v", err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `admin_audit_logs` WHERE resource_type = ? AND resource_id = ?")).
		WithArgs("spot_correction", int64(12)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	list, total, err := NewAuditRepository(gormDB).List(context.Background(), AuditLogFilter{
		ResourceType: "spot_correction", ResourceID: 12, Offset: 20, Limit: 20,
	})
	if err != nil || total != 0 || list == nil || len(list) != 0 {
		t.Fatalf("List() = (%+v, %d, %v), want empty list", list, total, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestAuditRepositoryFindReviewStatesUsesORM(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	gormDB, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("create GORM DB: %v", err)
	}
	repo := NewAuditRepository(gormDB)

	correctionQuery := mock.ExpectQuery(regexp.QuoteMeta("SELECT status, review_note FROM `spot_corrections` WHERE id = ? LIMIT ?"))
	correctionQuery.WithArgs(int64(12), 1).
		WillReturnRows(sqlmock.NewRows([]string{"status", "review_note"}).AddRow("pending", ""))
	correction, err := repo.FindCorrectionReviewState(context.Background(), 12)
	if err != nil || correction == nil || correction.Status != "pending" {
		t.Fatalf("FindCorrectionReviewState() = (%+v, %v)", correction, err)
	}

	spotQuery := mock.ExpectQuery(regexp.QuoteMeta("SELECT visible_status, publish_status FROM `fishing_spots` WHERE id = ? LIMIT ?"))
	spotQuery.WithArgs(int64(13), 1).
		WillReturnRows(sqlmock.NewRows([]string{"visible_status", "publish_status"}).AddRow(0, 0))
	spot, err := repo.FindFishingSpotReviewState(context.Background(), 13)
	if err != nil || spot == nil || spot.VisibleStatus != 0 || spot.PublishStatus != 0 {
		t.Fatalf("FindFishingSpotReviewState() = (%+v, %v)", spot, err)
	}

	reportQuery := mock.ExpectQuery(regexp.QuoteMeta("SELECT moderation_status, visible_scope FROM `fishing_records` WHERE id = ? LIMIT ?"))
	reportQuery.WithArgs(int64(14), 1).
		WillReturnRows(sqlmock.NewRows([]string{"moderation_status", "visible_scope"}).AddRow("rejected", "public"))
	report, err := repo.FindPublicReportReviewState(context.Background(), 14)
	if err != nil || report == nil || report.ModerationStatus != "rejected" || report.VisibleScope != "public" {
		t.Fatalf("FindPublicReportReviewState() = (%+v, %v)", report, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestAuditRepositoryFindReviewStateReturnsNilForMissingResource(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	gormDB, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("create GORM DB: %v", err)
	}
	missingQuery := mock.ExpectQuery(regexp.QuoteMeta("SELECT status, review_note FROM `spot_corrections` WHERE id = ? LIMIT ?"))
	missingQuery.WithArgs(int64(99), 1).WillReturnRows(sqlmock.NewRows([]string{"status", "review_note"}))
	state, err := NewAuditRepository(gormDB).FindCorrectionReviewState(context.Background(), 99)
	if err != nil || state != nil {
		t.Fatalf("FindCorrectionReviewState() = (%+v, %v), want nil state", state, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
