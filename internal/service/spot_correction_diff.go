package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"fishing-notes-admin-api/internal/model"
	"fishing-notes-admin-api/internal/types"
)

type spotCorrectionPayload struct {
	Message string                     `json:"message"`
	Changes map[string]json.RawMessage `json:"changes"`
}

type spotCorrectionFieldDefinition struct {
	Label     string
	MaxLength int
	Value     func(*model.FishingSpot) string
}

var spotCorrectionFields = map[string]spotCorrectionFieldDefinition{
	"name":      {Label: "名称", MaxLength: 128, Value: func(spot *model.FishingSpot) string { return spot.Name }},
	"province":  {Label: "省份", MaxLength: 64, Value: func(spot *model.FishingSpot) string { return spot.Province }},
	"city":      {Label: "城市", MaxLength: 64, Value: func(spot *model.FishingSpot) string { return spot.City }},
	"district":  {Label: "区县", MaxLength: 64, Value: func(spot *model.FishingSpot) string { return spot.District }},
	"address":   {Label: "地址", MaxLength: 255, Value: func(spot *model.FishingSpot) string { return spot.Address }},
	"tagText":   {Label: "标签", MaxLength: 64, Value: func(spot *model.FishingSpot) string { return spot.TagText }},
	"tagType":   {Label: "收费类型", MaxLength: 32, Value: func(spot *model.FishingSpot) string { return spot.TagType }},
	"scene":     {Label: "场景结论", MaxLength: 64, Value: func(spot *model.FishingSpot) string { return spot.Scene }},
	"sceneHint": {Label: "场景建议", MaxLength: 255, Value: func(spot *model.FishingSpot) string { return spot.SceneHint }},
	"latitude":  {Label: "纬度", Value: func(spot *model.FishingSpot) string { return strconv.FormatFloat(spot.Latitude, 'f', -1, 64) }},
	"longitude": {Label: "经度", Value: func(spot *model.FishingSpot) string { return strconv.FormatFloat(spot.Longitude, 'f', -1, 64) }},
}

var spotCorrectionFieldOrder = []string{
	"name", "province", "city", "district", "address", "tagText", "tagType", "scene", "sceneHint", "latitude", "longitude",
}

func buildSpotCorrectionDiff(item *model.SpotCorrection) types.SpotCorrectionDiff {
	result := types.SpotCorrectionDiff{Fields: []types.SpotCorrectionFieldChange{}}
	if item == nil {
		return result
	}
	payload := spotCorrectionPayload{}
	if err := json.Unmarshal([]byte(item.Content), &payload); err != nil || payload.Changes == nil {
		result.Message = strings.TrimSpace(item.Content)
		return result
	}
	result.Structured = true
	result.Message = strings.TrimSpace(payload.Message)
	for _, field := range spotCorrectionFieldOrder {
		raw, exists := payload.Changes[field]
		if !exists {
			continue
		}
		definition, ok := spotCorrectionFields[field]
		if !ok || item.CurrentSpot == nil {
			continue
		}
		proposed, ok := rawValueText(raw)
		if !ok {
			continue
		}
		current := definition.Value(item.CurrentSpot)
		result.Fields = append(result.Fields, types.SpotCorrectionFieldChange{
			Field: field, Label: definition.Label, Current: current, Proposed: proposed,
		})
	}
	return result
}

func selectedSpotCorrectionChanges(item *model.SpotCorrection, fields []string) (map[string]string, error) {
	diff := buildSpotCorrectionDiff(item)
	if !diff.Structured {
		if len(fields) > 0 {
			return nil, fmt.Errorf("spot correction does not contain structured fields")
		}
		return nil, nil
	}
	if item == nil || item.CurrentSpot == nil {
		return nil, fmt.Errorf("spot correction current spot is unavailable")
	}
	if len(diff.Fields) == 0 {
		return nil, fmt.Errorf("spot correction does not contain supported fields")
	}
	available := make(map[string]string, len(diff.Fields))
	for _, field := range diff.Fields {
		available[field.Field] = field.Proposed
	}
	if len(fields) == 0 {
		if err := validateSpotCorrectionChanges(available); err != nil {
			return nil, err
		}
		return available, nil
	}
	selected := make(map[string]string, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		value, ok := available[field]
		if !ok {
			return nil, fmt.Errorf("spot correction field %q is not available", field)
		}
		selected[field] = value
	}
	if err := validateSpotCorrectionChanges(selected); err != nil {
		return nil, err
	}
	return selected, nil
}

func spotCorrectionAuditChanges(item *model.SpotCorrection, changes map[string]string) map[string]map[string]string {
	if item == nil || len(changes) == 0 {
		return nil
	}
	diff := buildSpotCorrectionDiff(item)
	currentByField := make(map[string]string, len(diff.Fields))
	for _, field := range diff.Fields {
		currentByField[field.Field] = field.Current
	}
	result := make(map[string]map[string]string, len(changes))
	for field, proposed := range changes {
		result[field] = map[string]string{
			"before": currentByField[field],
			"after":  proposed,
		}
	}
	return result
}

func validateSpotCorrectionChanges(changes map[string]string) error {
	for field, value := range changes {
		definition, ok := spotCorrectionFields[field]
		if !ok {
			continue
		}
		if definition.MaxLength > 0 && len([]rune(value)) > definition.MaxLength {
			return fmt.Errorf("spot correction field %q exceeds %d characters", field, definition.MaxLength)
		}
		if field != "latitude" && field != "longitude" {
			continue
		}
		coordinate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("spot correction field %q must be a number", field)
		}
		if field == "latitude" && (coordinate < -90 || coordinate > 90) {
			return fmt.Errorf("spot correction latitude is out of range")
		}
		if field == "longitude" && (coordinate < -180 || coordinate > 180) {
			return fmt.Errorf("spot correction longitude is out of range")
		}
	}
	return nil
}

func rawValueText(raw json.RawMessage) (string, bool) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil || value == nil {
		return "", false
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed), true
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(typed), true
	default:
		return "", false
	}
}
