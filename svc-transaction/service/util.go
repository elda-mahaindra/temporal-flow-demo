package service

import (
	"fmt"

	"svc-transaction/store/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// decimalToPgNumeric converts a decimal.Decimal to pgtype.Numeric
func (service *Service) decimalToPgNumeric(d decimal.Decimal) (pgtype.Numeric, error) {
	return pgtype.Numeric{
		Int:   d.Coefficient(),
		Exp:   int32(d.Exponent()),
		Valid: true,
	}, nil
}

// pgNumericToDecimal converts a pgtype.Numeric to decimal.Decimal
func (service *Service) pgNumericToDecimal(n pgtype.Numeric) (decimal.Decimal, error) {
	if !n.Valid {
		return decimal.Zero, fmt.Errorf("numeric value is not valid")
	}

	return decimal.NewFromBigInt(n.Int, n.Exp), nil
}

// mapCurrencyToEnum maps a currency string to the corresponding enum value
func (service *Service) mapCurrencyToEnum(currency string) sqlc.CoreCurrencyCode {
	switch currency {
	case "USD":
		return sqlc.CoreCurrencyCodeUSD
	case "EUR":
		return sqlc.CoreCurrencyCodeEUR
	case "GBP":
		return sqlc.CoreCurrencyCodeGBP
	case "JPY":
		return sqlc.CoreCurrencyCodeJPY
	case "CAD":
		return sqlc.CoreCurrencyCodeCAD
	case "AUD":
		return sqlc.CoreCurrencyCodeAUD
	case "CHF":
		return sqlc.CoreCurrencyCodeCHF
	case "CNY":
		return sqlc.CoreCurrencyCodeCNY
	case "SGD":
		return sqlc.CoreCurrencyCodeSGD
	case "HKD":
		return sqlc.CoreCurrencyCodeHKD
	default:
		return sqlc.CoreCurrencyCodeUSD // Default fallback
	}
}
