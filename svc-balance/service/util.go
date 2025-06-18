package service

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// validateCurrency validates currency code using the comprehensive currency system
func (service *Service) validateCurrency(currency string) error {
	_, err := service.getCurrencyInfo(currency)
	if err != nil {
		return fmt.Errorf("unsupported currency: %s", currency)
	}
	return nil
}

// pgNumericToDecimal converts pgtype.Numeric to decimal.Decimal
func (service *Service) pgNumericToDecimal(pgNum pgtype.Numeric) (decimal.Decimal, error) {
	if !pgNum.Valid {
		return decimal.Zero, nil
	}

	// Convert pgtype.Numeric to string then to decimal.Decimal
	str := pgNum.Int.String()
	if pgNum.Exp < 0 {
		// Add decimal point
		exp := int(-pgNum.Exp)
		if len(str) <= exp {
			// Pad with zeros
			str = "0." + fmt.Sprintf("%0*s", exp, str)
		} else {
			// Insert decimal point
			pos := len(str) - exp
			str = str[:pos] + "." + str[pos:]
		}
	}

	return decimal.NewFromString(str)
}
