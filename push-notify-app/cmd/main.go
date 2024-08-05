package main

import (
	"context"
	"os"

	"github.com/murasame29/image-registry-push-notify/sample-app/cmd/config"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/log"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/routes"
	"github.com/murasame29/image-registry-push-notify/sample-app/internal/server"
)

func init() {
	if err := config.LoadEnv(); err != nil {
		log.Error(context.TODO(), "failed to load env. error: %v", err)
		os.Exit(1)
	}
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	route := routes.NewRoutes()

	if err := server.New(route).RunWithGracefulShutdown(context.Background()); err != nil {
		return err
	}
	return nil
}
