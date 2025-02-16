package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// scheduler will pick all urlMonitors and then will send a head req and store relevent data to db

func NewScheduler() *cron.Cron {
	cron := cron.New()
	cron.AddFunc("* * * * *", func() {
		fmt.Println("job runnig at ", time.Now())
	})
	return cron
}
