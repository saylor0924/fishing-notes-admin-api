package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestListAllRolesPropagatesQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectQuery("SELECT \\* FROM `admin_roles` ORDER BY sort ASC,id ASC").WillReturnError(wantErr)

	_, err = NewAdminRepository(openBusinessTestGORM(t, db)).ListAllRoles(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected query error to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestFindUserByUsernameUsesGORMAndMapsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT \\* FROM `admin_users` WHERE username = \\? LIMIT \\?").
		WithArgs("admin", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "nickname", "status", "is_super", "last_login_at"}))

	_, err = NewAdminRepository(openBusinessTestGORM(t, db)).FindUserByUsername(context.Background(), "admin")
	if !errors.Is(err, sqlx.ErrNotFound) {
		t.Fatalf("FindUserByUsername() error = %v, want not-found sentinel", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestUpdateAdminUserPropagatesGORMError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	wantErr := errors.New("database unavailable")
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `admin_users` SET").
		WithArgs("管理员", int64(1), sqlmock.AnyArg(), int64(9)).
		WillReturnError(wantErr)
	mock.ExpectRollback()

	err = NewAdminRepository(openBusinessTestGORM(t, db)).UpdateAdminUser(context.Background(), 9, "管理员", 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("UpdateAdminUser() error = %v, want %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
