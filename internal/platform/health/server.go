package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type Checker func(ctx context.Context) error

type NamedChecker struct {
	Name    string
	Checker Checker
}

type Server struct {
	addr     string
	checkers []NamedChecker
	logger   *zerolog.Logger
}

func New(addr string, checkers []NamedChecker, logger *zerolog.Logger) *Server {
	return &Server{
		addr:     addr,
		checkers: checkers,
		logger:   logger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleLiveness)
	mux.HandleFunc("/readyz", s.handleReadiness)

	srv := &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info().Str("addr", s.addr).Msg("health server started")
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.logger.Info().Msg("health server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Error().Err(err).Msg("health server shutdown failed")
			return err
		}
		s.logger.Info().Msg("health server stopped")
		return nil
	}
}

func (s *Server) handleLiveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	type result struct {
		name string
		err  error
	}

	results := make([]result, len(s.checkers))
	eg, ctx := errgroup.WithContext(ctx)
	for i, nc := range s.checkers {
		eg.Go(func() error {
			results[i] = result{name: nc.Name, err: nc.Checker(ctx)}
			return nil
		})
	}
	eg.Wait()

	checks := make(map[string]string, len(results))
	allOK := true
	for _, res := range results {
		if res.err != nil {
			checks[res.name] = res.err.Error()
			allOK = false
		} else {
			checks[res.name] = "ok"
		}
	}

	status := http.StatusOK
	if !allOK {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, map[string]any{"ok": allOK, "checks": checks})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
