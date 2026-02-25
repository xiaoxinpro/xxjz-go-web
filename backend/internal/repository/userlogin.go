package repository

import "database/sql"

const weixinLoginName = "Weixin"

// GetUIDByWeixinOpenID returns the uid bound to this Weixin openid, or 0 if not bound.
func (r *UserRepo) GetUIDByWeixinOpenID(openid string) (int64, error) {
	var uid int64
	err := r.db.QueryRow(
		`SELECT uid FROM xxjz_user_login WHERE login_name = ? AND login_id = ?`,
		weixinLoginName, openid,
	).Scan(&uid)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return uid, nil
}

// InsertWeixinLogin binds openid to uid. Returns error if openid already bound.
func (r *UserRepo) InsertWeixinLogin(uid int64, openid, sessionKey, unionid string) error {
	if sessionKey == "" {
		sessionKey = "null"
	}
	if unionid == "" {
		unionid = "null"
	}
	_, err := r.db.Exec(
		`INSERT INTO xxjz_user_login (uid, login_name, login_id, login_key, login_token) VALUES (?, ?, ?, ?, ?)`,
		uid, weixinLoginName, openid, sessionKey, unionid,
	)
	return err
}
