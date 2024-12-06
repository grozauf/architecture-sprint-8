package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/tbaehler/gin-keycloak/pkg/ginkeycloak"
)

const (
	SERVER_PORT_ENV = "SERVER_PORT"
	SERVER_PORT     = "8000"

	KEYCLOAK_URL_ENV   = "KEYCLOAK_URL"
	KEYCLOAK_URL       = "http://keycloak:8080"
	KEYCLOAK_REALM_ENV = "KEYCLOAK_REALM"
	KEYCLOAK_REALM     = "reports-realm"

	ALLOWED_USER_ROLE = "prothetic_user"
)

func main() {
	viper.AutomaticEnv()
	viper.SetDefault(SERVER_PORT_ENV, SERVER_PORT)
	viper.SetDefault(KEYCLOAK_URL_ENV, KEYCLOAK_URL)
	viper.SetDefault(KEYCLOAK_REALM_ENV, KEYCLOAK_REALM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf := cors.DefaultConfig()
	conf.AllowAllOrigins = true
	conf.AllowCredentials = true
	conf.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}

	router := gin.Default()
	router.Use(cors.New(conf))

	keycloakConfig := ginkeycloak.BuilderConfig{
		Url:   viper.GetString(KEYCLOAK_URL_ENV),
		Realm: viper.GetString(KEYCLOAK_REALM_ENV),
	}

	router.Use(
		ginkeycloak.
			NewAccessBuilder(keycloakConfig).
			RestrictButForRealm(ALLOWED_USER_ROLE).
			Build(),
	)

	router.GET("/reports", func(context *gin.Context) {
		time.Sleep(time.Second) // имитируем загрузку отчёта
		context.JSON(200, gin.H{
			"result": "ok",
		})
	})

	srv := &http.Server{
		Addr:    ":" + viper.GetString(SERVER_PORT_ENV),
		Handler: router.Handler(),
	}

	// Ловим SIGINT для GS.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		if err := srv.Shutdown(ctx); err != nil {
			log.Panicf("gracefull shutdown failed: %s\n", err.Error())
		}
	}()

	log.Printf("starting server on '%s'\n", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Panicf("server listen failed: %s\n", err)
	}

	log.Println("server stopped")
}
