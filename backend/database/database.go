package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB       *gorm.DB
	SystemDB *gorm.DB
)

// Connect opens both App DB (RLS) and System DB (Bypass RLS) connections,
// configures the connection pool and GORM logger.
func Connect(dsn, systemDsn string) {
	if dsn == "" || systemDsn == "" {
		log.Fatal("[db] DATABASE_URL ve SYSTEM_DATABASE_URL bos olamaz")
	}

	// Production'da SQL log gurultusunu azalt.
	logLevel := logger.Info
	if os.Getenv("APP_ENV") == "production" {
		logLevel = logger.Error
	}
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		log.Fatalf("[db] App DB baglanti hatasi: %v", err)
	}
	tunePool(DB)
	log.Println("[db] App DB (radikal_app) baglandi")

	SystemDB, err = gorm.Open(postgres.Open(systemDsn), gormCfg)
	if err != nil {
		log.Fatalf("[db] System DB baglanti hatasi: %v", err)
	}
	tunePool(SystemDB)
	log.Println("[db] System DB (radikal_system) baglandi")
}

// tunePool, yuk altinda baglanti tukenmesini onlemek icin makul limitler uygular.
func tunePool(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("[db] pool ayari alinamadi: %v", err)
		return
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
}
