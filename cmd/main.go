package main

import (
	"fmt"
	"net/http"

	"nitinjuyal1610/uptimeMonitor/internal/scheduler"
	"nitinjuyal1610/uptimeMonitor/internal/server"
	"strings"
)

func main() {

	srv, server := server.New()

	// start scheduler
	schedulerService := scheduler.NewScheduler(srv.Services)
	schedulerService.Start()
	defer schedulerService.Stop()
	// start server
	fmt.Printf("server Running at port %s \n", strings.Split(server.Addr, ":")[1])
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}
}
