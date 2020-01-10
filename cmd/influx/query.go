package main

import (
	"context"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/repl"
	_ "github.com/influxdata/flux/stdlib"
	platform "github.com/influxdata/influxdb"
	_ "github.com/influxdata/influxdb/query/stdlib"
	"github.com/spf13/cobra"
)

var queryFlags struct {
	OrgID string
	Org   string
}

func cmdQuery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query [query literal or @/path/to/query.flux]",
		Short: "Execute a Flux query",
		Long: `Execute a literal Flux query provided as a string,
or execute a literal Flux query contained in a file by specifying the file prefixed with an @ sign.`,
		Args: cobra.ExactArgs(1),
		RunE: wrapCheckSetup(fluxQueryF),
	}

	opts := flagOpts{
		{
			DestP:      &queryFlags.OrgID,
			Flag:       "org-id",
			Desc:       "The organization ID",
			Persistent: true,
		},
		{
			DestP:      &queryFlags.Org,
			Flag:       "org",
			Short:      'o',
			Desc:       "The organization name",
			Persistent: true,
		},
	}
	opts.mustRegister(cmd)

	return cmd
}

func fluxQueryF(cmd *cobra.Command, args []string) error {
	if flags.local {
		return fmt.Errorf("local flag not supported for query command")
	}

	if (queryFlags.OrgID != "" && queryFlags.Org != "") || (queryFlags.OrgID == "" && queryFlags.Org == "") {
		return fmt.Errorf("must specify exactly one of org or org-id")
	}

	q, err := repl.LoadQuery(args[0])
	if err != nil {
		return fmt.Errorf("failed to load query: %v", err)
	}

	var orgID platform.ID

	if queryFlags.OrgID != "" {
		if err := orgID.DecodeFromString(queryFlags.OrgID); err != nil {
			return fmt.Errorf("failed to decode org-id: %v", err)
		}
	}

	if queryFlags.Org != "" {
		orgSvc, err := newOrganizationService()
		if err != nil {
			return fmt.Errorf("failed to initialized organization service client: %v", err)
		}

		filter := platform.OrganizationFilter{Name: &queryFlags.Org}
		o, err := orgSvc.FindOrganization(context.Background(), filter)
		if err != nil {
			return fmt.Errorf("failed to retrieve organization %q: %v", queryFlags.Org, err)
		}

		orgID = o.ID
	}

	flux.FinalizeBuiltIns()

	r, err := getFluxREPL(flags.host, flags.token, flags.skipVerify, orgID)
	if err != nil {
		return fmt.Errorf("failed to get the flux REPL: %v", err)
	}

	if err := r.Input(q); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}

	return nil
}
