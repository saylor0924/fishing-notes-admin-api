package assembler

import (
	"database/sql"
	"time"
)

func NullTimePtr(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}

	formatted := Time(value.Time)
	return &formatted
}

func Time(value time.Time) string {
	return value.Format(time.DateTime)
}
