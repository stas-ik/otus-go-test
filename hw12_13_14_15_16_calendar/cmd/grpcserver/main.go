//go:build grpcapi

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/app"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/logger"
	grpcserver "github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/server/grpc"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage"
	memorystorage "github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	config, err := NewConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logg := logger.New(config.Logger.Level)

	var stor storage.Storage
	switch config.Storage.Type {
	case "memory":
		stor = memorystorage.New()
		logg.Info("Using in-memory storage")
	case "sql":
		sqlStorage := sqlstorage.New(config.Database.DSN)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := sqlStorage.Connect(ctx); err != nil {
			logg.Error(fmt.Sprintf("Failed to connect to database: %v", err))
			cancel()
			os.Exit(1)
		}
		cancel()
		stor = sqlStorage
		logg.Info("Using SQL storage")
	default:
		logg.Error(fmt.Sprintf("Unknown storage type: %s", config.Storage.Type))
		os.Exit(1)
	}

	application := app.New(logg, stor)

	srv, err := grpcserver.New(logg, application, config.Server.GRPCHost, config.Server.GRPCPort)
	if err != nil {
		logg.Error("failed to create grpc server: " + err.Error())
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		<-ctx.Done()
		srv.Stop()
		if sqlStor, ok := stor.(*sqlstorage.Storage); ok {
			_ = sqlStor.Close(context.Background())
		}
	}()

	logg.Info("grpc calendar is running...")
	if err := srv.Start(); err != nil {
		logg.Error("failed to start grpc server: " + err.Error())
		cancel()
		os.Exit(1)
	}
}
