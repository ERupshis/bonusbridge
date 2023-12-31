package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/erupshis/bonusbridge/internal/accrual"
	"github.com/erupshis/bonusbridge/internal/accrual/client"
	"github.com/erupshis/bonusbridge/internal/accrual/workerspool"
	"github.com/erupshis/bonusbridge/internal/auth"
	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	postgresUsers "github.com/erupshis/bonusbridge/internal/auth/users/managers"
	"github.com/erupshis/bonusbridge/internal/bonuses"
	bonusesStorage "github.com/erupshis/bonusbridge/internal/bonuses/storage"
	postgresBonuses "github.com/erupshis/bonusbridge/internal/bonuses/storage/managers"
	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders"
	ordersStorage "github.com/erupshis/bonusbridge/internal/orders/storage"
	postgresOrders "github.com/erupshis/bonusbridge/internal/orders/storage/managers"
	"github.com/go-chi/chi/v5"
)

func main() {
	//config.
	cfg := config.Parse()

	//log system.
	log, err := logger.CreateZapLogger(cfg.LogLevel)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create logger: %v", err)
	}
	defer log.Sync()

	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	databaseConn, err := db.CreateConnection(ctxWithCancel, cfg, log)
	if err != nil {
		log.Info("failed to connect to users database: %v", err)
	}

	//authentication.
	usersStorage := postgresUsers.Create(databaseConn, log)
	jwtGenerator := jwtgenerator.Create(cfg.JWTKey, 2, log)
	authController := auth.CreateController(usersStorage, jwtGenerator, log)

	//orders.
	ordersManager := postgresOrders.Create(databaseConn, log)
	ordersStrg := ordersStorage.Create(ordersManager, log)
	ordersController := orders.CreateController(ordersStrg, log)

	//bonuses.
	bonusesManager := postgresBonuses.Create(databaseConn, log)
	bonusesStrg := bonusesStorage.Create(bonusesManager, log)
	bonusesController := bonuses.CreateController(bonusesStrg, log)

	//accrual(orders update) system.
	workersPool := workerspool.Create(4, log)
	defer workersPool.CloseJobsChan()
	defer workersPool.CloseResultsChan()

	requestClient := client.CreateDefault(log)
	accrualController := accrual.CreateController(ordersStrg, bonusesStrg, requestClient, workersPool, cfg, log)
	accrualController.Run(ctxWithCancel, 5)

	//controllers mounting.
	router := chi.NewRouter()
	router.Use(log.LogHandler)

	router.Mount("/api/user/register", authController.RouteRegister())
	router.Mount("/api/user/login", authController.RouteLoginer())

	router.Group(func(r chi.Router) {
		r.Use(authController.AuthorizeUser(data.RoleUser))

		r.Mount("/api/user/orders", ordersController.Route())
		r.Mount("/api/user/balance", bonusesController.RouteBonuses())
		r.Mount("/api/user/withdrawals", bonusesController.RouteWithdrawals())
	})

	//server launch.
	go func() {
		log.Info("server is launching with Host setting: %s", cfg.HostAddr)
		if err := http.ListenAndServe(cfg.HostAddr, router); err != nil {
			log.Info("server refused to start with error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
