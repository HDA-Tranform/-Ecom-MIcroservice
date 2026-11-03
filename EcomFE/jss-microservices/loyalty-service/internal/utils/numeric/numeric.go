package utils

import (
	"github.com/jackc/pgx/v5/pgtype"
)

func ToNumeric(val float64) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(val) // convert float64 -> pgtype.Numeric
	return num
}

func NumericToFloat64(n pgtype.Numeric) float64 {
	var f float64
	_ = n.Scan(&f) // scan sang float64
	return f
}

func Int32(val int32) pgtype.Int4 {
	return pgtype.Int4{
		Int32: val,
		Valid: true,
	}
}
