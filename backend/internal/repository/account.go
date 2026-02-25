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
	return r.SumByTimeAndTypeAndClass(uid, typ, 0, start, end)
}

// SumByTimeAndTypeAndClass returns sum for jiid=uid, zhifu=typ; if classID != 0 filters by acclassid. start/end optional (IsZero = no filter).
func (r *AccountRepo) SumByTimeAndTypeAndClass(uid int64, typ int, classID int64, start, end time.Time) (float64, error) {
	query := `SELECT COALESCE(SUM(acmoney),0) FROM xxjz_account WHERE jiid = ? AND zhifu = ?`
	args := []interface{}{uid, typ}
	if classID != 0 {
		query += ` AND acclassid = ?`
		args = append(args, classID)
	}
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

// FindRow is the unified row shape for account or transfer list (id, money, classid, class, typeid, type, funds, time, mark).
// Fid is set only for single-account get (get_id) for edit form.
type FindRow struct {
	ID      int64   `json:"id"`
	Money   float64 `json:"money"`
	ClassID int64   `json:"classid"`
	Class   string  `json:"class"`
	TypeID  int     `json:"typeid"`
	Type    string  `json:"type"`
	Funds   string  `json:"funds"`
	Fid     int64   `json:"fid,omitempty"`
	Time    int64   `json:"time"`
	Mark    string  `json:"mark"`
}

// ListByUserWithClassFunds returns account rows for the user with class name and funds name, ordered by actime DESC.
// Limit and offset are for pagination; use 0,0 to get all (for merge-then-paginate we use large limit and do pagination in service).
func (r *AccountRepo) ListByUserWithClassFunds(uid int64, limit, offset int) ([]FindRow, error) {
	query := `SELECT a.acid AS id, a.acmoney AS money, a.acclassid AS classid,
		COALESCE(c.classname,'') AS class, a.zhifu AS typeid,
		CASE a.zhifu WHEN 1 THEN '收入' WHEN 2 THEN '支出' ELSE '' END AS type,
		COALESCE(f.fundsname,'') AS funds, a.actime AS time, COALESCE(a.acremark,'') AS mark
		FROM xxjz_account a
		LEFT JOIN xxjz_account_class c ON c.classid = a.acclassid
		LEFT JOIN xxjz_account_funds f ON f.fundsid = a.fid
		WHERE a.jiid = ?
		ORDER BY a.actime DESC, a.acid DESC`
	args := []interface{}{uid}
	if limit > 0 {
		query += ` LIMIT ? OFFSET ?`
		args = append(args, limit, offset)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []FindRow
	for rows.Next() {
		var row FindRow
		err := rows.Scan(&row.ID, &row.Money, &row.ClassID, &row.Class, &row.TypeID, &row.Type, &row.Funds, &row.Time, &row.Mark)
		if err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

// CountByUser returns total account count for the user.
func (r *AccountRepo) CountByUser(uid int64) (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM xxjz_account WHERE jiid = ?`, uid).Scan(&n)
	return n, err
}

// Insert inserts one account row; returns acid.
func (r *AccountRepo) Insert(jiid int64, acmoney float64, acclassid, actime int64, acremark string, zhifu, fid int64) (int64, error) {
	res, err := r.db.Exec(`INSERT INTO xxjz_account (acmoney, acclassid, actime, acremark, jiid, zhifu, fid) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		acmoney, acclassid, actime, acremark, jiid, zhifu, fid)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Delete deletes one account row by acid and jiid; returns rows affected.
func (r *AccountRepo) Delete(acid, jiid int64) (int64, error) {
	res, err := r.db.Exec(`DELETE FROM xxjz_account WHERE acid = ? AND jiid = ?`, acid, jiid)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// HasDefaultFunds reports whether the user has any account rows with fid = -1 (默认账户).
func (r *AccountRepo) HasDefaultFunds(uid int64) (bool, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM xxjz_account WHERE jiid = ? AND fid = -1`, uid).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// FundsStats returns income sum, expense sum, and record count for one fund (fid) and user.
// over = in - out; caller may add transfer init for fid > 0 if needed.
func (r *AccountRepo) FundsStats(uid, fid int64) (sumIn, sumOut float64, count int, err error) {
	base := `SELECT COALESCE(SUM(acmoney),0) FROM xxjz_account WHERE jiid = ? AND fid = ? AND zhifu = ?`
	err = r.db.QueryRow(base, uid, fid, 1).Scan(&sumIn)
	if err != nil {
		return 0, 0, 0, err
	}
	err = r.db.QueryRow(base, uid, fid, 2).Scan(&sumOut)
	if err != nil {
		return 0, 0, 0, err
	}
	err = r.db.QueryRow(`SELECT COUNT(*) FROM xxjz_account WHERE jiid = ? AND fid = ?`, uid, fid).Scan(&count)
	if err != nil {
		return 0, 0, 0, err
	}
	return sumIn, sumOut, count, nil
}

// ReassignFunds moves all account rows from oldFID to newFID for the user.
func (r *AccountRepo) ReassignFunds(uid, oldFID, newFID int64) (int64, error) {
	res, err := r.db.Exec(`UPDATE xxjz_account SET fid = ? WHERE jiid = ? AND fid = ?`, newFID, uid, oldFID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// GetByID returns one account row by acid and jiid with class and funds names, or nil if not found. Includes fid for edit form.
func (r *AccountRepo) GetByID(acid, jiid int64) (*FindRow, error) {
	query := `SELECT a.acid AS id, a.acmoney AS money, a.acclassid AS classid,
		COALESCE(c.classname,'') AS class, a.zhifu AS typeid,
		CASE a.zhifu WHEN 1 THEN '收入' WHEN 2 THEN '支出' ELSE '' END AS type,
		COALESCE(f.fundsname,'') AS funds, a.fid, a.actime AS time, COALESCE(a.acremark,'') AS mark
		FROM xxjz_account a
		LEFT JOIN xxjz_account_class c ON c.classid = a.acclassid
		LEFT JOIN xxjz_account_funds f ON f.fundsid = a.fid
		WHERE a.acid = ? AND a.jiid = ?`
	var row FindRow
	err := r.db.QueryRow(query, acid, jiid).Scan(&row.ID, &row.Money, &row.ClassID, &row.Class, &row.TypeID, &row.Type, &row.Funds, &row.Fid, &row.Time, &row.Mark)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Update updates one account row by acid and jiid.
func (r *AccountRepo) Update(acid, jiid int64, acmoney float64, acclassid, actime int64, acremark string, zhifu, fid int64) (int64, error) {
	res, err := r.db.Exec(`UPDATE xxjz_account SET acmoney=?, acclassid=?, actime=?, acremark=?, zhifu=?, fid=? WHERE acid=? AND jiid=?`,
		acmoney, acclassid, actime, acremark, zhifu, fid, acid, jiid)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// AccountFilter for filtered list/count. Zero values mean no filter.
type AccountFilter struct {
	Fid       int64
	Zhifu     int   // 0=all, 1=收入, 2=支出
	Acclassid int64
	StartTime int64
	EndTime   int64
	Acremark  string
}

func (r *AccountRepo) listQuery(uid int64, f AccountFilter, limit, offset int) (string, []interface{}) {
	query := `SELECT a.acid AS id, a.acmoney AS money, a.acclassid AS classid,
		COALESCE(c.classname,'') AS class, a.zhifu AS typeid,
		CASE a.zhifu WHEN 1 THEN '收入' WHEN 2 THEN '支出' ELSE '' END AS type,
		COALESCE(f.fundsname,'') AS funds, a.actime AS time, COALESCE(a.acremark,'') AS mark
		FROM xxjz_account a
		LEFT JOIN xxjz_account_class c ON c.classid = a.acclassid
		LEFT JOIN xxjz_account_funds f ON f.fundsid = a.fid
		WHERE a.jiid = ?`
	args := []interface{}{uid}
	if f.Fid != 0 && f.Fid != -1 {
		query += ` AND a.fid = ?`
		args = append(args, f.Fid)
	}
	if f.Zhifu == 1 || f.Zhifu == 2 {
		query += ` AND a.zhifu = ?`
		args = append(args, f.Zhifu)
	}
	if f.Acclassid != 0 {
		query += ` AND a.acclassid = ?`
		args = append(args, f.Acclassid)
	}
	if f.StartTime > 0 {
		query += ` AND a.actime >= ?`
		args = append(args, f.StartTime)
	}
	if f.EndTime > 0 {
		query += ` AND a.actime <= ?`
		args = append(args, f.EndTime)
	}
	if f.Acremark != "" {
		query += ` AND a.acremark LIKE ?`
		args = append(args, "%"+f.Acremark+"%")
	}
	query += ` ORDER BY a.actime DESC, a.acid DESC`
	if limit > 0 {
		query += ` LIMIT ? OFFSET ?`
		args = append(args, limit, offset)
	}
	return query, args
}

// ListByUserFiltered returns account rows with optional filters.
func (r *AccountRepo) ListByUserFiltered(uid int64, f AccountFilter, limit, offset int) ([]FindRow, error) {
	q, args := r.listQuery(uid, f, limit, offset)
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []FindRow
	for rows.Next() {
		var row FindRow
		if err := rows.Scan(&row.ID, &row.Money, &row.ClassID, &row.Class, &row.TypeID, &row.Type, &row.Funds, &row.Time, &row.Mark); err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

// CountByUserFiltered returns count with optional filters.
func (r *AccountRepo) CountByUserFiltered(uid int64, f AccountFilter) (int, error) {
	query := `SELECT COUNT(*) FROM xxjz_account a WHERE a.jiid = ?`
	args := []interface{}{uid}
	if f.Fid != 0 && f.Fid != -1 {
		query += ` AND a.fid = ?`
		args = append(args, f.Fid)
	}
	if f.Zhifu == 1 || f.Zhifu == 2 {
		query += ` AND a.zhifu = ?`
		args = append(args, f.Zhifu)
	}
	if f.Acclassid != 0 {
		query += ` AND a.acclassid = ?`
		args = append(args, f.Acclassid)
	}
	if f.StartTime > 0 {
		query += ` AND a.actime >= ?`
		args = append(args, f.StartTime)
	}
	if f.EndTime > 0 {
		query += ` AND a.actime <= ?`
		args = append(args, f.EndTime)
	}
	if f.Acremark != "" {
		query += ` AND a.acremark LIKE ?`
		args = append(args, "%"+f.Acremark+"%")
	}
	var n int
	err := r.db.QueryRow(query, args...).Scan(&n)
	return n, err
}

// SumByUserFiltered returns sum in (zhifu=1) and sum out (zhifu=2) with optional filters.
func (r *AccountRepo) SumByUserFiltered(uid int64, f AccountFilter) (sumIn, sumOut float64, err error) {
	base := `SELECT COALESCE(SUM(acmoney),0) FROM xxjz_account WHERE jiid = ?`
	args := []interface{}{uid}
	if f.Fid != 0 && f.Fid != -1 {
		base += ` AND fid = ?`
		args = append(args, f.Fid)
	}
	if f.Acclassid != 0 {
		base += ` AND acclassid = ?`
		args = append(args, f.Acclassid)
	}
	if f.StartTime > 0 {
		base += ` AND actime >= ?`
		args = append(args, f.StartTime)
	}
	if f.EndTime > 0 {
		base += ` AND actime <= ?`
		args = append(args, f.EndTime)
	}
	if f.Acremark != "" {
		base += ` AND acremark LIKE ?`
		args = append(args, "%"+f.Acremark+"%")
	}
	qIn := base + ` AND zhifu = 1`
	qOut := base + ` AND zhifu = 2`
	if f.Zhifu == 1 {
		err = r.db.QueryRow(qIn, args...).Scan(&sumIn)
		return sumIn, 0, err
	}
	if f.Zhifu == 2 {
		err = r.db.QueryRow(qOut, args...).Scan(&sumOut)
		return 0, sumOut, err
	}
	_ = r.db.QueryRow(qIn, args...).Scan(&sumIn)
	_ = r.db.QueryRow(qOut, args...).Scan(&sumOut)
	return sumIn, sumOut, nil
}
