package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func DBClient(ctx context.Context) *pgx.Conn {
    fmt.Println(os.Getenv("DATABASE_URL"))
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database🤦‍♂️: %v\n", err)
		os.Exit(1)
	}
	return conn
}