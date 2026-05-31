package main

import (
	"context"
	"os"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/app/bankportapi"
)

func main() {
	if err := bankportapi.Run(context.Background(), os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}
