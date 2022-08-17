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

func NewUsers(conn pgxtype.Querier) Users {
	return Users{conn: conn}
}

func (r *Users) Create(email, login, password string) error {
	h := sha512.New()
	h.Write([]byte(password))
	password = hex.EncodeToString(h.Sum(nil))

	_, err := r.conn.Exec(context.Background(), `INSERT INTO users(email, login, password) VALUES ($1, $2, $3)`,
		email, login, password)
	return err
}

func (r *Users) IdByCredentials(login, password string) (int64, bool, error) {
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
	var exists bool
	err := r.conn.QueryRow(context.Background(), `SELECT EXISTS (SELECT FROM users WHERE email = $1)`, email).
		Scan(&exists)
	return exists, err
}

func (r *Users) LoginExists(login string) (bool, error) {
	var exists bool
	err := r.conn.QueryRow(context.Background(), `SELECT EXISTS (SELECT FROM users WHERE login = $1)`, login).
		Scan(&exists)
	return exists, err
}
