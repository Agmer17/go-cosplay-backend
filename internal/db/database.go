package db

import (
	"context"
	"time"

	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool  *pgxpool.Pool
	Query *generated.Queries
}

func NewDatabase(url string, ctx context.Context) *Database {
	println("LOG : TRYING TO PARSE DATABASE URL....")
	cf, err := pgxpool.ParseConfig(url)
	if err != nil {
		panic("couldn't parse the database config! err : " + err.Error())
	}

	cf.MaxConns = 15
	cf.MinConns = 3
	cf.MaxConnIdleTime = 20 * time.Minute
	cf.MaxConnLifetime = 10 * time.Minute

	println("LOG : TRYING TO CREATE DATABASE....")
	pool, err := pgxpool.NewWithConfig(ctx, cf)
	if err != nil {
		panic("couldn't create the database err : " + err.Error())
	}
	query := generated.New(pool)
	return &Database{
		Pool:  pool,
		Query: query,
	}
}
