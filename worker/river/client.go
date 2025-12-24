package riverCommon

import (
	"context"
	"time"

	DBclient "github.com/balobas/sport_city_common/clients/database"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
)

var driver riverdriver.Driver[pgx.Tx] = &riverpgxv5.Driver{}

type Config interface {
	Queues() map[string]int
	QueueNames() []string
	MaxAttempts() int
	MaxWorkers() int
	NextRetry() time.Duration
	JobTimeout() time.Duration
	FetchCooldown() time.Duration
	FetchPollInterval() time.Duration
}

type Client struct {
	r        *river.Client[pgx.Tx]
	w        *river.Workers
	cfg      Config
	dbClient DBclient.ClientDB
}

func NewClient(cfg Config, dbClient DBclient.ClientDB) (*Client, error) {
	driver := riverpgxv5.New(dbClient.GetMasterPool())
	workers := river.NewWorkers()

	c := &Client{
		w:        workers,
		cfg:      cfg,
		dbClient: dbClient,
	}

	queues := make(map[string]river.QueueConfig, len(cfg.Queues()))
	for name, maxWrkrs := range cfg.Queues() {
		queues[name] = river.QueueConfig{
			MaxWorkers: maxWrkrs,
		}
	}

	riverCfg := &river.Config{
		FetchCooldown:     cfg.FetchCooldown(),
		FetchPollInterval: cfg.FetchPollInterval(),
		JobTimeout:        cfg.JobTimeout(),
		MaxAttempts:       cfg.MaxAttempts(),
		Queues:            queues,
		RetryPolicy:       c,
		Workers:           c.w,
	}

	rClient, err := river.NewClient(driver, riverCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create river client")
	}

	c.r = rClient

	return c, nil
}

func (c *Client) NextRetry(job *rivertype.JobRow) time.Time {
	return time.Now().Add(c.cfg.NextRetry())
}

func (c *Client) Start(ctx context.Context) error {
	return c.r.Start(ctx)
}

func (c *Client) Stop(ctx context.Context) error {
	return c.r.Stop(ctx)
}

func AddWorker[T river.JobArgs](c *Client, worker river.Worker[T]) {
	river.AddWorker(c.w, worker)
}

func (c *Client) InsertRiver(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) error {
	tx, isInTx := c.dbClient.GetTxFromCtx(ctx)
	if isInTx {
		if _, err := c.r.InsertTx(ctx, tx, args, opts); err != nil {
			return errors.Wrap(err, "river.InsertTx()")
		}
		return nil
	}

	if _, err := c.r.Insert(ctx, args, opts); err != nil {
		return errors.Wrap(err, "river.Insert()")
	}
	return nil
}
