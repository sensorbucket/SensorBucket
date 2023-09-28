package main

import (
	"fmt"
	"net/url"
	"strconv"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/pkg/api"
)

type config struct {
	store            dbStore
	mailSender       mailer
	apiClient        *api.APIClient
	dataTimeout      int64
	daysTillDeletion int64
	errorThreshold   int64
	checkLastHours   int64
	fromEmail        string
	toEmail          string
}

func baseConfig() *config {
	return &config{}
}

func (c *config) withSensorbucketAndTracingDb() *config {
	sensorbucketDb, err := createDB(env.Must("DB_DSN_SENSORBUCKET"))
	if err != nil {
		panic(fmt.Errorf("create conn to sensorbucket db: %w", err))
	}
	tracingDb, err := createDB(env.Must("DB_DSN_TRACING"))
	if err != nil {
		panic(fmt.Errorf("create conn to sensorbucket db: %w", err))
	}
	c.store = dbStore{
		sensorbucketDb: sensorbucketDb,
		tracingDb:      tracingDb,
	}
	return c
}

func (c *config) withMailer() *config {
	c.mailSender = &emailSender{
		username: env.Must("SMTP_USERNAME"),
		password: env.Must("SMTP_PASSWORD"),
		host:     env.Must("SMTP_HOST"),
	}
	c.toEmail = env.Must("TO_EMAIL")
	c.fromEmail = env.Must("FROM_EMAIL")
	return c
}

func (c *config) withApiClient() *config {
	sbURL, err := url.Parse(env.Must("SB_API"))
	if err != nil {
		panic(fmt.Errorf("could not parse SB_API url: %w", err))
	}
	cfg := api.NewConfiguration()
	cfg.Scheme = sbURL.Scheme
	cfg.Host = sbURL.Host
	c.apiClient = api.NewAPIClient(cfg)
	return c
}

func (c *config) withDataTimeout() *config {
	val, err := strconv.ParseInt(env.Must("DATA_TIMEOUT"), 10, 64)
	if err != nil {
		panic(err)
	}
	c.dataTimeout = val
	return c
}

func (c *config) withDaysTillDeletionConfig() *config {
	val, err := strconv.ParseInt(env.Must("DAYS_TILL_DELETION"), 10, 64)
	if err != nil {
		panic(err)
	}
	c.daysTillDeletion = val
	return c
}

func (c *config) withErrorThreshold() *config {
	val, err := strconv.ParseInt(env.Could("ERROR_TRESHOLD", "3"), 10, 64)
	if err != nil {
		panic(err)
	}
	c.errorThreshold = val
	return c
}

func (c *config) withTimeframe() *config {
	val, err := strconv.ParseInt(env.Could("CHECK_LAST_X_HOURS", "12"), 10, 64)
	if err != nil {
		panic(err)
	}
	c.checkLastHours = val
	return c
}
