package database

import (
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
	if r.conn == nil {
		return ErrPostgresNotInitialized
	}

	h := sha512.New()
	h.Write([]byte(password))
	password = hex.EncodeToString(h.Sum(nil))

	_, err := r.conn.Exec(context.Background(), `INSERT INTO users(email, login, password) VALUES ($1, $2, $3)`,
		email, login, password)
	return err
}

func (r *Users) IdByCredentials(login, password string) (int64, bool, error) {
	if r.conn == nil {
		return 0, false, ErrPostgresNotInitialized
	}

	h := sha512.New()
	h.Write([]byte(password))
	password = hex.EncodeToString(h.Sum(nil))

	row, err := r.conn.Query(context.Background(), `SELECT id FROM users WHERE login = $1 AND password = $2`,
		login, password)
	if err != nil {
		return 0, false, err
	}

	var userId int64
	var exists bool
	for row.Next() {
		exists = true
		err = row.Scan(&userId)
	}

	return userId, exists, err
}

func (r *Users) RoleById(userId int64) (string, bool, error) {
	if r.conn == nil {
		return "", false, ErrPostgresNotInitialized
	}

	row, err := r.conn.Query(context.Background(), `SELECT role FROM users WHERE id = $1`, userId)
	if err != nil {
		return "", false, err
	}

	var role string
	var exists bool
	for row.Next() {
		exists = true
		err = row.Scan(&role)
	}

	return role, exists, err
}

func (r *Users) EmailExists(email string) (bool, error) {
	if r.conn == nil {
		return false, ErrPostgresNotInitialized
	}

	var exists bool
	err := r.conn.QueryRow(context.Background(), `SELECT EXISTS (SELECT FROM users WHERE email = $1)`, email).
		Scan(&exists)
	return exists, err
}

func (r *Users) LoginAndEmailExists(login, email string) (bool, bool, error) {
	if r.conn == nil {
		return false, false, ErrPostgresNotInitialized
	}

	var existsLogin, existsEmail bool
	err := r.conn.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT FROM users WHERE login = $1), EXISTS(SELECT FROM users WHERE email = $2)`, login, email).
		Scan(&existsLogin, &existsEmail)
	return existsLogin, existsEmail, err
}

func (r *Users) UpdateRoleById(userId int64, role string) error {
	if r.conn == nil {
		return ErrPostgresNotInitialized
	}

	_, err := r.conn.Exec(context.Background(), `UPDATE users SET role = $1 WHERE id = $2`, role, userId)
	return err
}
