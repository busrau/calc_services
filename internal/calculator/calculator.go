// Package calculator implements arithmetic operations used by the HTTP API.
package calculator

import (
	"errors"
	"math"
)

var (
	ErrDivisionByZero   = errors.New("division by zero")
	ErrNegativeSqrt     = errors.New("square root of a negative number")
	ErrInvalidNumber    = errors.New("operand must be a finite number")
	ErrInvalidOperation = errors.New("operation is undefined for the given operands")
	ErrResultOutOfRange = errors.New("result is out of range")
)

// Service performs calculator operations. It is safe for concurrent use.
type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Add(a, b float64) (float64, error) {
	return finiteResult(a+b, a, b)
}

func (s *Service) Subtract(a, b float64) (float64, error) {
	return finiteResult(a-b, a, b)
}

func (s *Service) Multiply(a, b float64) (float64, error) {
	return finiteResult(a*b, a, b)
}

func (s *Service) Divide(a, b float64) (float64, error) {
	if err := requireFinite(a, b); err != nil {
		return 0, err
	}
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return finiteResult(a / b)
}

// Power returns base raised to exponent (base^exponent).
func (s *Service) Power(base, exponent float64) (float64, error) {
	if err := requireFinite(base, exponent); err != nil {
		return 0, err
	}
	result := math.Pow(base, exponent)
	if math.IsNaN(result) {
		return 0, ErrInvalidOperation
	}
	return finiteResult(result)
}

func (s *Service) Sqrt(a float64) (float64, error) {
	if err := requireFinite(a); err != nil {
		return 0, err
	}
	if a < 0 {
		return 0, ErrNegativeSqrt
	}
	return finiteResult(math.Sqrt(a))
}

// Percentage returns percent% of value (value * percent / 100).
func (s *Service) Percentage(value, percent float64) (float64, error) {
	if err := requireFinite(value, percent); err != nil {
		return 0, err
	}
	return finiteResult(value * percent / 100)
}

func requireFinite(values ...float64) error {
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return ErrInvalidNumber
		}
	}
	return nil
}

func finiteResult(result float64, operands ...float64) (float64, error) {
	if err := requireFinite(operands...); err != nil {
		return 0, err
	}
	if math.IsNaN(result) {
		return 0, ErrInvalidOperation
	}
	if math.IsInf(result, 0) {
		return 0, ErrResultOutOfRange
	}
	return result, nil
}
