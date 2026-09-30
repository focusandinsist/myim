package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"myim/internal/migration"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	dsn := flag.String("dsn", os.Getenv("MYIM_DATABASE_DSN"), "PostgreSQL connection string (defaults to MYIM_DATABASE_DSN)")
	command := flag.String("command", "up", "migration command: up or down")
	steps := flag.Int("steps", 1, "number of migrations to roll back")
	timeout := flag.Duration("timeout", 30*time.Second, "connection, lock and migration timeout")
	flag.Parse()
	if flag.NArg() != 0 {
		return errors.New("unexpected positional arguments; use -command up or -command down")
	}
	if *dsn == "" {
		return errors.New("-dsn or MYIM_DATABASE_DSN is required")
	}
	if *timeout <= 0 {
		return errors.New("-timeout must be positive")
	}
	db, err := sql.Open("pgx", *dsn)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	switch *command {
	case "up":
		err = migration.Up(ctx, db)
	case "down":
		err = migration.Down(ctx, db, *steps)
	default:
		err = fmt.Errorf("unknown migration command %q", *command)
	}
	return err
}
