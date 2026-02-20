package repository

import (
	"database/sql"
	"time"
)

type AccountRepo struct {
	db *sql.DB
}

func NewAccountRepo(db *sql.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

// SumByTimeAndType returns sum of acmoney for jiid=uid, zhifu=typ, and optional time range. If start IsZero, no start filter; if end IsZero, no end filter.
func (r *AccountRepo) SumByTimeAndType(uid int64, typ int, start, end time.Time) (float64, error) {
	query := `SELECT COALESCE(SUM(acmoney),0) FROM xxjz_account WHERE jiid = ? AND zhifu = ?`
	args := []interface{}{uid, typ}
	if !start.IsZero() {
		query += ` AND actime >= ?`
		args = append(args, start.Unix())
	}
	if !end.IsZero() {
		query += ` AND actime <= ?`
		args = append(args, end.Unix())
	}
	var sum float64
	err := r.db.QueryRow(query, args...).Scan(&sum)
	return sum, err
}
