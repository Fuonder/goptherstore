package main

import (
	"context"
	"fmt"
	"github.com/Fuonder/goptherstore.git/internal/accrualservice"
	"github.com/Fuonder/goptherstore.git/internal/httpserver"
	"github.com/Fuonder/goptherstore.git/internal/logger"
	st "github.com/Fuonder/goptherstore.git/internal/storage"
	"github.com/Fuonder/goptherstore.git/internal/storage/postrge"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"log"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	DBConn, err := postrge.NewConnection(ctx, CliOptions.DatabaseDSN)
	if err != nil {
		return err
	}

	g := new(errgroup.Group)
	jobsCh := make(chan st.MartOrder, 10)
	defer close(jobsCh)

	storage := postrge.NewPsqlStorage(ctx, DBConn, []byte(CliOptions.Key), jobsCh)

	service, err := httpserver.NewService(CliOptions.APIAddress.String(), storage)
	if err != nil {
		return err
	}

	BonusAPIService := accrualservice.NewBonusAPIService(storage, jobsCh, CliOptions.AccrualAddress.String())

	g.Go(func() error {
		err = service.Run()
		if err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		err = BonusAPIService.Run()
		if err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Log.Debug("exit with error", zap.Error(err))
		cancel()
		return err
	}
	return nil
}
