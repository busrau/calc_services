package calculator

import (
	"errors"
	"math"
	"testing"
)

func TestAdd(t *testing.T) {
	svc := New()
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr error
	}{
		{name: "positive", a: 2, b: 3, want: 5},
		{name: "negative", a: -4, b: -1.5, want: -5.5},
		{name: "with zero", a: 0, b: 9, want: 9},
		{name: "nan operand", a: math.NaN(), b: 1, wantErr: ErrInvalidNumber},
		{name: "inf operand", a: 1, b: math.Inf(1), wantErr: ErrInvalidNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Add(tt.a, tt.b)
			assertResult(t, got, err, tt.want, tt.wantErr)
		})
	}
}

func TestSubtract(t *testing.T) {
	svc := New()
	got, err := svc.Subtract(10, 4)
	assertResult(t, got, err, 6, nil)

	_, err = svc.Subtract(math.Inf(-1), 1)
	if !errors.Is(err, ErrInvalidNumber) {
		t.Fatalf("Subtract(inf): error = %v, want %v", err, ErrInvalidNumber)
	}
}

func TestMultiply(t *testing.T) {
	svc := New()
	got, err := svc.Multiply(3, 4)
	assertResult(t, got, err, 12, nil)

	_, err = svc.Multiply(1e308, 10)
	if !errors.Is(err, ErrResultOutOfRange) {
		t.Fatalf("Multiply overflow: error = %v, want %v", err, ErrResultOutOfRange)
	}
}

func TestDivide(t *testing.T) {
	svc := New()
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr error
	}{
		{name: "exact", a: 10, b: 2, want: 5},
		{name: "fraction", a: 1, b: 4, want: 0.25},
		{name: "negative divisor", a: 9, b: -3, want: -3},
		{name: "zero dividend", a: 0, b: 5, want: 0},
		{name: "division by zero", a: 5, b: 0, wantErr: ErrDivisionByZero},
		{name: "zero by zero", a: 0, b: 0, wantErr: ErrDivisionByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Divide(tt.a, tt.b)
			assertResult(t, got, err, tt.want, tt.wantErr)
		})
	}
}

func TestPower(t *testing.T) {
	svc := New()
	tests := []struct {
		name      string
		base, exp float64
		want      float64
		wantErr   error
	}{
		{name: "integer power", base: 2, exp: 8, want: 256},
		{name: "zero exponent", base: 5, exp: 0, want: 1},
		{name: "fractional exponent", base: 9, exp: 0.5, want: 3},
		{name: "negative exponent", base: 2, exp: -2, want: 0.25},
		{name: "undefined real result", base: -4, exp: 0.5, wantErr: ErrInvalidOperation},
		{name: "overflow", base: 10, exp: 400, wantErr: ErrResultOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Power(tt.base, tt.exp)
			assertResult(t, got, err, tt.want, tt.wantErr)
		})
	}
}

func TestSqrt(t *testing.T) {
	svc := New()
	tests := []struct {
		name    string
		a       float64
		want    float64
		wantErr error
	}{
		{name: "perfect square", a: 16, want: 4},
		{name: "zero", a: 0, want: 0},
		{name: "non integer", a: 2.25, want: 1.5},
		{name: "negative", a: -1, wantErr: ErrNegativeSqrt},
		{name: "nan", a: math.NaN(), wantErr: ErrInvalidNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Sqrt(tt.a)
			assertResult(t, got, err, tt.want, tt.wantErr)
		})
	}
}

func TestPercentage(t *testing.T) {
	svc := New()
	got, err := svc.Percentage(200, 10)
	assertResult(t, got, err, 20, nil)

	got, err = svc.Percentage(50, 0)
	assertResult(t, got, err, 0, nil)

	_, err = svc.Percentage(math.NaN(), 10)
	if !errors.Is(err, ErrInvalidNumber) {
		t.Fatalf("Percentage(nan): error = %v, want %v", err, ErrInvalidNumber)
	}
}

func assertResult(t *testing.T, got float64, err error, want float64, wantErr error) {
	t.Helper()
	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("result = %v, want %v", got, want)
	}
}
