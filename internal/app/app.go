package app

import (
	"context"
	"database/sql"

	"github.com/spelens-gud/golangci-scope/internal/config"
)

type App struct {
}

func New(ctx context.Context, conn *sql.DB, cfg *config.Config) (*App, error) {
	//q := db.New(conn)
	return &App{}, nil
}
