package dao

import (
	"database/sql"

	"github.com/giaptai/finan-test-short-url/model"
)

type URLRespository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) *URLRespository {
	return &URLRespository{db: db}
}

// get all
func (r *URLRespository) GetAll(limit, offset int) ([]*model.URL, error) {
	query := `SELECT id, short_code, original_url, clicks, created_at FROM url 
	ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	urls := []*model.URL{}

	for rows.Next() {
		url := &model.URL{}
		err := rows.Scan(
			&url.ID,
			&url.ShortCode,
			&url.OriginalURL,
			&url.Clicks,
			&url.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

// get one
func (r *URLRespository) GetByShortCode(shortCode string) (*model.URL, error) {
	query := `SELECT id, short_code, original_url, clicks, created_at 
	FROM url WHERE short_code = $1`
	url := &model.URL{}
	row := r.db.QueryRow(query, shortCode)
	if err := row.Scan(&url.ID, &url.ShortCode, &url.OriginalURL, &url.Clicks, &url.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return url, nil
}

// add new
func (r *URLRespository) Add(url *model.URL) error {
	query := `INSERT INTO url (short_code, original_url, clicks, created_at) 
	VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(query,
		url.ShortCode,
		url.OriginalURL,
		url.Clicks,
		url.CreatedAt).
		Scan(&url.ID)
	return err
}

// upd one - Atomic operation - avoid race condition
func (r *URLRespository) IncrementClicks(shortCode string) error {
	query := `UPDATE url SET clicks = clicks + 1 WHERE short_code = $1`
	_, err := r.db.Exec(query, shortCode)
	return err
}

// check one
func (r *URLRespository) ExistsByShortCode(shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM url WHERE short_code = $1)`
	var exists bool
	err := r.db.QueryRow(query, shortCode).Scan(&exists)
	return exists, err
}

// del one
func (r *URLRespository) Del(shortCode string) error {
	query := `DELETE FROM url WHERE short_code = $1`
	_, err := r.db.Exec(query, shortCode)
	return err
}
