package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// reportHour is when the daily report toast fires (local time).
const reportHour = 21 // 9:00 PM

// runScheduler fires the daily report once per day, at or
// after reportHour. It wakes up only every 30 seconds, so
// it adds nothing to the daemon's footprint.
func runScheduler(ctx context.Context, analyzer *Analyzer, notifier Notifier, lg *log.Logger) {
	reported := ""
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			now := time.Now()
			if now.Hour() < reportHour {
				continue
			}
			day := now.Format("2006-01-02")
			if day == reported {
				continue // already reported today
			}
			reported = day

			msg, err := dailyReport(analyzer)
			if err != nil {
				lg.Printf("daily report: %v", err)
				continue
			}
			notifier("Arjan Daily Report", msg)
			lg.Printf("daily report: %s", msg)
		}
	}
}

func fmtDur(secs int64) string {
	h, m := secs/3600, (secs%3600)/60
	if h > 0 {
		return fmt.Sprintf("%dh %02dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
