package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestArithmetic(t *testing.T) {
	tests := []struct{ route, a, b, result string }{
		{"sum", "1", "2", "3"}, {"sum", "-5", "2", "-3"},
		{"sub", "5", "2", "3"}, {"sub", "-5", "-2", "-3"},
		{"mul", "3", "2", "6"}, {"mul", "-5", "2", "-10"},
		{"div", "5", "2", "2"}, {"div", "-5", "2", "-2"},
		{"div", "5", "-2", "-2"}, {"div", "-5", "-2", "2"},
		{"div", "0", "-2", "0"}, {"div", "-1", "2", "0"},
		{"sum", "9223372036854775807", "1", "9223372036854775808"},
	}
	handler := newHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	for _, tt := range tests {
		t.Run(tt.route+"/"+tt.a+"/"+tt.b, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/"+tt.route+"?term_one="+tt.a+"&term_two="+tt.b, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != 200 || strings.TrimSpace(rec.Body.String()) != `{"result":`+tt.result+`}` {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestValidation(t *testing.T) {
	handler := newHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	for _, route := range []string{"sum", "sub", "mul", "div"} {
		for _, query := range []string{"", "term_one=1", "term_two=2", "term_one=abc&term_two=2", "term_one=1&term_two=1.5", "term_one=&term_two=1", "term_one=1&term_one=2&term_two=3"} {
			t.Run(route+"/"+query, func(t *testing.T) {
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, httptest.NewRequest("GET", "/api/"+route+"?"+query, nil))
				if rec.Code != 422 {
					t.Fatalf("status=%d", rec.Code)
				}
			})
		}
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/api/div?term_one=1&term_two=0", nil))
	if rec.Code != 400 {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestRoutesAndLogs(t *testing.T) {
	tests := []struct {
		method, path       string
		status             int
		route, methodLabel string
	}{
		{"GET", "/", 200, "/", "GET"},
		{"GET", "/healthz", 200, "/healthz", "GET"},
		{"GET", "/readyz", 200, "/readyz", "GET"},
		{"GET", "/version", 200, "/version", "GET"},
		{"GET", "/private-secret?term_one=987654321", 404, "unmatched", "GET"},
		{"POST", "/api/sum", 405, "/api/sum", "POST"},
		{"CUSTOM", "/api/sum", 405, "/api/sum", "OTHER"},
		{"GET", "/api/sum?term_one=987654321&term_two=1", 200, "/api/sum", "GET"},
	}
	for _, tt := range tests {
		t.Run(tt.method+tt.path, func(t *testing.T) {
			var logs bytes.Buffer
			handler := newHandler(slog.New(slog.NewJSONHandler(&logs, nil)))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
			if rec.Code != tt.status {
				t.Fatalf("status=%d", rec.Code)
			}
			if tt.path == "/" && !strings.Contains(rec.Body.String(), "Vamos calcular?") {
				t.Fatal("HTML ausente")
			}
			if tt.path == "/healthz" || tt.path == "/readyz" {
				if strings.TrimSpace(rec.Body.String()) != `{"status":"ok"}` {
					t.Fatal(rec.Body.String())
				}
			}
			if tt.path == "/version" && !strings.Contains(rec.Body.String(), version) {
				t.Fatal("versão ausente")
			}
			if tt.status == 405 && rec.Header().Get("Allow") != "GET" {
				t.Fatal("Allow ausente")
			}
			var record map[string]any
			if err := json.Unmarshal(logs.Bytes(), &record); err != nil {
				t.Fatal(err)
			}
			if record["route"] != tt.route || record["method"] != tt.methodLabel || record["status"] != float64(tt.status) {
				t.Fatal(record)
			}
			if strings.Contains(logs.String(), "987654321") || strings.Contains(logs.String(), "private-secret") {
				t.Fatal("log expõe valores privados")
			}
			if record["duration_seconds"].(float64) < 0 {
				t.Fatal("duração negativa")
			}
		})
	}
}
