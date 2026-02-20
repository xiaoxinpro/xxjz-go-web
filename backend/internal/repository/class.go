package repository

import (
	"database/sql"
)

type ClassRepo struct {
	db *sql.DB
}

func NewClassRepo(db *sql.DB) *ClassRepo {
	return &ClassRepo{db: db}
}

type Class struct {
	ClassID   int64
	ClassName string
	ClassType int // 1 income, 2 out
	Ufid      int64
	Sort      int
}

func (r *ClassRepo) ListByUID(uid int64, classType int) ([]Class, error) {
	query := `SELECT classid, classname, classtype, ufid, sort FROM xxjz_account_class WHERE ufid = ?`
	args := []interface{}{uid}
	if classType > 0 {
		query += ` AND classtype = ?`
		args = append(args, classType)
	}
	query += ` ORDER BY sort, classid`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Class
	for rows.Next() {
		var c Class
		if err := rows.Scan(&c.ClassID, &c.ClassName, &c.ClassType, &c.Ufid, &c.Sort); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *ClassRepo) GetByID(classID, uid int64) (*Class, error) {
	row := r.db.QueryRow(`SELECT classid, classname, classtype, ufid, sort FROM xxjz_account_class WHERE classid = ? AND ufid = ?`, classID, uid)
	var c Class
	err := row.Scan(&c.ClassID, &c.ClassName, &c.ClassType, &c.Ufid, &c.Sort)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ClassRepo) Create(name string, classType int, uid int64, sort int) (int64, error) {
	res, err := r.db.Exec(`INSERT INTO xxjz_account_class (classname, classtype, ufid, sort) VALUES (?, ?, ?, ?)`, name, classType, uid, sort)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ClassRepo) UpdateName(classID, uid int64, name string) error {
	_, err := r.db.Exec(`UPDATE xxjz_account_class SET classname = ? WHERE classid = ? AND ufid = ?`, name, classID, uid)
	return err
}

func (r *ClassRepo) Delete(classID, uid int64) error {
	_, err := r.db.Exec(`DELETE FROM xxjz_account_class WHERE classid = ? AND ufid = ?`, classID, uid)
	return err
}
