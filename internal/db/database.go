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

func (db *Database) Transaction(
	ctx context.Context,
	fn func(*generated.Queries) error,
) error {

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := db.Query.WithTx(tx)

	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
