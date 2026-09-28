package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"fishing-notes-admin-api/internal/permission"
)

type swaggerDocument struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

func TestRBACRoleAPIsHaveSwaggerPaths(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	documentPath := filepath.Join(filepath.Dir(file), "../../desc/swagger.json")
	content, err := os.ReadFile(documentPath)
	if err != nil {
		t.Fatalf("read swagger document: %v", err)
	}
	var document swaggerDocument
	if err := json.Unmarshal(content, &document); err != nil {
		t.Fatalf("parse swagger document: %v", err)
	}

	for _, definition := range permission.BuiltinDefinitions() {
		if definition.Type != "api" || (!strings.HasPrefix(definition.Code, "rbac.role") && !strings.HasPrefix(definition.Code, "review.task.") && !strings.HasPrefix(definition.Code, "public.fishing_report.") && !strings.HasPrefix(definition.Code, "media.asset.") && !strings.HasPrefix(definition.Code, "public.spot.application.") && !strings.HasPrefix(definition.Code, "banner.")) {
			continue
		}
		path := definition.Path
		for _, segment := range strings.Split(path, "/") {
			if strings.HasPrefix(segment, ":") {
				path = strings.Replace(path, segment, "{"+strings.TrimPrefix(segment, ":")+"}", 1)
			}
		}
		methods := document.Paths[path]
		if methods == nil {
			t.Errorf("permission %q has no swagger path %q", definition.Code, path)
			continue
		}
		if _, ok := methods[strings.ToLower(definition.Method)]; !ok {
			t.Errorf("permission %q has no swagger operation %s %s", definition.Code, definition.Method, path)
		}
	}
	if _, ok := document.Paths["/api/v1/audit-logs"]["get"]; !ok {
		t.Error("audit log endpoint GET /api/v1/audit-logs is missing from Swagger")
	}
	if _, ok := document.Paths["/api/v1/audit-logs/export"]["get"]; !ok {
		t.Error("audit log export endpoint GET /api/v1/audit-logs/export is missing from Swagger")
	}
	if _, ok := document.Paths["/api/v1/dashboard/overview"]["get"]; !ok {
		t.Error("dashboard overview endpoint GET /api/v1/dashboard/overview is missing from Swagger")
	}
}
