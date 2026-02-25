package repository

import (
	"database/sql"
)

// TransferRepo accesses xxjz_account_transfer.
type TransferRepo struct {
	db *sql.DB
}

func NewTransferRepo(db *sql.DB) *TransferRepo {
	return &TransferRepo{db: db}
}

// ListByUserWithFunds returns transfer rows for the user as FindRow: classid=0, class="转账".
// Each transfer is one row: typeid=2 (支出), funds=source funds name (money left source account).
func (r *TransferRepo) ListByUserWithFunds(uid int64, limit, offset int) ([]FindRow, error) {
	query := `SELECT t.tid AS id, t.money, 0 AS classid, '转账' AS class, 2 AS typeid, '支出' AS type,
		COALESCE(f.fundsname,'') AS funds, t.time, COALESCE(t.mark,'') AS mark
		FROM xxjz_account_transfer t
		LEFT JOIN xxjz_account_funds f ON f.fundsid = t.source_fid
		WHERE t.uid = ?
		ORDER BY t.time DESC, t.tid DESC`
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

// CountByUser returns total transfer count for the user.
func (r *TransferRepo) CountByUser(uid int64) (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM xxjz_account_transfer WHERE uid = ?`, uid).Scan(&n)
	return n, err
}

// Insert inserts one transfer row; returns tid.
func (r *TransferRepo) Insert(uid int64, money float64, source_fid, target_fid, time int64, mark string) (int64, error) {
	res, err := r.db.Exec(`INSERT INTO xxjz_account_transfer (uid, money, source_fid, target_fid, time, mark) VALUES (?, ?, ?, ?, ?, ?)`,
		uid, money, source_fid, target_fid, time, mark)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Delete deletes one transfer row by tid and uid; returns rows affected.
func (r *TransferRepo) Delete(tid, uid int64) (int64, error) {
	res, err := r.db.Exec(`DELETE FROM xxjz_account_transfer WHERE tid = ? AND uid = ?`, tid, uid)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ReassignFunds moves transfer rows from oldFID to newFID for the user (both source_fid and target_fid).
func (r *TransferRepo) ReassignFunds(uid, oldFID, newFID int64) (int64, error) {
	var total int64
	// 更新转出账户
	if res, err := r.db.Exec(`UPDATE xxjz_account_transfer SET source_fid = ? WHERE uid = ? AND source_fid = ?`, newFID, uid, oldFID); err != nil {
		return 0, err
	} else {
		n, _ := res.RowsAffected()
		total += n
	}
	// 更新转入账户
	if res, err := r.db.Exec(`UPDATE xxjz_account_transfer SET target_fid = ? WHERE uid = ? AND target_fid = ?`, newFID, uid, oldFID); err != nil {
		return 0, err
	} else {
		n, _ := res.RowsAffected()
		total += n
	}
	return total, nil
}

// TransferFilter for filtered list/count. Zero values mean no filter. Direction: 0=all, 1=in (target_fid), 2=out (source_fid).
type TransferFilter struct {
	Fid       int64
	Direction int   // 0=all, 1=in, 2=out
	StartTime int64
	EndTime   int64
	Mark      string
}

func (r *TransferRepo) listQuery(uid int64, f TransferFilter, limit, offset int) (string, []interface{}) {
	query := `SELECT t.tid AS id, t.money, 0 AS classid, '转账' AS class, 2 AS typeid, '支出' AS type,
		COALESCE(f.fundsname,'') AS funds, t.time, COALESCE(t.mark,'') AS mark
		FROM xxjz_account_transfer t
		LEFT JOIN xxjz_account_funds f ON f.fundsid = t.source_fid
		WHERE t.uid = ?`
	args := []interface{}{uid}
	if f.Fid != 0 {
		if f.Direction == 1 {
			query += ` AND t.target_fid = ?`
		} else if f.Direction == 2 {
			query += ` AND t.source_fid = ?`
		} else {
			query += ` AND (t.source_fid = ? OR t.target_fid = ?)`
			args = append(args, f.Fid, f.Fid)
		}
	}
	if f.Fid != 0 && f.Direction != 0 && f.Direction != 1 && f.Direction != 2 {
		// already handled above
	}
	if f.Fid != 0 && (f.Direction == 1 || f.Direction == 2) {
		args = append(args, f.Fid)
	}
	if f.StartTime > 0 {
		query += ` AND t.time >= ?`
		args = append(args, f.StartTime)
	}
	if f.EndTime > 0 {
		query += ` AND t.time <= ?`
		args = append(args, f.EndTime)
	}
	if f.Mark != "" {
		query += ` AND t.mark LIKE ?`
		args = append(args, "%"+f.Mark+"%")
	}
	query += ` ORDER BY t.time DESC, t.tid DESC`
	if limit > 0 {
		query += ` LIMIT ? OFFSET ?`
		args = append(args, limit, offset)
	}
	return query, args
}

// ListByUserFiltered returns transfer rows with optional filters.
func (r *TransferRepo) ListByUserFiltered(uid int64, f TransferFilter, limit, offset int) ([]FindRow, error) {
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
func (r *TransferRepo) CountByUserFiltered(uid int64, f TransferFilter) (int, error) {
	query := `SELECT COUNT(*) FROM xxjz_account_transfer t WHERE t.uid = ?`
	args := []interface{}{uid}
	if f.Fid != 0 {
		if f.Direction == 1 {
			query += ` AND t.target_fid = ?`
		} else if f.Direction == 2 {
			query += ` AND t.source_fid = ?`
		} else {
			query += ` AND (t.source_fid = ? OR t.target_fid = ?)`
			args = append(args, f.Fid, f.Fid)
		}
	}
	if f.Fid != 0 && (f.Direction == 1 || f.Direction == 2) {
		args = append(args, f.Fid)
	}
	if f.StartTime > 0 {
		query += ` AND t.time >= ?`
		args = append(args, f.StartTime)
	}
	if f.EndTime > 0 {
		query += ` AND t.time <= ?`
		args = append(args, f.EndTime)
	}
	if f.Mark != "" {
		query += ` AND t.mark LIKE ?`
		args = append(args, "%"+f.Mark+"%")
	}
	var n int
	err := r.db.QueryRow(query, args...).Scan(&n)
	return n, err
}

// SumByUserFiltered returns sum in (target side) and sum out (source side) with optional filters.
func (r *TransferRepo) SumByUserFiltered(uid int64, f TransferFilter) (sumIn, sumOut float64, err error) {
	baseIn := `SELECT COALESCE(SUM(money),0) FROM xxjz_account_transfer WHERE uid = ?`
	baseOut := `SELECT COALESCE(SUM(money),0) FROM xxjz_account_transfer WHERE uid = ?`
	argsIn := []interface{}{uid}
	argsOut := []interface{}{uid}
	if f.Fid != 0 {
		if f.Direction != 2 {
			baseIn += ` AND target_fid = ?`
			argsIn = append(argsIn, f.Fid)
		}
		if f.Direction != 1 {
			baseOut += ` AND source_fid = ?`
			argsOut = append(argsOut, f.Fid)
		}
	}
	if f.StartTime > 0 {
		baseIn += ` AND time >= ?`
		baseOut += ` AND time >= ?`
		argsIn = append(argsIn, f.StartTime)
		argsOut = append(argsOut, f.StartTime)
	}
	if f.EndTime > 0 {
		baseIn += ` AND time <= ?`
		baseOut += ` AND time <= ?`
		argsIn = append(argsIn, f.EndTime)
		argsOut = append(argsOut, f.EndTime)
	}
	if f.Mark != "" {
		baseIn += ` AND mark LIKE ?`
		baseOut += ` AND mark LIKE ?`
		argsIn = append(argsIn, "%"+f.Mark+"%")
		argsOut = append(argsOut, "%"+f.Mark+"%")
	}
	_ = r.db.QueryRow(baseIn, argsIn...).Scan(&sumIn)
	_ = r.db.QueryRow(baseOut, argsOut...).Scan(&sumOut)
	return sumIn, sumOut, nil
}
