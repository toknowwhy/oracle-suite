package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/toknowwhy/theunit-oracle/internal/config"
	ethereumConfig "github.com/toknowwhy/theunit-oracle/internal/config/ethereum"
	feedsConfig "github.com/toknowwhy/theunit-oracle/internal/config/feeds"
	spectreConfig "github.com/toknowwhy/theunit-oracle/internal/config/spectre"
	transportConfig "github.com/toknowwhy/theunit-oracle/internal/config/transport"
	"github.com/toknowwhy/theunit-oracle/pkg/datastore"
	"github.com/toknowwhy/theunit-oracle/pkg/log"
	logLogrus "github.com/toknowwhy/theunit-oracle/pkg/log/logrus"
	"github.com/toknowwhy/theunit-oracle/pkg/spectre"
	"github.com/toknowwhy/theunit-oracle/pkg/transport"
)

type Config struct {
	Transport transportConfig.Transport `json:"transport"`
	Ethereum  ethereumConfig.Ethereum   `json:"ethereum"`
	Spectre   spectreConfig.Spectre     `json:"spectre"`
	Feeds     feedsConfig.Feeds         `json:"feeds"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) Configure(d Dependencies) (transport.Transport, datastore.Datastore, *spectre.Spectre, error) {
	sig, err := c.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, nil, nil, err
	}
	cli, err := c.Ethereum.ConfigureEthereumClient(sig)
	if err != nil {
		return nil, nil, nil, err
	}
	fed, err := c.Feeds.Addresses()
	if err != nil {
		return nil, nil, nil, err
	}
	tra, err := c.Transport.Configure(transportConfig.Dependencies{
		Context: d.Context,
		Signer:  sig,
		Feeds:   fed,
		Logger:  d.Logger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	dat, err := c.Spectre.ConfigureDatastore(spectreConfig.DatastoreDependencies{
		Context:   d.Context,
		Signer:    sig,
		Transport: tra,
		Feeds:     fed,
		Logger:    d.Logger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	spe, err := c.Spectre.ConfigureSpectre(spectreConfig.Dependencies{
		Context:        d.Context,
		Signer:         sig,
		Datastore:      dat,
		EthereumClient: cli,
		Logger:         d.Logger,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return tra, dat, spe, nil
}

type Services struct {
	ctxCancel context.CancelFunc
	Transport transport.Transport
	Datastore datastore.Datastore
	Spectre   *spectre.Spectre
}

func PrepareServices(ctx context.Context, opts *options) (*Services, error) {
	var err error
	ctx, ctxCancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			ctxCancel()
		}
	}()

	// Load config file:
	err = config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// Logger:
	ll, err := logrus.ParseLevel(opts.LogVerbosity)
	if err != nil {
		return nil, err
	}
	lr := logrus.New()
	lr.SetLevel(ll)
	lr.SetFormatter(opts.LogFormat.Formatter())
	logger := logLogrus.New(lr)

	// Services:
	tra, dat, spe, err := opts.Config.Configure(Dependencies{
		Context: ctx,
		Logger:  logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load Spectre configuration: %w", err)
	}

	return &Services{
		ctxCancel: ctxCancel,
		Transport: tra,
		Datastore: dat,
		Spectre:   spe,
	}, nil
}

func (s *Services) Start() error {
	var err error
	if err = s.Transport.Start(); err != nil {
		return err
	}
	if err = s.Datastore.Start(); err != nil {
		return err
	}
	if err = s.Spectre.Start(); err != nil {
		return err
	}
	return nil
}

func (s *Services) CancelAndWait() {
	s.ctxCancel()
	s.Transport.Wait()
	s.Datastore.Wait()
	s.Spectre.Wait()
}
