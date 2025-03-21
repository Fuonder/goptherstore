package main

import (
	"context"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/httpserver"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	"github.com/Fuonder/goptherstore.git/internal/storage/postrge"
	"go.uber.org/zap"
	"log"
	"time"
)

func main() {
	err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}
	if err := logger.Initialize(CliOptions.LogLevel); err != nil {
		panic(fmt.Errorf("method main: %v", err))
	}
	logger.Log.Info("Flags parsed",
		zap.String("flags", CliOptions.String()))

	logger.Log.Info("Starting service")
	if err = run(); err != nil {
		logger.Log.Fatal("", zap.Error(err))
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	DBConn, err := postrge.NewConnection(ctx, CliOptions.DatabaseDSN)
	if err != nil {
		return err
	}
	storage := postrge.NewPsqlStorage(ctx, DBConn)

	service := httpserver.NewService(CliOptions.APIAddress.String(), storage)
	return service.Run()
}
