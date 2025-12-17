package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_calendar/internal/app"                          //nolint:depguard
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_calendar/internal/logger"                       //nolint:depguard
	internalhttp "github.com/stas-ik/otus-go-test/hw12_13_14_15_calendar/internal/server/http"     //nolint:depguard
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_calendar/internal/storage"                      //nolint:depguard
	memorystorage "github.com/stas-ik/otus-go-test/hw12_13_14_15_calendar/internal/storage/memory" //nolint:depguard
	sqlstorage "github.com/stas-ik/otus-go-test/hw12_13_14_15_calendar/internal/storage/sql"       //nolint:depguard
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

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
		// Завершаем контекст после установления соединения
		cancel()
		stor = sqlStorage
		logg.Info("Using SQL storage")
	default:
		logg.Error(fmt.Sprintf("Unknown storage type: %s", config.Storage.Type))
		os.Exit(1)
	}

	calendar := app.New(logg, stor)

	server := internalhttp.NewServer(logg, calendar, config.Server.Host, config.Server.Port)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	// do not defer cancel here to avoid gocritic exitAfterDefer warning when using os.Exit

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}

		if sqlStor, ok := stor.(*sqlstorage.Storage); ok {
			if err := sqlStor.Close(ctx); err != nil {
				logg.Error("failed to close database connection: " + err.Error())
			}
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1)
	}
}
