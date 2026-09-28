package service

import (
	"strings"
	"testing"

	"fishing-notes-admin-api/internal/model"
)

func TestBuildSpotCorrectionDiffMapsStructuredChanges(t *testing.T) {
	item := &model.SpotCorrection{
		Content:     `{"message":"修正地址和标签","changes":{"address":"新河岸","tagType":"paid","unknown":"ignored"}}`,
		CurrentSpot: &model.FishingSpot{Address: "旧河岸", TagType: "free"},
	}

	diff := buildSpotCorrectionDiff(item)
	if !diff.Structured || diff.Message != "修正地址和标签" || len(diff.Fields) != 2 {
		t.Fatalf("diff = %+v", diff)
	}
	changes, err := selectedSpotCorrectionChanges(item, []string{"address"})
	if err != nil || changes["address"] != "新河岸" || len(changes) != 1 {
		t.Fatalf("selected changes = %+v, err = %v", changes, err)
	}
}

func TestSelectedSpotCorrectionChangesDefaultsToAllStructuredFields(t *testing.T) {
	item := &model.SpotCorrection{
		Content:     `{"changes":{"name":"新名称"}}`,
		CurrentSpot: &model.FishingSpot{Name: "旧名称"},
	}
	changes, err := selectedSpotCorrectionChanges(item, nil)
	if err != nil || len(changes) != 1 || changes["name"] != "新名称" {
		t.Fatalf("changes = %+v, err = %v", changes, err)
	}
}

func TestBuildSpotCorrectionDiffKeepsPlainTextAsMessage(t *testing.T) {
	diff := buildSpotCorrectionDiff(&model.SpotCorrection{Content: "位置描述不准确"})
	if diff.Structured || diff.Message != "位置描述不准确" || len(diff.Fields) != 0 {
		t.Fatalf("diff = %+v", diff)
	}
}

func TestSelectedSpotCorrectionChangesValidatesCoordinates(t *testing.T) {
	item := &model.SpotCorrection{
		Content:     `{"changes":{"latitude":"91"}}`,
		CurrentSpot: &model.FishingSpot{Latitude: 30},
	}
	if _, err := selectedSpotCorrectionChanges(item, nil); err == nil {
		t.Fatal("expected latitude validation error")
	}
}

func TestSelectedSpotCorrectionChangesRejectsUnsupportedStructuredFields(t *testing.T) {
	item := &model.SpotCorrection{
		Content:     `{"changes":{"unsupported":"value"}}`,
		CurrentSpot: &model.FishingSpot{},
	}
	if _, err := selectedSpotCorrectionChanges(item, nil); err == nil {
		t.Fatal("expected unsupported field error")
	}
}

func TestSelectedSpotCorrectionChangesRejectsValuesLongerThanSpotColumns(t *testing.T) {
	item := &model.SpotCorrection{
		Content:     `{"changes":{"tagText":"` + strings.Repeat("字", 65) + `"}}`,
		CurrentSpot: &model.FishingSpot{},
	}
	if _, err := selectedSpotCorrectionChanges(item, nil); err == nil {
		t.Fatal("expected maximum field length error")
	}
}

func TestSpotCorrectionAuditChangesKeepsBeforeAndAfterValues(t *testing.T) {
	item := &model.SpotCorrection{
		Content:     `{"changes":{"address":"新河岸","latitude":31.2}}`,
		CurrentSpot: &model.FishingSpot{Address: "旧河岸", Latitude: 30},
	}
	auditChanges := spotCorrectionAuditChanges(item, map[string]string{"address": "新河岸", "latitude": "31.2"})
	if auditChanges["address"]["before"] != "旧河岸" || auditChanges["address"]["after"] != "新河岸" {
		t.Fatalf("address audit changes = %+v", auditChanges["address"])
	}
	if auditChanges["latitude"]["before"] != "30" || auditChanges["latitude"]["after"] != "31.2" {
		t.Fatalf("latitude audit changes = %+v", auditChanges["latitude"])
	}
}
