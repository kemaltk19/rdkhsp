package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"radikal-hesap/config"
	"radikal-hesap/database"
	"radikal-hesap/router"
	"radikal-hesap/services"
	"radikal-hesap/utils"
)

func main() {
	cfg := config.Load()

	// Production'da Gin debug modunu kapat.
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	database.Connect(cfg.DatabaseURL, cfg.SystemDatabaseURL)
	database.Migrate()
	utils.InitMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)

	// Start background expense scheduler
	expenseService := services.NewExpenseService()
	scheduler := services.NewExpenseScheduler(expenseService)
	scheduler.Start(context.Background())

	r := router.Setup()

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("[server] http://localhost:%s (env=%s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[server] dinleme hatasi: %v", err)
		}
	}()

	// Graceful shutdown: SIGINT/SIGTERM'de aktif istekleri tamamlamaya calis.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[server] kapatiliyor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[server] kapatma hatasi: %v", err)
	}
	log.Println("[server] temiz kapandi")
}
