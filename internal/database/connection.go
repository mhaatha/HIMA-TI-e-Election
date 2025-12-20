package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/config"
)

func ConnectDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DBURL)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}
