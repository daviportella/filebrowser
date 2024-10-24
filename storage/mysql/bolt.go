// db/mysql.go
package mysql

import (
	"database/sql"
	"log"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
	_ "github.com/go-sql-driver/mysql"
)

// NewStorage cria um storage.Storage baseado em MySQL
func NewStorage(db *sql.DB) (*storage.Storage, error) {
	// Inicializa os backends
	userStore := users.NewStorage(usersBackend{db: db})
	shareStore := share.NewStorage(shareBackend{db: db})
	settingsStore := settings.NewStorage(settingsBackend{db: db})
	authStore := auth.NewStorage(authBackend{db: db}, userStore)

	// Verifica/cria a versão na tabela
	if err := initializeTables(db); err != nil {
		log.Println("Error initializing tables: " + err.Error())
		return nil, err
	}

	return &storage.Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
	}, nil
}

func initializeTables(db *sql.DB) error {
	// SQL para criar as tabelas necessárias
	queries := []string{
		`CREATE TABLE IF NOT EXISTS config (
            key_name VARCHAR(50) PRIMARY KEY,
            value JSON NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS users (
            id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
            username VARCHAR(100) UNIQUE NOT NULL,
            password VARCHAR(100) NOT NULL,
            fs_path VARCHAR(255) NOT NULL,
            perm JSON,
            commands JSON,
            sorting JSON,
            locale VARCHAR(50),
            single_click BOOLEAN,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS settings (
            id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
            key_name VARCHAR(50) UNIQUE NOT NULL,
            value JSON NOT NULL
        )`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Println("Error initializing tables: " + err.Error())
			return err
		}
	}

	// Insere a versão inicial
	_, err := db.Exec("INSERT INTO config (key_name, value) VALUES ('version', JSON_OBJECT('version', 2)) ON DUPLICATE KEY UPDATE value = VALUES(value)")
	return err
}
