package main

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb/http"
	"github.com/influxdata/influxdb/kit/signals"
	"github.com/spf13/cobra"
)

var deleteFlags http.DeleteRequest

func cmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete points from an influxDB bucket",
		Short: "Delete points from influxDB",
		Long: `Delete points from influxDB, by specify start, end time
	and a sql like predicate string.`,
		RunE: wrapCheckSetup(fluxDeleteF),
	}

	opts := flagOpts{
		{
			DestP:      &deleteFlags.OrgID,
			Flag:       "org-id",
			Desc:       "The ID of the organization that owns the bucket",
			Persistent: true,
		},
		{
			DestP:      &deleteFlags.Org,
			Flag:       "org",
			Short:      'o',
			Desc:       "The name of the organization that owns the bucket",
			Persistent: true,
		},
		{
			DestP:      &deleteFlags.BucketID,
			Flag:       "bucket-id",
			Desc:       "The ID of the destination bucket",
			Persistent: true,
		},
		{
			DestP:      &deleteFlags.Bucket,
			Flag:       "bucket",
			Desc:       "The name of destination bucket",
			EnvVar:     "BUCKET_NAME",
			Persistent: true,
		},
	}
	opts.mustRegister(cmd)

	cmd.PersistentFlags().StringVar(&deleteFlags.Start, "start", "", "the start time in RFC3339Nano format, exp 2009-01-02T23:00:00Z")
	cmd.PersistentFlags().StringVar(&deleteFlags.Stop, "stop", "", "the stop time in RFC3339Nano format, exp 2009-01-02T23:00:00Z")
	cmd.PersistentFlags().StringVarP(&deleteFlags.Predicate, "predicate", "p", "", "sql like predicate string, exp 'tag1=\"v1\" and (tag2=123)'")

	return cmd
}

func fluxDeleteF(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if deleteFlags.Org == "" && deleteFlags.OrgID == "" {
		return fmt.Errorf("please specify one of org or org-id")
	}

	if deleteFlags.Bucket == "" && deleteFlags.BucketID == "" {
		return fmt.Errorf("please specify one of bucket or bucket-id")
	}

	if deleteFlags.Start == "" || deleteFlags.Stop == "" {
		return fmt.Errorf("both start and stop are required")
	}

	s := &http.DeleteService{
		Addr:               flags.host,
		Token:              flags.token,
		InsecureSkipVerify: flags.skipVerify,
	}

	ctx = signals.WithStandardSignals(ctx)
	if err := s.DeleteBucketRangePredicate(ctx, deleteFlags); err != nil && err != context.Canceled {
		return fmt.Errorf("failed to delete data: %v", err)
	}

	return nil
}
