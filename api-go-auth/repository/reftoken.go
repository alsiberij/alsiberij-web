package repository

import (
	"auth/models"
	"context"
	"github.com/jackc/pgtype/pgxtype"
)

type (
	RefreshTokens struct {
		conn pgxtype.Querier
	}
)

func (r *RefreshTokens) Create(userId int64, token string) error {
	_, err := r.conn.Exec(context.Background(), `INSERT INTO refresh_tokens("userId", token) VALUES ($1, $2)`,
		userId, token)
	return err
}

func (r *RefreshTokens) ByToken(token string) (models.RefreshToken, bool, error) {
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

func (r *RefreshTokens) ByTokenNotExpired(token string) (models.RefreshToken, bool, error) {
	rows, err := r.conn.Query(context.Background(),
		`SELECT t.id, t.token, t."isExpired", t."issuedAt", t."lastUsedAt",
       			u.id, u.email, u.login, u.password, u.role, u."createdAt"
			FROM refresh_tokens AS t JOIN users AS u ON t."userId" = u.id
			WHERE t.token = $1 AND "isExpired" IS FALSE`, token)
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

func (r *RefreshTokens) SetExpired(id int64) (bool, error) {
	tag, err := r.conn.Exec(context.Background(), `UPDATE refresh_tokens SET "isExpired" = TRUE WHERE id = $1`, id)
	return tag.RowsAffected() > 0, err
}

func (r *RefreshTokens) UpdateLastUsageTime(token string) (bool, error) {
	tag, err := r.conn.Exec(context.Background(), `UPDATE refresh_tokens SET "lastUsedAt" = CURRENT_TIMESTAMP WHERE token = $1`,
		token)
	return tag.RowsAffected() > 0, err
}

func (r *RefreshTokens) SetExpiredByUserId(userId int64) (bool, error) {
	tag, err := r.conn.Exec(context.Background(), `UPDATE refresh_tokens SET "isExpired" = TRUE WHERE "userId" = $1`, userId)
	return tag.RowsAffected() > 0, err
}

func (r *RefreshTokens) SetExpiredByToken(token string) (bool, error) {
	tag, err := r.conn.Exec(context.Background(), `UPDATE refresh_tokens SET "isExpired" = TRUE WHERE token = $1`, token)
	return tag.RowsAffected() > 0, err
}

func (r *RefreshTokens) SetExpiredByTokenBelongingUser(token string) (bool, error) {
	tag, err := r.conn.Exec(context.Background(), `UPDATE refresh_tokens SET "isExpired" = TRUE WHERE "userId" = (SELECT "userId" FROM refresh_tokens WHERE token = $1)`, token)
	return tag.RowsAffected() > 0, err
}

func (r *RefreshTokens) SetExpiredByTokenBelongingUserExceptCurrent(token string) (bool, error) {
	tag, err := r.conn.Exec(context.Background(), `UPDATE refresh_tokens SET "isExpired" = TRUE WHERE "userId" = (SELECT "userId" FROM refresh_tokens WHERE token = $1) AND token != $1`, token)
	return tag.RowsAffected() > 0, err
}
