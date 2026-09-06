package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/busrau/calc_services/internal/calculator"
)

func TestHealth(t *testing.T) {
	mux := NewMux(calculator.New())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestBinaryOperations(t *testing.T) {
	mux := NewMux(calculator.New())
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
		wantResult float64
		wantError  string
	}{
		{name: "add", path: "/v1/add", body: `{"a":2,"b":3}`, wantStatus: http.StatusOK, wantResult: 5},
		{name: "subtract", path: "/v1/subtract", body: `{"a":10,"b":4}`, wantStatus: http.StatusOK, wantResult: 6},
		{name: "multiply", path: "/v1/multiply", body: `{"a":3,"b":7}`, wantStatus: http.StatusOK, wantResult: 21},
		{name: "divide", path: "/v1/divide", body: `{"a":9,"b":3}`, wantStatus: http.StatusOK, wantResult: 3},
		{name: "power", path: "/v1/power", body: `{"a":2,"b":10}`, wantStatus: http.StatusOK, wantResult: 1024},
		{name: "zero is valid", path: "/v1/add", body: `{"a":0,"b":0}`, wantStatus: http.StatusOK, wantResult: 0},
		{name: "division by zero", path: "/v1/divide", body: `{"a":1,"b":0}`, wantStatus: http.StatusBadRequest, wantError: calculator.ErrDivisionByZero.Error()},
		{name: "missing field", path: "/v1/add", body: `{"a":1}`, wantStatus: http.StatusBadRequest, wantError: "fields a and b are required"},
		{name: "invalid json", path: "/v1/add", body: `{`, wantStatus: http.StatusBadRequest, wantError: "invalid JSON request body"},
		{name: "unknown field", path: "/v1/add", body: `{"a":1,"b":2,"c":3}`, wantStatus: http.StatusBadRequest, wantError: "invalid JSON request body"},
		{name: "wrong types", path: "/v1/add", body: `{"a":"x","b":1}`, wantStatus: http.StatusBadRequest, wantError: "invalid JSON request body"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			assertJSONBody(t, rec, tt.wantStatus, tt.wantResult, tt.wantError)
		})
	}
}

func TestSqrtAndPercentage(t *testing.T) {
	mux := NewMux(calculator.New())

	t.Run("sqrt", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/sqrt", strings.NewReader(`{"a":81}`))
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}
		assertJSONBody(t, rec, http.StatusOK, 9, "")
	})

	t.Run("negative sqrt", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/sqrt", strings.NewReader(`{"a":-4}`))
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		assertJSONBody(t, rec, http.StatusBadRequest, 0, calculator.ErrNegativeSqrt.Error())
	})

	t.Run("sqrt missing a", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/sqrt", strings.NewReader(`{}`))
		mux.ServeHTTP(rec, req)
		assertJSONBody(t, rec, http.StatusBadRequest, 0, "field a is required")
	})

	t.Run("percentage", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/percentage", strings.NewReader(`{"value":80,"percent":25}`))
		mux.ServeHTTP(rec, req)
		assertJSONBody(t, rec, http.StatusOK, 20, "")
	})

	t.Run("percentage missing fields", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/percentage", strings.NewReader(`{"value":80}`))
		mux.ServeHTTP(rec, req)
		assertJSONBody(t, rec, http.StatusBadRequest, 0, "fields value and percent are required")
	})
}

func TestMethodNotAllowed(t *testing.T) {
	mux := NewMux(calculator.New())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/add", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestWriteCalcErrorInternal(t *testing.T) {
	mux := NewMux(stubCalc{err: errors.New("boom")})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/add", strings.NewReader(`{"a":1,"b":2}`))
	mux.ServeHTTP(rec, req)
	assertJSONBody(t, rec, http.StatusInternalServerError, 0, "internal server error")
}

type stubCalc struct {
	err error
}

func (s stubCalc) Add(float64, float64) (float64, error)        { return 0, s.err }
func (s stubCalc) Subtract(float64, float64) (float64, error)   { return 0, s.err }
func (s stubCalc) Multiply(float64, float64) (float64, error)   { return 0, s.err }
func (s stubCalc) Divide(float64, float64) (float64, error)     { return 0, s.err }
func (s stubCalc) Power(float64, float64) (float64, error)      { return 0, s.err }
func (s stubCalc) Sqrt(float64) (float64, error)                { return 0, s.err }
func (s stubCalc) Percentage(float64, float64) (float64, error) { return 0, s.err }

func assertJSONBody(t *testing.T, rec *httptest.ResponseRecorder, status int, wantResult float64, wantError string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, status, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}

	if wantError != "" {
		var resp errorResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			t.Fatalf("decode error response: %v", err)
		}
		if resp.Error != wantError {
			t.Fatalf("error = %q, want %q", resp.Error, wantError)
		}
		return
	}

	var resp successResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode success response: %v", err)
	}
	if resp.Result != wantResult {
		t.Fatalf("result = %v, want %v", resp.Result, wantResult)
	}
}
