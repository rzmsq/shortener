package stat

import "time"

type ClickStat struct {
	ID        int       `json:"id"`
	UserAgent string    `json:"user_agent"`
	TimeClick time.Time `json:"time_click"`
}
