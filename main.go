package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed web/index.html
var calculator string

const version = "1.0.1"

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// As rotas são explícitas para que caminhos arbitrários nunca entrem nos logs.
func newHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		status := http.StatusOK
		route := "unmatched"
		switch r.URL.Path {
		case "/", "/api/sum", "/api/sub", "/api/mul", "/api/div", "/healthz", "/readyz", "/version":
			route = r.URL.Path
		}
		method := r.Method
		switch method {
		case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE", "CONNECT":
		default:
			method = "OTHER"
		}
		defer func() {
			logger.Info("request", "method", method, "route", route, "status", status,
				"duration_seconds", time.Since(started).Seconds())
		}()
		fail := func(code int, detail string) {
			status = code
			writeJSON(w, code, map[string]string{"detail": detail})
		}
		if route == "unmatched" {
			fail(http.StatusNotFound, "Rota não encontrada")
			return
		}
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			fail(http.StatusMethodNotAllowed, "Método não permitido")
			return
		}
		switch route {
		case "/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.WriteString(w, calculator)
		case "/healthz", "/readyz":
			writeJSON(w, status, map[string]string{"status": "ok"})
		case "/version":
			writeJSON(w, status, map[string]string{"version": version})
		default:
			// big.Int evita overflow e conserva o resultado como número inteiro no JSON.
			terms := make([]*big.Int, 2)
			for i, name := range []string{"term_one", "term_two"} {
				values := r.URL.Query()[name]
				if len(values) != 1 {
					fail(http.StatusUnprocessableEntity, "Informe uma vez cada parâmetro: term_one e term_two")
					return
				}
				value, ok := new(big.Int).SetString(values[0], 10)
				if !ok {
					fail(http.StatusUnprocessableEntity, "Os parâmetros devem ser números inteiros")
					return
				}
				terms[i] = value
			}
			a, b := terms[0], terms[1]
			result := new(big.Int)
			switch route {
			case "/api/sum":
				result.Add(a, b)
			case "/api/sub":
				result.Sub(a, b)
			case "/api/mul":
				result.Mul(a, b)
			case "/api/div":
				if b.Sign() == 0 {
					fail(http.StatusBadRequest, "Divisão por zero")
					return
				}
				// Quo trunca em direção a zero, inclusive para negativos.
				result.Quo(a, b)
			}
			writeJSON(w, status, map[string]*big.Int{"result": result})
		}
	})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	server := &http.Server{
		Addr: ":8000", Handler: newHandler(logger),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown", "error", err.Error())
		}
	}()
	logger.Info("starting", "port", 8000, "version", version)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server", "error", err.Error())
		os.Exit(1)
	}
}
