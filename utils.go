package cornacchia

import (
	"fmt"
	"time"
)

// https://github.com/bakins/alertmanager-webhook-example/blob/master/util.go

// MarshalJSON emits a timestamp suitable for use in json
func (t *Timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(*t).Format(time.RFC3339)
	stamp := fmt.Sprint(ts)
	return []byte(stamp), nil
}

// UnmarshalJSON reads back data from json
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = s[1 : len(s)-1]
	ts, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		ts, err = time.Parse("2006-01-02T15:04:05Z07:00", s)
		if err != nil {
			return err
		}
	}
	*t = Timestamp(ts)
	return nil
}
