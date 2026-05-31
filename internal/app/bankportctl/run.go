package bankportctl

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/config"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/store"
)

var ErrUsage = errors.New("invalid bankportctl usage")

type App struct {
	Config     config.Config
	Repository *store.Repository
	Stdout     io.Writer
	Stderr     io.Writer
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		return err
	}
	return App{
		Config:     cfg,
		Repository: store.NewSeededRepository(cfg),
		Stdout:     stdout,
		Stderr:     stderr,
	}.Run(ctx, args)
}

func (a App) Run(ctx context.Context, args []string) error {
	if a.Repository == nil {
		return errors.New("repository is required")
	}
	if a.Stdout == nil {
		a.Stdout = io.Discard
	}
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return a.usage()
	}

	switch args[0] {
	case "apps":
		return a.apps(ctx, args[1:])
	case "rate-limits":
		return a.rateLimits(ctx, args[1:])
	case "usage":
		return a.usageReport(ctx, args[1:])
	default:
		return fmt.Errorf("%w: unknown command %q", ErrUsage, args[0])
	}
}

func (a App) apps(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "list" {
		return fmt.Errorf("%w: expected apps list", ErrUsage)
	}
	format, err := parseFormat("apps list", args[1:])
	if err != nil {
		return err
	}
	apps, err := a.Repository.PartnerApps(ctx)
	if err != nil {
		return err
	}
	if format == "json" {
		return writeJSON(a.Stdout, apps)
	}

	table := tabwriter.NewWriter(a.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(table, "DEVELOPER_APP_ID\tPARTNER_ID\tRATE_LIMIT_RPM\tSCOPES")
	for _, app := range apps {
		_, _ = fmt.Fprintf(table, "%s\t%s\t%d\t%s\n", app.DeveloperAppID, app.PartnerID, app.RateLimitPerMinute, strings.Join(app.Scopes, ","))
	}
	return table.Flush()
}

func (a App) rateLimits(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "inspect" {
		return fmt.Errorf("%w: expected rate-limits inspect", ErrUsage)
	}
	format, err := parseFormat("rate-limits inspect", args[1:])
	if err != nil {
		return err
	}
	policies, err := a.Repository.RateLimitPolicies(ctx)
	if err != nil {
		return err
	}
	if format == "json" {
		return writeJSON(a.Stdout, policies)
	}

	table := tabwriter.NewWriter(a.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(table, "DEVELOPER_APP_ID\tLIMIT_PER_MINUTE\tPARTITION\tBACKING")
	for _, policy := range policies {
		_, _ = fmt.Fprintf(table, "%s\t%d\t%s\t%s\n", policy.DeveloperAppID, policy.LimitPerMinute, policy.PartitionStrategy, policy.DistributedBacking)
	}
	return table.Flush()
}

func (a App) usageReport(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "report" {
		return fmt.Errorf("%w: expected usage report", ErrUsage)
	}
	format, err := parseFormat("usage report", args[1:])
	if err != nil {
		return err
	}
	report, err := a.Repository.UsageReport(ctx)
	if err != nil {
		return err
	}
	if format == "json" {
		return writeJSON(a.Stdout, report)
	}

	table := tabwriter.NewWriter(a.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(table, "METRIC\tVALUE")
	_, _ = fmt.Fprintf(table, "partners\t%d\n", report.PartnerCount)
	_, _ = fmt.Fprintf(table, "developer_apps\t%d\n", report.DeveloperAppCount)
	_, _ = fmt.Fprintf(table, "accounts\t%d\n", report.AccountCount)
	_, _ = fmt.Fprintf(table, "pix_transfers\t%d\n", report.PixTransferCount)
	_, _ = fmt.Fprintf(table, "payouts\t%d\n", report.PayoutCount)
	_, _ = fmt.Fprintf(table, "refunds\t%d\n", report.RefundCount)
	_, _ = fmt.Fprintf(table, "events\t%d\n", report.EventCount)
	_, _ = fmt.Fprintf(table, "webhook_endpoints\t%d\n", report.WebhookEndpointCount)
	_, _ = fmt.Fprintf(table, "webhook_deliveries\t%d\n", report.WebhookDeliveryCount)
	_, _ = fmt.Fprintf(table, "audit_entries\t%d\n", report.AuditEntryCount)
	return table.Flush()
}

func (a App) usage() error {
	_, _ = fmt.Fprintln(a.Stdout, `bankportctl manages the local BankPort sandbox.

Usage:
  bankportctl apps list [--format table|json]
  bankportctl rate-limits inspect [--format table|json]
  bankportctl usage report [--format table|json]

Commands that require durable state, such as webhook replay and token rotation,
are intentionally not implemented in the in-memory sandbox CLI.`)
	return nil
}

func parseFormat(name string, args []string) (string, error) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	format := flags.String("format", "table", "output format: table or json")
	if err := flags.Parse(args); err != nil {
		return "", fmt.Errorf("%w: %v", ErrUsage, err)
	}
	if flags.NArg() != 0 {
		return "", fmt.Errorf("%w: unexpected argument %q", ErrUsage, flags.Arg(0))
	}
	switch *format {
	case "table", "json":
		return *format, nil
	default:
		return "", fmt.Errorf("%w: unsupported format %q", ErrUsage, *format)
	}
}

func writeJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
