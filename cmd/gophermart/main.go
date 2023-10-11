package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/erupshis/bonusbridge/internal/auth"
	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	ramUsers "github.com/erupshis/bonusbridge/internal/auth/users/managers/ram"
	"github.com/erupshis/bonusbridge/internal/auth/users/userdata"
	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/logger"
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

	//authentication.
	usersStorage := ramUsers.Create(log)
	jwtGenerator := jwtgenerator.Create(cfg.JWTKey, 2, log)
	authController := auth.CreateAuthenticator(usersStorage, jwtGenerator, log)

	//ctxWithCancel, cancel := context.WithCancel(context.Background())
	//defer cancel()
	//
	////orders.
	//storageManager, err := postgresOrders.CreatePostgreDB(ctxWithCancel, cfg, log)
	//if err != nil {
	//	log.Info("failed to connect to orders database: %v", err)
	//	return
	//}
	//
	//ordersStorage := storage.Create(storageManager, log)
	//ordersController := controller.CreateController(ordersStorage, log)

	//controllers mounting.
	router := chi.NewRouter()
	//router.Mount("/", authController.Route()) TODO: main page plug.
	router.Mount("/api/user/register", authController.RouteRegister())
	router.Mount("/api/user/login", authController.RouteLoginer())
	router.Mount("/api/user/orders", authController.AuthorizeUser(chi.NewRouter(), userdata.RoleUser))

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
