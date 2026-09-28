package assembler

import (
	"testing"

	"fishing-notes-admin-api/internal/model"
	"fishing-notes-admin-api/internal/types"
)

func TestAdminUsersUsesEmptyRolesList(t *testing.T) {
	users := AdminUsers([]model.AdminUser{{ID: 7, Username: "operator"}}, map[int64][]types.RoleItem{})
	if users[0].Roles == nil {
		t.Fatal("expected an empty roles list, got nil")
	}
	if len(users[0].Roles) != 0 {
		t.Fatalf("expected no roles, got %d", len(users[0].Roles))
	}
}
