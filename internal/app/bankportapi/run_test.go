package bankportapi

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunRejectsInvalidConfigBeforeListen(t *testing.T) {
	t.Setenv("BANKPORT_ENV", "production")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run(context.Background(), &stdout, &stderr)

	if err == nil {
		t.Fatal("expected invalid production config to fail before listen")
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no normal startup logs, got %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "bankport_api_invalid_config") {
		t.Fatalf("expected invalid config startup log, got %s", stderr.String())
	}
}
