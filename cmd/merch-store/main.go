package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"merch-store-service/internal/app"
	"merch-store-service/internal/infra/config"
)

func main() {
	cfg := config.LoadConfig()

	srv := app.New(cfg)

	go srv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Printf("stopping server signal %s\n", sign.String())

	srv.Stop()

	log.Println("server stopped")
}
