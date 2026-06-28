package utils

import (
	"encoding/json"
	"time"
)

const APITimeFormat = time.RFC3339

type APITime struct {
	time.Time
}

func NewAPITime(value time.Time) APITime {
	return APITime{Time: value.UTC().Truncate(time.Second)}
}

func FormatAPITime(value time.Time) string {
	return NewAPITime(value).Format(APITimeFormat)
}

func (t APITime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(FormatAPITime(t.Time))
}
