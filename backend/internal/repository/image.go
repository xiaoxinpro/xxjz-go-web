package repository

import (
	"database/sql"
)

// ImageRow represents one row from xxjz_account_image for listing.
type ImageRow struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Savepath string `json:"savepath"`
	Savename string `json:"savename"`
	Time     int64  `json:"time"`
}

// ImageRepo accesses xxjz_account_image.
type ImageRepo struct {
	db *sql.DB
}

func NewImageRepo(db *sql.DB) *ImageRepo {
	return &ImageRepo{db: db}
}

// ListByAcid returns images for the given account (uid+acid).
func (r *ImageRepo) ListByAcid(uid, acid int64) ([]ImageRow, error) {
	rows, err := r.db.Query(`SELECT id, name, savepath, savename, time FROM xxjz_account_image WHERE uid = ? AND acid = ? ORDER BY time DESC, id DESC`,
		uid, acid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ImageRow
	for rows.Next() {
		var row ImageRow
		if err := rows.Scan(&row.ID, &row.Name, &row.Savepath, &row.Savename, &row.Time); err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

// GetByID returns one image row by id and uid (for delete: resolve path).
func (r *ImageRepo) GetByID(uid, imageID int64) (savepath, savename string, err error) {
	err = r.db.QueryRow(`SELECT savepath, savename FROM xxjz_account_image WHERE id = ? AND uid = ?`, imageID, uid).Scan(&savepath, &savename)
	return savepath, savename, err
}

// Insert inserts one image row; returns id.
func (r *ImageRepo) Insert(uid, acid int64, name, ext, savepath, savename string, size int, md5 string, t int64) (int64, error) {
	res, err := r.db.Exec(`INSERT INTO xxjz_account_image (uid, acid, name, type, size, ext, md5, savepath, savename, time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uid, acid, name, "image/"+ext, size, ext, md5, savepath, savename, t)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateAcid sets acid for the given image (set_image: bind image to account).
func (r *ImageRepo) UpdateAcid(uid, imageID, acid int64) (int64, error) {
	res, err := r.db.Exec(`UPDATE xxjz_account_image SET acid = ? WHERE id = ? AND uid = ?`, acid, imageID, uid)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Delete deletes one image by uid, acid, imageID. If imageID is 0, deletes all images for that acid.
func (r *ImageRepo) Delete(uid, acid, imageID int64) (int64, error) {
	if imageID != 0 {
		res, err := r.db.Exec(`DELETE FROM xxjz_account_image WHERE uid = ? AND acid = ? AND id = ?`, uid, acid, imageID)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	}
	res, err := r.db.Exec(`DELETE FROM xxjz_account_image WHERE uid = ? AND acid = ?`, uid, acid)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// CountByAcid returns number of images for the account.
func (r *ImageRepo) CountByAcid(uid, acid int64) (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM xxjz_account_image WHERE uid = ? AND acid = ?`, uid, acid).Scan(&n)
	return n, err
}

// ListByAcidForDelete returns all image rows for an acid (for deleting files when deleting account).
func (r *ImageRepo) ListByAcidForDelete(uid, acid int64) ([]ImageRow, error) {
	return r.ListByAcid(uid, acid)
}
