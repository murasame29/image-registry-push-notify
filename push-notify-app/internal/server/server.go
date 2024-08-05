package server

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/murasame29/image-registry-push-notify/sample-app/internal/log"
	"golang.org/x/sync/errgroup"
)

const DefaultShutdownTimeout time.Duration = 10

// Server はHTTPサーバーを表します。
type Server struct {
	port            int
	host            string
	shutdownTimeout time.Duration

	srv *http.Server
}

// New はサーバーを生成します。
func New(handler http.Handler, opts ...Option) *Server {
	server := &Server{
		port:            8080,
		host:            "localhost",
		shutdownTimeout: DefaultShutdownTimeout,
		srv:             new(http.Server),
	}

	for _, opt := range opts {
		opt(server)
	}

	server.srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", server.host, server.port),
		Handler: handler,
	}

	return server
}

// Run はサーバーを起動します。
func (s *Server) Run(ctx context.Context) error {
	log.Info(ctx, "server starting. %s", s.srv.Addr)
	return s.srv.ListenAndServe()
}

// Shutdown はサーバーを停止します。
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info(ctx, "server shutdown ...")
	return s.srv.Shutdown(ctx)
}

// RunWithGracefulShutdown はgraceful shutdownを行うサーバーを起動します。
func (s *Server) RunWithGracefulShutdown(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGALRM)
	defer stop()

	errWg, errCtx := errgroup.WithContext(ctx)
	errWg.Go(func() error {
		if err := s.Run(ctx); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("listen And Serve error : %s", err.Error())
		}

		return nil
	})

	errWg.Go(func() error {
		<-errCtx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		return s.Shutdown(ctx)
	})

	err := errWg.Wait()

	if err != context.Canceled && err != nil {
		log.Error(ctx, err.Error())
		return err
	}

	log.Info(ctx, "server shutdown completed")
	return nil
}
