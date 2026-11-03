package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	dbSource := os.Getenv("DB_SOURCE")

	conn, err := pgxpool.New(ctx, dbSource)
	if err != nil {
		log.Fatal("cannot connect db:", err)
	}
	defer conn.Close()

	// Đọc file seed.sql
	content, err := os.ReadFile("./scripts/seed/seed.sql")
	if err != nil {
		log.Fatal("cannot read seed.sql:", err)
	}

	// Thực thi trực tiếp
	_, err = conn.Exec(ctx, string(content))
	if err != nil {
		log.Fatal("failed to execute seed.sql:", err)
	}

	fmt.Println("✅ Seed.sql executed successfully")
}
