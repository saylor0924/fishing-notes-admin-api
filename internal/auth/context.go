package auth

import (
	"context"
	"encoding/json"
	"strconv"
)

func UserIDFromContext(ctx context.Context) (int64, bool) {
	return Int64Claim(ctx, "userId")
}

func Int64Claim(ctx context.Context, key string) (int64, bool) {
	if ctx == nil {
		return 0, false
	}

	value := ctx.Value(key)
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case json.Number:
		n, err := v.Int64()
		return n, err == nil
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		return n, err == nil
	default:
		return 0, false
	}
}
