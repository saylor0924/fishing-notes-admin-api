package audit

import "encoding/json"

// Event 描述一次需要留痕的后台操作。
type Event struct {
	ActorUserID  int64
	Action       string
	ResourceType string
	ResourceID   int64
	Detail       any
}

func MarshalDetail(value any) (string, error) {
	if value == nil {
		return "{}", nil
	}

	content, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
