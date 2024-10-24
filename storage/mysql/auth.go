package mysql

import (
	"database/sql"
	"encoding/json"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
)

type authBackend struct {
	db *sql.DB
}

func (s authBackend) Get(t settings.AuthMethod) (auth.Auther, error) {
	var auther auth.Auther

	// switch t {
	// case auth.MethodJSONAuth:
	// 	auther = &auth.JSONAuth{}
	// case auth.MethodProxyAuth:
	// 	auther = &auth.ProxyAuth{}
	// case auth.MethodHookAuth:
	// 	auther = &auth.HookAuth{}
	// case auth.MethodNoAuth:
	// 	auther = &auth.NoAuth{}
	// default:
	// 	return nil, errors.ErrInvalidAuthMethod
	// }

	auther = &auth.JSONAuth{}

	var data []byte
	err := s.db.QueryRow("SELECT value FROM config WHERE key_name = 'auther'").Scan(&data)
	if err == sql.ErrNoRows {
		return auther, errors.ErrNotExist
	}
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, auther)
	return auther, err
}

func (s authBackend) Save(a auth.Auther) error {
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
        INSERT INTO config (key_name, value) 
        VALUES ('auther', ?) 
        ON DUPLICATE KEY UPDATE value = VALUES(value)
    `, data)
	return err
}
