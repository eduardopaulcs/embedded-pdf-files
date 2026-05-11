package service

import (
	"fmt"
	"time"
)

func HumanDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	if minutes >= 60 {
		hours := minutes / 60
		mins := minutes % 60
		if mins == 0 {
			if hours == 1 {
				return "1 hour"
			}
			return fmt.Sprintf("%d hours", hours)
		}
		if hours == 1 {
			return fmt.Sprintf("1 hour %d minutes", mins)
		}
		return fmt.Sprintf("%d hours %d minutes", hours, mins)
	}
	if minutes == 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", minutes)
}
