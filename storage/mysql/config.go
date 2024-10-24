// db/config.go
package mysql

import (
	"database/sql"
	"encoding/json"

	"github.com/filebrowser/filebrowser/v2/settings"
)

type settingsBackend struct {
	db *sql.DB
}

func (s settingsBackend) Get() (*settings.Settings, error) {
	set := &settings.Settings{}
	var data []byte

	err := s.db.QueryRow("SELECT value FROM settings WHERE key_name = 'settings'").Scan(&data)
	if err == sql.ErrNoRows {
		return set, nil
	}
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, set)
	return set, err
}

func (s settingsBackend) Save(set *settings.Settings) error {
	data, err := json.Marshal(set)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
        INSERT INTO settings (key_name, value) 
        VALUES ('settings', ?) 
        ON DUPLICATE KEY UPDATE value = VALUES(value)
    `, data)

	return err
}

func (s settingsBackend) GetServer() (*settings.Server, error) {
	server := &settings.Server{}
	var data []byte

	err := s.db.QueryRow("SELECT value FROM settings WHERE key_name = 'server'").Scan(&data)
	if err == sql.ErrNoRows {
		return server, nil
	}
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, server)
	return server, err
}

func (s settingsBackend) SaveServer(server *settings.Server) error {
	data, err := json.Marshal(server)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
        INSERT INTO settings (key_name, value) 
        VALUES ('server', ?) 
        ON DUPLICATE KEY UPDATE value = VALUES(value)
    `, data)

	return err
}
