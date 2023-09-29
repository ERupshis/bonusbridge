package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/erupshis/bonusbridge/internal/controllers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/go-chi/chi"
)

func main() {
	log, err := logger.CreateZapLogger("info")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create logger: %v", err)
	}

	authController := controllers.CreateAuthenticator(log)

	//controllers mounting.
	router := chi.NewRouter()
	router.Mount("/api/user/", authController.Route())

	go func() {
		log.Info("server is launching with Host setting: %s", `localhost:8080`)
		if err := http.ListenAndServe(`localhost:8080`, router); err != nil {
			log.Info("server refused to start with error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
