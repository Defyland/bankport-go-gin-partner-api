package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/app/bankportctl"
)

func main() {
	if err := bankportctl.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, bankportctl.ErrUsage) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
