// db/share.go
package mysql

import (
	"database/sql"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/share"
)

type shareBackend struct {
	db *sql.DB
}

func (s shareBackend) All() ([]*share.Link, error) {
	rows, err := s.db.Query("SELECT * FROM shares")
	if err == sql.ErrNoRows {
		return nil, errors.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*share.Link
	for rows.Next() {
		link := &share.Link{}
		err := rows.Scan(
			&link.Hash,
			&link.UserID,
			&link.Path,
			&link.Expire,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func (s shareBackend) FindByUserID(id uint) ([]*share.Link, error) {
	rows, err := s.db.Query("SELECT * FROM shares WHERE user_id = ?", id)
	if err == sql.ErrNoRows {
		return nil, errors.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*share.Link
	for rows.Next() {
		link := &share.Link{}
		err := rows.Scan(
			&link.Hash,
			&link.UserID,
			&link.Path,
			&link.Expire,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func (s shareBackend) GetByHash(hash string) (*share.Link, error) {
	link := &share.Link{}
	err := s.db.QueryRow("SELECT * FROM shares WHERE hash = ?", hash).Scan(
		&link.Hash,
		&link.UserID,
		&link.Path,
		&link.Expire,
	)
	if err == sql.ErrNoRows {
		return nil, errors.ErrNotExist
	}
	return link, err
}

func (s shareBackend) GetPermanent(path string, id uint) (*share.Link, error) {
	link := &share.Link{}
	err := s.db.QueryRow(
		"SELECT * FROM shares WHERE path = ? AND expire = 0 AND user_id = ? LIMIT 1",
		path, id,
	).Scan(
		&link.Hash,
		&link.UserID,
		&link.Path,
		&link.Expire,
	)
	if err == sql.ErrNoRows {
		return nil, errors.ErrNotExist
	}
	return link, err
}

func (s shareBackend) Gets(path string, id uint) ([]*share.Link, error) {
	rows, err := s.db.Query("SELECT * FROM shares WHERE path = ? AND user_id = ?", path, id)
	if err == sql.ErrNoRows {
		return nil, errors.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*share.Link
	for rows.Next() {
		link := &share.Link{}
		err := rows.Scan(
			&link.Hash,
			&link.UserID,
			&link.Path,
			&link.Expire,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func (s shareBackend) Save(l *share.Link) error {
	_, err := s.db.Exec(`
        INSERT INTO shares (
            hash, user_id, path, expire 
        ) VALUES (?, ?, ?, ?)
        ON DUPLICATE KEY UPDATE
            user_id = VALUES(user_id),
            path = VALUES(path),
            expire = VALUES(expire),
            password = VALUES(password),
            updated_at = VALUES(updated_at)
    `,
		l.Hash,
		l.UserID,
		l.Path,
		l.Expire,
	)
	return err
}

func (s shareBackend) Delete(hash string) error {
	_, err := s.db.Exec("DELETE FROM shares WHERE hash = ?", hash)
	return err
}
