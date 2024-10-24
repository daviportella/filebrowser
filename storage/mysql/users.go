// db/users.go
package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	fbErrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/users"
)

type usersBackend struct {
	db *sql.DB
}

func (st usersBackend) GetBy(i interface{}) (*users.User, error) {
	user := &users.User{}
	var query string
	var arg interface{}

	switch v := i.(type) {
	case uint:
		query = "SELECT * FROM users WHERE id = ?"
		arg = v
	case string:
		query = "SELECT * FROM users WHERE username = ?"
		arg = v
	default:
		return nil, fbErrors.ErrInvalidDataType
	}

	row := st.db.QueryRow(query, arg)

	var permJSON, commandsJSON, sortingJSON []byte
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Scope,
		&permJSON,
		&commandsJSON,
		&sortingJSON,
		&user.Locale,
		&user.SingleClick,
	)

	if err == sql.ErrNoRows {
		return nil, fbErrors.ErrNotExist
	}
	if err != nil {
		return nil, err
	}

	// Deserializa os campos JSON
	if err := json.Unmarshal(permJSON, &user.Perm); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(commandsJSON, &user.Commands); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(sortingJSON, &user.Sorting); err != nil {
		return nil, err
	}

	return user, nil
}

func (st usersBackend) Gets() ([]*users.User, error) {
	query := "SELECT * FROM users"
	rows, err := st.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allUsers []*users.User
	for rows.Next() {
		user := &users.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return allUsers, nil
}

// func (st usersBackend) Gets() ([]*users.User, error) {
// 	rows, err := st.db.Query("SELECT * FROM users")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var users []*users.User
// 	for rows.Next() {
// 		user := &users.User{}
// 		var permJSON, commandsJSON, sortingJSON []byte

// 		err := rows.Scan(
// 			&user.ID,
// 			&user.Username,
// 			&user.Password,
// 			&user.Scope,
// 			&permJSON,
// 			&commandsJSON,
// 			&sortingJSON,
// 			&user.Locale,
// 			&user.SingleClick,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}

// 		// Deserializa os campos JSON
// 		if err := json.Unmarshal(permJSON, &user.Perm); err != nil {
// 			return nil, err
// 		}
// 		if err := json.Unmarshal(commandsJSON, &user.Commands); err != nil {
// 			return nil, err
// 		}
// 		if err := json.Unmarshal(sortingJSON, &user.Sorting); err != nil {
// 			return nil, err
// 		}

// 		users = append(users, user)
// 	}

// 	return users, nil
// }

func (st usersBackend) Save(user *users.User) error {
	// Serializa os campos estruturados para JSON
	permJSON, err := json.Marshal(user.Perm)
	if err != nil {
		return err
	}
	commandsJSON, err := json.Marshal(user.Commands)
	if err != nil {
		return err
	}
	sortingJSON, err := json.Marshal(user.Sorting)
	if err != nil {
		return err
	}

	query := `
        INSERT INTO users (
            username, password, fs_path, perm, commands, sorting, locale, single_click
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ON DUPLICATE KEY UPDATE
            password = VALUES(password),
            fs_path = VALUES(fs_path),
            perm = VALUES(perm),
            commands = VALUES(commands),
            sorting = VALUES(sorting),
            locale = VALUES(locale),
            single_click = VALUES(single_click)
    `

	_, err = st.db.Exec(query,
		user.Username,
		user.Password,
		user.Scope,
		permJSON,
		commandsJSON,
		sortingJSON,
		user.Locale,
		user.SingleClick,
	)

	if err != nil {
		return err
	}

	return nil
}

func (st usersBackend) DeleteByID(id uint) error {
	_, err := st.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func (st usersBackend) DeleteByUsername(username string) error {
	_, err := st.db.Exec("DELETE FROM users WHERE username = ?", username)
	return err
}

func (st usersBackend) Update(user *users.User, fields ...string) error {
	if len(fields) == 0 {
		return st.Save(user)
	}

	updates := make([]string, 0)
	args := make([]interface{}, 0)

	for _, field := range fields {
		updates = append(updates, fmt.Sprintf("%s = ?", field))
		switch field {
		case "Perm":
			j, err := json.Marshal(user.Perm)
			if err != nil {
				return err
			}
			args = append(args, j)
		case "Commands":
			j, err := json.Marshal(user.Commands)
			if err != nil {
				return err
			}
			args = append(args, j)
		case "Sorting":
			j, err := json.Marshal(user.Sorting)
			if err != nil {
				return err
			}
			args = append(args, j)
		default:
			val := reflect.ValueOf(user).Elem().FieldByName(field)
			if !val.IsValid() {
				return fmt.Errorf("invalid field: %s", field)
			}
			args = append(args, val.Interface())
		}
	}

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(updates, ", "))
	args = append(args, user.ID)

	_, err := st.db.Exec(query, args...)
	return err
}
