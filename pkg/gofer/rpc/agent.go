package rpc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/rpc"

	"github.com/toknowwhy/theunit-oracle/pkg/gofer"
	"github.com/toknowwhy/theunit-oracle/pkg/log"
)

const AgentLoggerTag = "GOFER_AGENT"

type AgentConfig struct {
	// Gofer instance which will be used by the agent. If this instance
	// implements the gofer.StartableGofer interface, the Start and Stop
	// methods are called whenever corresponding Agent's Start and
	// Stop are called.
	Gofer gofer.Gofer
	// Network is used for the rpc.Listener function.
	Network string
	// Address is used for the rpc.Listener function.
	Address string
	Logger  log.Logger
}

// Agent creates and manages an RPC server for remote Gofer calls.
type Agent struct {
	ctx    context.Context
	doneCh chan struct{}

	api      *API
	rpc      *rpc.Server
	listener net.Listener
	gofer    gofer.Gofer
	network  string
	address  string
	log      log.Logger
}

// NewAgent returns a new Agent instance.
func NewAgent(ctx context.Context, cfg AgentConfig) (*Agent, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	server := &Agent{
		ctx:    ctx,
		doneCh: make(chan struct{}),
		api: &API{
			gofer: cfg.Gofer,
			log:   cfg.Logger.WithField("tag", AgentLoggerTag),
		},
		rpc:     rpc.NewServer(),
		gofer:   cfg.Gofer,
		network: cfg.Network,
		address: cfg.Address,
		log:     cfg.Logger.WithField("tag", AgentLoggerTag),
	}

	err := server.rpc.Register(server.api)
	if err != nil {
		return nil, err
	}
	server.rpc.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	return server, nil
}

// Start starts the RPC server.
func (s *Agent) Start() error {
	s.log.Infof("Starting")
	var err error

	// Start Gofer if necessary:
	if sg, ok := s.gofer.(gofer.StartableGofer); ok {
		err = sg.Start()
		if err != nil {
			return err
		}
	}

	// Start RPC server:
	s.listener, err = net.Listen(s.network, s.address)
	if err != nil {
		return err
	}
	go func() {
		err := http.Serve(s.listener, nil)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.WithError(err).Error("RPC server crashed")
		}
	}()

	go s.contextCancelHandler()
	return nil
}

// Wait waits until agent's context is cancelled.
func (s *Agent) Wait() {
	<-s.doneCh
}

func (s *Agent) contextCancelHandler() {
	defer func() { close(s.doneCh) }()
	defer s.log.Info("Stopped")
	<-s.ctx.Done()

	err := s.listener.Close()
	if err != nil {
		s.log.WithError(err).Error("Unable to close RPC listener")
	}
}
