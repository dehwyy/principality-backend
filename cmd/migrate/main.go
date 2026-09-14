package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/dehwyy/principality-backend/internal/app/ds"
	"github.com/dehwyy/principality-backend/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.Archaeologist{},
		&ds.Principality{},
		&ds.PrincipalityLike{},
	)
	if err != nil {
		panic("cant migrate db")
	}

	err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_principality_one_draft_per_archaeologist
		ON principality (created_by) WHERE principality_status = 'draft'`).Error
	if err != nil {
		panic("cant create draft index")
	}
}
