package database

import (
	"database/sql"
	"log"
	"sync"

	"github.com/barluscuda/laodvx-server-auth/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func Connect(cfg config.DatabaseConfig) *gorm.DB {
	once.Do(func() {
		var err error
		db, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
		if err != nil {
			log.Fatal("failed to connect to database:", err)
		}
		configurePool(db, cfg)
		log.Println("database connected")
	})
	return db
}

func configurePool(db *gorm.DB, cfg config.DatabaseConfig) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get database pool:", err)
	}

	applyPoolSize(sqlDB, cfg.MaxOpenConns, cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
}

func applyPoolSize(db *sql.DB, maxOpenConns, maxIdleConns int) {
	if maxOpenConns > 0 {
		db.SetMaxOpenConns(maxOpenConns)
	}
	if maxIdleConns > 0 {
		if maxOpenConns > 0 && maxIdleConns > maxOpenConns {
			maxIdleConns = maxOpenConns
		}
		db.SetMaxIdleConns(maxIdleConns)
	}
}
