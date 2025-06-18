package service

import (
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestValidateCurrency(t *testing.T) {
	t.Parallel()

	service := createTestService()

	tests := []struct {
		name        string
		currency    string
		expectError bool
	}{
		{
			name:        "Valid USD currency",
			currency:    "USD",
			expectError: false,
		},
		{
			name:        "Valid EUR currency",
			currency:    "EUR",
			expectError: false,
		},
		{
			name:        "Valid JPY currency",
			currency:    "JPY",
			expectError: false,
		},
		{
			name:        "Invalid currency",
			currency:    "INVALID",
			expectError: true,
		},
		{
			name:        "Empty currency",
			currency:    "",
			expectError: true,
		},
		{
			name:        "Case insensitive test",
			currency:    "usd", // lowercase should work due to normalization
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := service.validateCurrency(tt.currency)

			if tt.expectError {
				if err == nil {
					t.Error("validateCurrency() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("validateCurrency() error = %v", err)
				}
			}
		})
	}
}

func TestPgNumericToDecimal(t *testing.T) {
	t.Parallel()

	service := createTestService()

	tests := []struct {
		name           string
		pgNum          pgtype.Numeric
		expectedResult string
		expectError    bool
	}{
		{
			name: "Valid positive decimal",
			pgNum: pgtype.Numeric{
				Int:   big.NewInt(12345),
				Exp:   -2, // 123.45
				Valid: true,
			},
			expectedResult: "123.45",
			expectError:    false,
		},
		{
			name: "Valid integer",
			pgNum: pgtype.Numeric{
				Int:   big.NewInt(12345),
				Exp:   0,
				Valid: true,
			},
			expectedResult: "12345",
			expectError:    false,
		},
		{
			name: "Valid zero",
			pgNum: pgtype.Numeric{
				Int:   big.NewInt(0),
				Exp:   0,
				Valid: true,
			},
			expectedResult: "0",
			expectError:    false,
		},
		{
			name: "Valid small decimal",
			pgNum: pgtype.Numeric{
				Int:   big.NewInt(123),
				Exp:   -4, // 0.0123
				Valid: true,
			},
			expectedResult: "0.0123",
			expectError:    false,
		},
		{
			name: "Valid negative decimal",
			pgNum: pgtype.Numeric{
				Int:   big.NewInt(-12345),
				Exp:   -2, // -123.45
				Valid: true,
			},
			expectedResult: "-123.45",
			expectError:    false,
		},
		{
			name: "Invalid (NULL) numeric",
			pgNum: pgtype.Numeric{
				Int:   big.NewInt(12345),
				Exp:   -2,
				Valid: false,
			},
			expectedResult: "0",
			expectError:    false,
		},
		{
			name: "Very small decimal with padding",
			pgNum: pgtype.Numeric{
				Int:   big.NewInt(1),
				Exp:   -5, // 0.00001
				Valid: true,
			},
			expectedResult: "0.00001",
			expectError:    false,
		},
		{
			name: "Large integer with decimal places",
			pgNum: pgtype.Numeric{
				Int:   big.NewInt(123456789),
				Exp:   -3, // 123456.789
				Valid: true,
			},
			expectedResult: "123456.789",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := service.pgNumericToDecimal(tt.pgNum)

			if tt.expectError {
				if err == nil {
					t.Error("pgNumericToDecimal() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("pgNumericToDecimal() error = %v", err)
				return
			}

			if result.String() != tt.expectedResult {
				t.Errorf("pgNumericToDecimal() = %s, want %s", result.String(), tt.expectedResult)
			}
		})
	}
}

func TestPgNumericToDecimalEdgeCases(t *testing.T) {
	t.Parallel()

	service := createTestService()

	// Test with very large numbers
	t.Run("Large number", func(t *testing.T) {
		t.Parallel()

		bigInt := big.NewInt(0)
		bigInt.SetString("123456789012345678901234567890", 10)

		pgNum := pgtype.Numeric{
			Int:   bigInt,
			Exp:   -10,
			Valid: true,
		}

		result, err := service.pgNumericToDecimal(pgNum)
		if err != nil {
			t.Errorf("pgNumericToDecimal() error = %v", err)
			return
		}

		expected := "12345678901234567890.123456789"
		if result.String() != expected {
			t.Errorf("pgNumericToDecimal() = %s, want %s", result.String(), expected)
		}
	})

	// Test with zero exponent
	t.Run("Zero exponent", func(t *testing.T) {
		t.Parallel()

		pgNum := pgtype.Numeric{
			Int:   big.NewInt(123),
			Exp:   0,
			Valid: true,
		}

		result, err := service.pgNumericToDecimal(pgNum)
		if err != nil {
			t.Errorf("pgNumericToDecimal() error = %v", err)
			return
		}

		if result.String() != "123" {
			t.Errorf("pgNumericToDecimal() = %s, want %s", result.String(), "123")
		}
	})

	// Test with positive exponent (should not happen in normal PostgreSQL usage but test for robustness)
	t.Run("Positive exponent", func(t *testing.T) {
		t.Parallel()

		pgNum := pgtype.Numeric{
			Int:   big.NewInt(123),
			Exp:   2,
			Valid: true,
		}

		result, err := service.pgNumericToDecimal(pgNum)
		if err != nil {
			t.Errorf("pgNumericToDecimal() error = %v", err)
			return
		}

		// With positive exponent, it should just return the integer as-is
		if result.String() != "123" {
			t.Errorf("pgNumericToDecimal() = %s, want %s", result.String(), "123")
		}
	})
}
