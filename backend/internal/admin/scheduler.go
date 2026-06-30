package admin

import (
	"context"
	"time"
)

func StartMonthlyResetScheduler(s *Service) {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			_ = s.RunScheduledMonthlyReset(context.Background())
		}
	}()
}
