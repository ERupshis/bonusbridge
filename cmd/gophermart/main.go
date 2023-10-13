package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/erupshis/bonusbridge/internal/auth"
	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	postgresUsers "github.com/erupshis/bonusbridge/internal/auth/users/managers/postgresql"
	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders"
	"github.com/erupshis/bonusbridge/internal/orders/storage"
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

	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	//authentication.
	usersStorage, err := postgresUsers.CreateUsersPostgreDB(ctxWithCancel, cfg, log)
	if err != nil {
		log.Info("failed to connect to users database: %v", err)
	}

	jwtGenerator := jwtgenerator.Create(cfg.JWTKey, 2, log)
	authController := auth.CreateController(usersStorage, jwtGenerator, log)

	//orders.
	storageManager, err := postgresOrders.CreateOrdersPostgreDB(ctxWithCancel, cfg, log)
	if err != nil {
		log.Info("failed to connect to orders database: %v", err)
	}

	ordersStorage := storage.Create(storageManager, log)
	ordersController := orders.CreateController(ordersStorage, log)

	//controllers mounting.
	router := chi.NewRouter()
	router.Mount("/api/user/register", authController.RouteRegister())
	router.Mount("/api/user/login", authController.RouteLoginer())
	router.Mount("/api/user/orders", authController.AuthorizeUser(ordersController.Route(), data.RoleUser))

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
