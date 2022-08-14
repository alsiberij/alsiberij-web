package repository

import (
	"auth/models"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"github.com/jackc/pgtype/pgxtype"
)

type (
	Users struct {
		conn pgxtype.Querier
	}
)

func (r *Users) Create(email, login, password string) error {
	h := sha512.New()
	h.Write([]byte(password))
	password = hex.EncodeToString(h.Sum(nil))

	_, err := r.conn.Exec(context.Background(), `INSERT INTO users(email, login, password) VALUES ($1, $2, $3)`,
		email, login, password)
	if err != nil {
		return err
	}

	return nil
}

func (r *Users) ById(id int64) (models.User, bool, error) {
	row, err := r.conn.Query(context.Background(), `SELECT id, email, login, password, role, "createdAt", "isBanned" FROM users WHERE id = $1`,
		id)
	if err != nil {
		return models.User{}, false, err
	}

	var user models.User
	var exists bool
	for row.Next() {
		exists = true
		err = row.Scan(&user.Id, &user.Email, &user.Login, &user.Password, &user.Role, &user.CreatedAt, &user.IsBanned)
	}

	return user, exists, err
}

func (r *Users) AllShort() ([]models.UserShort, error) {
	row, err := r.conn.Query(context.Background(), `SELECT id, email, login, role, "createdAt", "isBanned" FROM users ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}

	var users []models.UserShort
	for row.Next() {
		var user models.UserShort
		err = row.Scan(&user.Id, &user.Email, &user.Login, &user.Role, &user.CreatedAt, &user.IsBanned)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *Users) ByCredentials(login, password string) (models.User, bool, error) {
	h := sha512.New()
	h.Write([]byte(password))
	password = hex.EncodeToString(h.Sum(nil))

	row, err := r.conn.Query(context.Background(), `SELECT id, email, login, password, role, "createdAt", "isBanned" FROM users WHERE login = $1 AND password = $2`,
		login, password)
	if err != nil {
		return models.User{}, false, err
	}

	var user models.User
	var exists bool
	for row.Next() {
		exists = true
		err = row.Scan(&user.Id, &user.Email, &user.Login, &user.Password, &user.Role, &user.CreatedAt, &user.IsBanned)
	}

	return user, exists, err
}

func (r *Users) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.conn.QueryRow(context.Background(), `SELECT EXISTS (SELECT FROM users WHERE email = $1)`, email).
		Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Users) LoginExists(login string) (bool, error) {
	var exists bool
	err := r.conn.QueryRow(context.Background(), `SELECT EXISTS (SELECT FROM users WHERE login = $1)`, login).
		Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Users) ChangeStatus(id int64, isBanned bool) (bool, error) {
	tag, err := r.conn.Exec(context.Background(), `UPDATE users SET "isBanned" = $1 WHERE id = $2`, isBanned, id)
	return tag.RowsAffected() > 0, err
}
