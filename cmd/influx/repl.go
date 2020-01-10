package main

import (
	"context"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/repl"
	_ "github.com/influxdata/flux/stdlib"
	platform "github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/http"
	"github.com/influxdata/influxdb/query"
	_ "github.com/influxdata/influxdb/query/stdlib"
	"github.com/spf13/cobra"
)

var replFlags struct {
	organization
}

func cmdREPL() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repl",
		Short: "Interactive Flux REPL (read-eval-print-loop)",
		Args:  cobra.NoArgs,
		RunE:  wrapCheckSetup(replF),
	}
	replFlags.organization.register(cmd, false)

	return cmd
}

func replF(cmd *cobra.Command, args []string) error {
	if flags.local {
		return fmt.Errorf("local flag not supported for repl command")
	}

	if err := replFlags.organization.validOrgFlags(); err != nil {
		return err
	}

	orgSVC, err := newOrganizationService()
	if err != nil {
		return err
	}

	orgID, err := replFlags.organization.getID(orgSVC)
	if err != nil {
		return err
	}

	flux.FinalizeBuiltIns()

	r, err := getFluxREPL(flags.host, flags.token, flags.skipVerify, orgID)
	if err != nil {
		return err
	}

	r.Run()
	return nil
}

func getFluxREPL(addr, token string, skipVerify bool, orgID platform.ID) (*repl.REPL, error) {
	qs := &http.FluxQueryService{
		Addr:               addr,
		Token:              token,
		InsecureSkipVerify: skipVerify,
	}
	q := &query.REPLQuerier{
		OrganizationID: orgID,
		QueryService:   qs,
	}
	// background context is OK here, and DefaultDependencies are noop deps.  Also safe
	// since we send all queries to the server side.
	return repl.New(context.Background(), flux.NewDefaultDependencies(), q), nil
}
