package scheduler

import (
	"log"
	service "nitinjuyal1610/uptimeMonitor/internal/services"

	"github.com/robfig/cron/v3"
)

type SchedulerService struct {
	urlService *service.UrlService
	cron       *cron.Cron
}

func NewScheduler(services *service.Services) *SchedulerService {
	sd := &SchedulerService{urlService: services.UrlService, cron: cron.New()}

	_, err := sd.cron.AddFunc("* * * * *", func() {
		log.Println("Running scheduled URL monitoring check...")
		sd.urlService.ProcessDueMonitorURLs()
	})

	if err != nil {
		log.Fatalf("Failed to schedule URL monitoring task: %v", err)
	}
	return sd
}

func (s *SchedulerService) Start() {
	log.Println("Starting scheduler...")
	s.cron.Start()
}

func (s *SchedulerService) Stop() {
	log.Println("Stopping scheduler...")
	s.cron.Stop()
}
