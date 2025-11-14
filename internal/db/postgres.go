package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func ConnectPostgres(url string) error {
	var err error
	Pool, err = pgxpool.New(context.Background(), url)
	return err
}
