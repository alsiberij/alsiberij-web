package repository

import (
	"auth/models"
	"context"
	"github.com/jackc/pgtype/pgxtype"
)

type (
	RefreshTokenRepository interface {
		Create(userId int64, token string) error
		ById(id int64) (models.RefreshToken, bool, error)
		ByUserId(userId int64) ([]models.RefreshToken, error)
		ByToken(token string) (models.RefreshToken, bool, error)
		All() ([]models.RefreshToken, error)
		SetExpired(id int64) error
		UpdateLastUsageTime(token string) error
		Delete(id int64) error
	}

	RefreshTokenPostgresRepository struct {
		conn pgxtype.Querier
	}
)

func (r *RefreshTokenPostgresRepository) Create(userId int64, token string) error {
	_, err := r.conn.Exec(context.Background(), `INSERT INTO refresh_tokens("userId", token) VALUES ($1, $2)`,
		userId, token)
	return err
}

func (r *RefreshTokenPostgresRepository) ById(id int64) (models.RefreshToken, bool, error) {
	rows, err := r.conn.Query(context.Background(),
		`SELECT t.id, t.token, t."isExpired", t."issuedAt", t."lastUsedAt",
       			u.id, u.email, u.login, u.password, u.role, u."createdAt"
			FROM refresh_tokens AS t JOIN users AS u ON t."userId" = u.id
			WHERE t.id = $1`, id)
	if err != nil {
		return models.RefreshToken{}, false, err
	}

	var refreshToken models.RefreshToken
	var exists bool
	for rows.Next() {
		exists = true
		err = rows.Scan(&refreshToken.Id, &refreshToken.Token, &refreshToken.IsExpired,
			&refreshToken.IssuedAt, &refreshToken.LastUsedAt, &refreshToken.User.Id, &refreshToken.User.Email,
			&refreshToken.User.Login, &refreshToken.User.Password, &refreshToken.User.Role, &refreshToken.User.CreatedAt)
		if err != nil {
			return models.RefreshToken{}, true, err
		}
	}
	return refreshToken, exists, err
}

func (r *RefreshTokenPostgresRepository) ByUserId(userId int64) ([]models.RefreshToken, error) {
	rows, err := r.conn.Query(context.Background(),
		`SELECT t.id, t.token, t."isExpired", t."issuedAt", t."lastUsedAt",
       			u.id, u.email, u.login, u.password, u.role, u."createdAt"
			FROM refresh_tokens AS t JOIN users AS u ON t."userId" = u.id
			WHERE u.id = $1`, userId)
	if err != nil {
		return nil, err
	}

	var refreshTokens []models.RefreshToken
	for rows.Next() {
		var refreshToken models.RefreshToken
		err = rows.Scan(&refreshToken.Id, &refreshToken.Token, &refreshToken.IsExpired,
			&refreshToken.IssuedAt, &refreshToken.LastUsedAt, &refreshToken.User.Id, &refreshToken.User.Email,
			&refreshToken.User.Login, &refreshToken.User.Password, &refreshToken.User.Role, &refreshToken.User.CreatedAt)
		if err != nil {
			return nil, err
		}
		refreshTokens = append(refreshTokens, refreshToken)
	}
	return refreshTokens, nil
}

func (r *RefreshTokenPostgresRepository) ByToken(token string) (models.RefreshToken, bool, error) {
	rows, err := r.conn.Query(context.Background(),
		`SELECT t.id, t.token, t."isExpired", t."issuedAt", t."lastUsedAt",
       			u.id, u.email, u.login, u.password, u.role, u."createdAt"
			FROM refresh_tokens AS t JOIN users AS u ON t."userId" = u.id
			WHERE t.token = $1`, token)
	if err != nil {
		return models.RefreshToken{}, false, err
	}

	var refreshToken models.RefreshToken
	var exists bool
	for rows.Next() {
		exists = true
		err = rows.Scan(&refreshToken.Id, &refreshToken.Token, &refreshToken.IsExpired,
			&refreshToken.IssuedAt, &refreshToken.LastUsedAt, &refreshToken.User.Id, &refreshToken.User.Email,
			&refreshToken.User.Login, &refreshToken.User.Password, &refreshToken.User.Role, &refreshToken.User.CreatedAt)
		if err != nil {
			return models.RefreshToken{}, true, err
		}
	}
	return refreshToken, exists, err
}

func (r *RefreshTokenPostgresRepository) All() ([]models.RefreshToken, error) {
	rows, err := r.conn.Query(context.Background(),
		`SELECT t.id, t.token, t."isExpired", t."issuedAt", t."lastUsedAt",
       			u.id, u.email, u.login, u.password, u.role, u."createdAt"
			FROM refresh_tokens AS t JOIN users AS u ON t."userId" = u.id`)
	if err != nil {
		return nil, err
	}

	var refreshTokens []models.RefreshToken
	for rows.Next() {
		var refreshToken models.RefreshToken
		err = rows.Scan(&refreshToken.Id, &refreshToken.Token, &refreshToken.IsExpired,
			&refreshToken.IssuedAt, &refreshToken.LastUsedAt, &refreshToken.User.Id, &refreshToken.User.Email,
			&refreshToken.User.Login, &refreshToken.User.Password, &refreshToken.User.Role, &refreshToken.User.CreatedAt)
		if err != nil {
			return nil, err
		}
		refreshTokens = append(refreshTokens, refreshToken)
	}
	return refreshTokens, nil
}

func (r *RefreshTokenPostgresRepository) SetExpired(id int64) error {
	_, err := r.conn.Exec(context.Background(), `UPDATE refresh_tokens SET "isExpired" = TRUE WHERE id = $1`, id)
	return err
}

func (r *RefreshTokenPostgresRepository) UpdateLastUsageTime(token string) error {
	_, err := r.conn.Exec(context.Background(), `UPDATE refresh_tokens SET "lastUsedAt" = CURRENT_TIMESTAMP WHERE token = $1`,
		token)
	return err
}

func (r *RefreshTokenPostgresRepository) Delete(id int64) error {
	_, err := r.conn.Query(context.Background(), `DELETE FROM refresh_tokens WHERE id = $1`, id)
	return err
}
