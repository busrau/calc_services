// Package httpapi exposes calculator operations over HTTP.
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/busrau/calc_services/internal/calculator"
)

const maxBodyBytes = 1 << 20 // 1 MiB

type Calculator interface {
	Add(a, b float64) (float64, error)
	Subtract(a, b float64) (float64, error)
	Multiply(a, b float64) (float64, error)
	Divide(a, b float64) (float64, error)
	Power(base, exponent float64) (float64, error)
	Sqrt(a float64) (float64, error)
	Percentage(value, percent float64) (float64, error)
}

type Server struct {
	calc Calculator
}

func NewServer(calc Calculator) *Server {
	return &Server{calc: calc}
}

func NewMux(calc Calculator) http.Handler {
	s := NewServer(calc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /v1/add", s.binary("add", s.calc.Add))
	mux.HandleFunc("POST /v1/subtract", s.binary("subtract", s.calc.Subtract))
	mux.HandleFunc("POST /v1/multiply", s.binary("multiply", s.calc.Multiply))
	mux.HandleFunc("POST /v1/divide", s.binary("divide", s.calc.Divide))
	mux.HandleFunc("POST /v1/power", s.binary("power", s.calc.Power))
	mux.HandleFunc("POST /v1/sqrt", s.unary("sqrt", s.calc.Sqrt))
	mux.HandleFunc("POST /v1/percentage", s.percentage)
	return mux
}

type binaryRequest struct {
	A *float64 `json:"a"`
	B *float64 `json:"b"`
}

type unaryRequest struct {
	A *float64 `json:"a"`
}

type percentageRequest struct {
	Value   *float64 `json:"value"`
	Percent *float64 `json:"percent"`
}

type successResponse struct {
	Operation string    `json:"operation"`
	Operands  []float64 `json:"operands"`
	Result    float64   `json:"result"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) binary(op string, fn func(float64, float64) (float64, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req binaryRequest
		if err := decodeJSON(w, r, &req); err != nil {
			return
		}
		if req.A == nil || req.B == nil {
			writeError(w, http.StatusBadRequest, "fields a and b are required")
			return
		}
		result, err := fn(*req.A, *req.B)
		if err != nil {
			writeCalcError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, successResponse{
			Operation: op,
			Operands:  []float64{*req.A, *req.B},
			Result:    result,
		})
	}
}

func (s *Server) unary(op string, fn func(float64) (float64, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req unaryRequest
		if err := decodeJSON(w, r, &req); err != nil {
			return
		}
		if req.A == nil {
			writeError(w, http.StatusBadRequest, "field a is required")
			return
		}
		result, err := fn(*req.A)
		if err != nil {
			writeCalcError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, successResponse{
			Operation: op,
			Operands:  []float64{*req.A},
			Result:    result,
		})
	}
}

func (s *Server) percentage(w http.ResponseWriter, r *http.Request) {
	var req percentageRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	if req.Value == nil || req.Percent == nil {
		writeError(w, http.StatusBadRequest, "fields value and percent are required")
		return
	}
	result, err := s.calc.Percentage(*req.Value, *req.Percent)
	if err != nil {
		writeCalcError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, successResponse{
		Operation: "percentage",
		Operands:  []float64{*req.Value, *req.Percent},
		Result:    result,
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if r.Body == nil {
		writeError(w, http.StatusBadRequest, "request body is required")
		return errors.New("empty body")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return err
	}
	if dec.More() {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return errors.New("extra json")
	}
	return nil
}

func writeCalcError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, calculator.ErrDivisionByZero),
		errors.Is(err, calculator.ErrNegativeSqrt),
		errors.Is(err, calculator.ErrInvalidNumber),
		errors.Is(err, calculator.ErrInvalidOperation),
		errors.Is(err, calculator.ErrResultOutOfRange):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
