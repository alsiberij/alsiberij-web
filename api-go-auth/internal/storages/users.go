package storages

import (
	"auth/internal/models"
	"auth/pkg/pgs"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"github.com/jackc/pgtype/pgxtype"
)

//TODO context

type (
	UserStorage struct {
		querier pgxtype.Querier
	}
)

func NewUserStorage(q pgxtype.Querier) models.UserStorage {
	return &UserStorage{querier: q}
}

func (r *UserStorage) CreateAndStore(email, login, password string) error {
	if r.querier == nil {
		return pgs.ErrNotInitialized
	}

	h := sha512.New()
	h.Write([]byte(password))
	password = hex.EncodeToString(h.Sum(nil))

	_, err := r.querier.Exec(context.Background(), `INSERT INTO users(email, login, password) VALUES ($1, $2, $3)`,
		email, login, password)
	return err
}

func (r *UserStorage) GetByCredentials(credentials models.UserCredentials) (*models.User, error) {
	if r.querier == nil {
		return nil, pgs.ErrNotInitialized
	}

	h := sha512.New()
	h.Write([]byte(credentials.Password))
	passwordHash := hex.EncodeToString(h.Sum(nil))

	rows, err := r.querier.Query(context.Background(), `SELECT id, email, login, password, role, "createdAt" FROM users WHERE login = $1 AND password = $2`,
		credentials.Login, passwordHash)
	if err != nil {
		return nil, err
	}

	var user *models.User
	for rows.Next() {
		user = &models.User{}
		err = rows.Scan(&user.Id, &user.Email, &user.Login, &user.Password, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (r *UserStorage) GetById(userId int64) (*models.User, error) {
	if r.querier == nil {
		return nil, pgs.ErrNotInitialized
	}

	rows, err := r.querier.Query(context.Background(), `SELECT id, email, login, password, role, "createdAt" FROM users WHERE id = $1`, userId)
	if err != nil {
		return nil, err
	}

	var user *models.User
	for rows.Next() {
		user = &models.User{}
		err = rows.Scan(&user.Id, &user.Email, &user.Login, &user.Password, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (r *UserStorage) EmailExists(email string) (bool, error) {
	if r.querier == nil {
		return false, pgs.ErrNotInitialized
	}

	var exists bool
	err := r.querier.QueryRow(context.Background(), `SELECT EXISTS(SELECT FROM users WHERE email = $1)`, email).
		Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserStorage) LoginExists(login string) (bool, error) {
	if r.querier == nil {
		return false, pgs.ErrNotInitialized
	}

	var exists bool
	err := r.querier.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT FROM users WHERE login = $1)`, login).
		Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserStorage) ChangeRole(userId int64, newRole string) error {
	if r.querier == nil {
		return pgs.ErrNotInitialized
	}

	_, err := r.querier.Exec(context.Background(), `UPDATE users SET role = $1 WHERE id = $2`, newRole, userId)
	return err
}
