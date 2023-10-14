package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/erupshis/bonusbridge/internal/accrual"
	"github.com/erupshis/bonusbridge/internal/accrual/client"
	"github.com/erupshis/bonusbridge/internal/auth"
	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	postgresUsers "github.com/erupshis/bonusbridge/internal/auth/users/managers/postgresql"
	"github.com/erupshis/bonusbridge/internal/bonuses"
	bonusesStorage "github.com/erupshis/bonusbridge/internal/bonuses/storage"
	postgresBonuses "github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql"
	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders"
	ordersStorage "github.com/erupshis/bonusbridge/internal/orders/storage"
	postgresOrders "github.com/erupshis/bonusbridge/internal/orders/storage/managers/postgresql"
	"github.com/go-chi/chi/v5"
)

func main() {
	//config.
	cfg := config.Parse()

	//log system.
	log, err := logger.CreateZapLogger("info")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create logger: %v", err)
	}
	defer log.Sync()

	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()
	dbMutex := &sync.RWMutex{}

	//authentication.
	usersStorage, err := postgresUsers.CreateUsersPostgreDB(ctxWithCancel, cfg, dbMutex, log)
	if err != nil {
		log.Info("failed to connect to users database: %v", err)
	}

	jwtGenerator := jwtgenerator.Create(cfg.JWTKey, 2, log)
	authController := auth.CreateController(usersStorage, jwtGenerator, log)

	//orders.
	ordersManager, err := postgresOrders.CreateOrdersPostgreDB(ctxWithCancel, cfg, dbMutex, log)
	if err != nil {
		log.Info("failed to connect to orders database: %v", err)
	}

	ordersStrg := ordersStorage.Create(ordersManager, log)
	ordersController := orders.CreateController(ordersStrg, log)

	//bonuses.
	bonusesManager, err := postgresBonuses.CreateBonusesPostgreDB(ctxWithCancel, cfg, dbMutex, log)
	if err != nil {
		log.Info("failed to connect to orders database: %v", err)
	}

	bonusesStrg := bonusesStorage.Create(bonusesManager, log)
	bonusesController := bonuses.CreateController(bonusesStrg, log)

	//accrual(orders update) system.
	requestClient := client.CreateDefault(log)
	accrualController := accrual.CreateController(ordersStrg, bonusesStrg, requestClient, cfg, log)
	accrualController.Run(ctxWithCancel, 5)

	//controllers mounting.
	router := chi.NewRouter()
	router.Mount("/api/user/register", authController.RouteRegister())
	router.Mount("/api/user/login", authController.RouteLoginer())
	router.Mount("/api/user/orders", authController.AuthorizeUser(ordersController.Route(), data.RoleUser))
	router.Mount("/api/user/balance", authController.AuthorizeUser(bonusesController.Route(), data.RoleUser))

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
