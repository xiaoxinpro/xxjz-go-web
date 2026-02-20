package repository

import (
	"database/sql"
)

type FundsRepo struct {
	db *sql.DB
}

func NewFundsRepo(db *sql.DB) *FundsRepo {
	return &FundsRepo{db: db}
}

type Fund struct {
	FundsID   int64
	FundsName string
	UID       int64
	Sort      int
}

func (r *FundsRepo) ListByUID(uid int64) ([]Fund, error) {
	rows, err := r.db.Query(`SELECT fundsid, fundsname, uid, sort FROM xxjz_account_funds WHERE uid = ? ORDER BY sort, fundsid`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Fund
	for rows.Next() {
		var f Fund
		if err := rows.Scan(&f.FundsID, &f.FundsName, &f.UID, &f.Sort); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, rows.Err()
}

func (r *FundsRepo) GetByID(fundsID, uid int64) (*Fund, error) {
	row := r.db.QueryRow(`SELECT fundsid, fundsname, uid, sort FROM xxjz_account_funds WHERE fundsid = ? AND uid = ?`, fundsID, uid)
	var f Fund
	err := row.Scan(&f.FundsID, &f.FundsName, &f.UID, &f.Sort)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FundsRepo) Create(name string, uid int64, sort int) (int64, error) {
	res, err := r.db.Exec(`INSERT INTO xxjz_account_funds (fundsname, uid, sort) VALUES (?, ?, ?)`, name, uid, sort)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *FundsRepo) UpdateName(fundsID, uid int64, name string) error {
	_, err := r.db.Exec(`UPDATE xxjz_account_funds SET fundsname = ? WHERE fundsid = ? AND uid = ?`, name, fundsID, uid)
	return err
}

func (r *FundsRepo) Delete(fundsID, uid int64) error {
	_, err := r.db.Exec(`DELETE FROM xxjz_account_funds WHERE fundsid = ? AND uid = ?`, fundsID, uid)
	return err
}
