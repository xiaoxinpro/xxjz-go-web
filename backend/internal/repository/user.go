package repository

import (
	"database/sql"
	"time"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

type User struct {
	UID      int64
	Username string
	Password string
	Email    string
	Utime    int64
}

func (r *UserRepo) GetByUsername(username string) (*User, error) {
	row := r.db.QueryRow(
		`SELECT uid, username, password, email, utime FROM xxjz_user WHERE username = ?`,
		username,
	)
	var u User
	err := row.Scan(&u.UID, &u.Username, &u.Password, &u.Email, &u.Utime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByUID(uid int64) (*User, error) {
	row := r.db.QueryRow(
		`SELECT uid, username, password, email, utime FROM xxjz_user WHERE uid = ?`,
		uid,
	)
	var u User
	err := row.Scan(&u.UID, &u.Username, &u.Password, &u.Email, &u.Utime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetEmail(uid int64) (string, error) {
	var email string
	err := r.db.QueryRow(`SELECT email FROM xxjz_user WHERE uid = ?`, uid).Scan(&email)
	if err != nil {
		return "", err
	}
	return email, nil
}

func (r *UserRepo) Create(username, password, email string) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO xxjz_user (username, password, email, utime) VALUES (?, ?, ?, ?)`,
		username, password, email, time.Now().Unix(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UserRepo) UpdateUsername(uid int64, username string) error {
	_, err := r.db.Exec(`UPDATE xxjz_user SET username = ? WHERE uid = ?`, username, uid)
	return err
}

func (r *UserRepo) UpdatePassword(uid int64, passwordHash string) error {
	_, err := r.db.Exec(`UPDATE xxjz_user SET password = ? WHERE uid = ?`, passwordHash, uid)
	return err
}

func (r *UserRepo) UsernameExists(username string, excludeUID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(1) FROM xxjz_user WHERE username = ? AND uid != ?`,
		username, excludeUID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepo) EmailExists(email string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(1) FROM xxjz_user WHERE email = ?`, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountUsers returns the number of users (for "initialized" check).
func (r *UserRepo) CountUsers() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(1) FROM xxjz_user`).Scan(&n)
	return n, err
}
