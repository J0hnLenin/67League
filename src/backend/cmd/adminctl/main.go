package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/J0hnLenin/67League/internal/domain/admin"
	"github.com/J0hnLenin/67League/internal/infrastructure/auth"
	"github.com/J0hnLenin/67League/internal/infrastructure/persistence/postgres"
)

func main() {
	login := flag.String("login", "", "admin login")
	password := flag.String("password", "", "admin password")
	dsn := flag.String("dsn", os.Getenv("LIGA_DB_DSN"), "database dsn")
	flag.Parse()

	if *login == "" || *password == "" || *dsn == "" {
		fmt.Fprintln(os.Stderr, "usage: adminctl -login X -password Y [-dsn ...]")
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := postgres.NewDB(ctx, *dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	hasher := auth.NewPasswordHasher(0)
	hash, err := hasher.Hash(*password)
	if err != nil {
		panic(err)
	}

	a, err := admin.New(*login, hash, timeNowUTC())
	if err != nil {
		panic(err)
	}

	repo := postgres.NewAdminRepo(db.Pool())
	if err := repo.Create(ctx, a); err != nil {
		fmt.Fprintln(os.Stderr, "create admin:", err)
		os.Exit(1)
	}
	fmt.Printf("created admin id=%d login=%s\n", a.ID, a.Login)
}

func timeNowUTC() time.Time { return time.Now().UTC() }
