package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"errors"

	"github.com/alxaxenov/ya-gophermart/internal/client/accrual"
	"github.com/alxaxenov/ya-gophermart/internal/config"
	"github.com/alxaxenov/ya-gophermart/internal/config/db"
	"github.com/alxaxenov/ya-gophermart/internal/logger"
	"github.com/alxaxenov/ya-gophermart/internal/middleware"
	"github.com/alxaxenov/ya-gophermart/internal/repo"
	"github.com/alxaxenov/ya-gophermart/internal/server"
	"github.com/alxaxenov/ya-gophermart/internal/service/gophermart"
	"github.com/alxaxenov/ya-gophermart/internal/service/user"
	accrualUpd "github.com/alxaxenov/ya-gophermart/internal/worker/accrual"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := logger.Initialize(); err != nil {
		log.Fatal(err)
	}
	if err := run(); err != nil {
		logger.Logger.Error(err.Error())
		os.Exit(1)
	}
}

const (
	timeoutInners = 10 * time.Second
	timeoutCommon = 30 * time.Second
)

func run() error {
	rootCtx, rootCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer rootCancel()

	g, ctx := errgroup.WithContext(rootCtx)

	context.AfterFunc(ctx, func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutCommon)
		defer cancel()

		<-ctx.Done()
		logger.Logger.Error("failed to shutdown gracefully")
		os.Exit(1)
	})

	cfg, err := config.ParseConfig()
	if err != nil {
		return err
	}

	connector := db.NewPGConnector(cfg.DBURI)
	_, err = connector.Open()
	if err != nil {
		return err
	}
	// Закрытие соединения к бд
	g.Go(func() error {
		defer log.Println("closed connector")
		<-ctx.Done()
		return connector.Close()
	})

	userRepo := repo.NewUserRepo(connector)
	userService := user.NewUserService(
		cfg.UserSecret,
		config.CookieAuthKey,
		cfg.TokenExpireHours,
		userRepo,
	)

	gophermartRepo := repo.NewGophermartRepo(connector)
	gophermartService := gophermart.NewGophermartService(gophermartRepo)

	GophermartHandler := server.NewGophermartHandler(userService, gophermartService)
	userMiddleware := middleware.NewUserAuthMiddleware(userService)

	accrualClient := accrual.NewHTTPClient(cfg.AcrualAddr, 3)
	accrualUpdater := accrualUpd.NewUpdater(ctx, gophermartRepo, accrualClient, cfg.WorkersNumber)
	// Старт фонового воркера
	workerWG := &sync.WaitGroup{}
	g.Go(func() error {
		workerWG.Add(1)
		defer workerWG.Done()
		accrualUpdater.Run(ctx)
		return nil
	})
	// Завершение работы воркера
	g.Go(func() error {
		<-ctx.Done()

		workerDone := make(chan struct{})
		go func() {
			workerWG.Wait()
			close(workerDone)
		}()
		select {
		case <-workerDone:
			log.Println("worker done")
		case <-time.After(timeoutInners):
			log.Println("worker timeout")
		}
		return nil
	})

	srv := server.NewServer(cfg.RunAddr, GophermartHandler, userMiddleware)
	// Старт сервера
	g.Go(func() error {
		if err := srv.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return fmt.Errorf("start server failed: %w", err)
		}
		return nil
	})
	// Завершение работы сервера
	g.Go(func() error {
		defer log.Println("server closed")
		<-rootCtx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeoutInners)
		defer shutdownCancel()
		if err := srv.Stop(shutdownCtx); err != nil {
			log.Printf("shutdown server error: %v", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
