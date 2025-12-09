package database

import (
	"context"
	"database/sql"
	// "log"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil { return nil, err }
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Minute * 5)
	if err := db.Ping(); err != nil { return nil, err }
	return db, nil
}

func ConnectMongo(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil { return nil, err }
	if err := client.Ping(ctx, nil); err != nil { return nil, err }
	return client, nil
}
