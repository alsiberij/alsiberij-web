package repository

import (
	"auth/models"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"github.com/jackc/pgtype/pgxtype"
)

type (
	UserRepository interface {
		Create(email, login, password string) error
		Get(id int64) (models.User, bool, error)
		All() ([]models.User, error)
		AllShort() ([]models.UserShort, error)
		GetByCredentials(login, password string) (models.User, bool, error)
		Delete(id int64) error
		EmailExists(email string) (bool, error)
		LoginExists(login string) (bool, error)
	}

	UserPostgresRepository struct {
		conn pgxtype.Querier
	}
)

func (r *UserPostgresRepository) Create(email, login, password string) error {
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

func (r *UserPostgresRepository) Get(id int64) (models.User, bool, error) {
	row, err := r.conn.Query(context.Background(), `SELECT id, email, login, password, role, "createdAt" FROM users WHERE id = $1`,
		id)
	if err != nil {
		return models.User{}, false, err
	}

	var user models.User
	var exists bool
	for row.Next() {
		exists = true
		err = row.Scan(&user.Id, &user.Email, &user.Login, &user.Password, &user.Role, &user.CreatedAt)
	}

	return user, exists, err
}

func (r *UserPostgresRepository) All() ([]models.User, error) {
	row, err := r.conn.Query(context.Background(), `SELECT id, email, login, password, role, "createdAt" FROM users`)
	if err != nil {
		return nil, err
	}

	var users []models.User
	for row.Next() {
		var user models.User
		err = row.Scan(&user.Id, &user.Email, &user.Login, &user.Password, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserPostgresRepository) AllShort() ([]models.UserShort, error) {
	row, err := r.conn.Query(context.Background(), `SELECT id, email, login, role, "createdAt" FROM users`)
	if err != nil {
		return nil, err
	}

	var users []models.UserShort
	for row.Next() {
		var user models.UserShort
		err = row.Scan(&user.Id, &user.Email, &user.Login, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserPostgresRepository) GetByCredentials(login, password string) (models.User, bool, error) {
	h := sha512.New()
	h.Write([]byte(password))
	password = hex.EncodeToString(h.Sum(nil))

	row, err := r.conn.Query(context.Background(), `SELECT id, email, login, password, role, "createdAt" FROM users WHERE login = $1 AND password = $2`,
		login, password)
	if err != nil {
		return models.User{}, false, err
	}

	var user models.User
	var exists bool
	for row.Next() {
		exists = true
		err = row.Scan(&user.Id, &user.Email, &user.Login, &user.Password, &user.Role, &user.CreatedAt)
	}

	return user, exists, err
}

func (r *UserPostgresRepository) Delete(id int64) error {
	_, err := r.conn.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *UserPostgresRepository) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.conn.QueryRow(context.Background(), `SELECT EXISTS (SELECT FROM users WHERE email = $1)`, email).
		Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserPostgresRepository) LoginExists(login string) (bool, error) {
	var exists bool
	err := r.conn.QueryRow(context.Background(), `SELECT EXISTS (SELECT FROM users WHERE login = $1)`, login).
		Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
