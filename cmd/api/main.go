package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/server"
)

func main() {
	ctx := context.Background()

	appInstance := app.NewApp(ctx)
	cfg := appInstance.Config
	serverCfg := cfg.Server

	ginRouter := server.StartServer(appInstance)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", serverCfg.Port),
		Handler:      ginRouter,
		ReadTimeout:  serverCfg.ReadTimeout,
		WriteTimeout: serverCfg.WriteTimeout,
	}

	log.Printf("Starting server on port %d in %s mode", serverCfg.Port, cfg.App.Env)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
