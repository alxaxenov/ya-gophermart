package main

import (
	"log"

	"context"

	"github.com/alxaxenov/ya-gophermart/internal/client/accrual"
	"github.com/alxaxenov/ya-gophermart/internal/config"
	"github.com/alxaxenov/ya-gophermart/internal/config/db"
	"github.com/alxaxenov/ya-gophermart/internal/handler"
	"github.com/alxaxenov/ya-gophermart/internal/logger"
	"github.com/alxaxenov/ya-gophermart/internal/middleware"
	"github.com/alxaxenov/ya-gophermart/internal/repo"
	"github.com/alxaxenov/ya-gophermart/internal/service/gophermart"
	"github.com/alxaxenov/ya-gophermart/internal/service/user"
	accrualUpd "github.com/alxaxenov/ya-gophermart/internal/worker/accrual"
)

func main() {
	if err := logger.Initialize(); err != nil {
		log.Fatal(err)
	}
	if err := run(); err != nil {
		logger.Logger.Fatal(err)
	}
}

func run() error {
	cfg, err := config.ParseConfig()
	if err != nil {
		return err
	}

	connector := db.NewPGConnector(cfg.DBURI)
	_, err = connector.Open()
	if err != nil {
		return err
	}
	defer connector.Close()

	userRepo := repo.NewUserRepo(connector)
	userService := user.NewUserService(
		cfg.UserSecret,
		config.CookieAuthKey,
		cfg.TokenExpireHours,
		userRepo,
	)

	gophermartRepo := repo.NewGophermartRepo(connector)
	gophermartService := gophermart.NewGophermartService(gophermartRepo)

	GophermartHandler := handler.NewGophermartHandler(userService, gophermartService)
	userMiddleware := middleware.NewUserAuthMiddleware(userService)

	accrualClient := accrual.NewHTTPClient(cfg.AcrualAddr, 3)
	ctx, c := context.WithCancel(context.Background())
	defer c()
	accrualUpdater := accrualUpd.NewUpdater(ctx, gophermartRepo, accrualClient, cfg.WorkersNumber)
	go accrualUpdater.Run(ctx)

	return handler.Serve(cfg.RunAddr, GophermartHandler, userMiddleware)

	//TODO graceful shutdown
	//TODO тесты
	//TODO документация
}
