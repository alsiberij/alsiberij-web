package database

import (
	"auth/models"
	"context"
	"github.com/jackc/pgtype/pgxtype"
	"time"
)

type (
	RefreshTokens struct {
		conn pgxtype.Querier
	}
)

func (r *RefreshTokens) Create(userId int64, token string) error {
	if r.conn == nil {
		return ErrPostgresNotInitialized
	}

	_, err := r.conn.Exec(context.Background(), `INSERT INTO refresh_tokens("userId", token) VALUES ($1, $2)`,
		userId, token)
	return err
}

func (r *RefreshTokens) ByToken(token string, lifetime time.Duration) (models.RefreshTokenWithUserData, bool, error) {
	if r.conn == nil {
		return models.RefreshTokenWithUserData{}, false, ErrPostgresNotInitialized
	}

	lifetime = lifetime / time.Second
	rows, err := r.conn.Query(context.Background(),
		`SELECT t.id AS "tokenId", u.id AS "userId", u.role AS "userRole"
				FROM refresh_tokens as t
         		JOIN users AS u ON t."userId" = u.id
				WHERE token = $1 AND EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - "lastUsedAt")) < $2 AND "isRevoked" IS FALSE`,
		token, lifetime)
	if err != nil {
		return models.RefreshTokenWithUserData{}, false, err
	}

	var exists bool
	var tokenWithData models.RefreshTokenWithUserData
	for rows.Next() {
		exists = true
		err = rows.Scan(&tokenWithData.Id, &tokenWithData.UserId, &tokenWithData.UserRole)
	}

	_, err = r.conn.Exec(context.Background(),
		`UPDATE refresh_tokens SET "lastUsedAt" = CURRENT_TIMESTAMP WHERE token = $1`, token)

	return tokenWithData, exists, err
}

func (r *RefreshTokens) RevokeToken(token string) (int64, error) {
	if r.conn == nil {
		return 0, ErrPostgresNotInitialized
	}

	tag, err := r.conn.Exec(context.Background(),
		`UPDATE refresh_tokens SET "isRevoked" = TRUE WHERE token = $1 AND "isRevoked" IS FALSE`, token)
	return tag.RowsAffected(), err
}

func (r *RefreshTokens) RevokeAllTokens(token string) (int64, error) {
	if r.conn == nil {
		return 0, ErrPostgresNotInitialized
	}

	tag, err := r.conn.Exec(context.Background(),
		`UPDATE refresh_tokens SET "isRevoked" = TRUE WHERE "userId" = (SELECT "userId" FROM refresh_tokens WHERE token = $1) AND "isRevoked" IS FALSE`, token)
	return tag.RowsAffected(), err
}

func (r *RefreshTokens) RevokeAllTokensExceptOne(token string) (int64, error) {
	if r.conn == nil {
		return 0, ErrPostgresNotInitialized
	}

	tag, err := r.conn.Exec(context.Background(),
		`UPDATE refresh_tokens SET "isRevoked" = TRUE WHERE "userId" = (SELECT "userId" FROM refresh_tokens WHERE token = $1) AND token != $1 AND "isRevoked" IS FALSE`, token)
	return tag.RowsAffected(), err
}

func (r *RefreshTokens) RevokeAllByUserId(userId int64) (int64, error) {
	if r.conn == nil {
		return 0, ErrPostgresNotInitialized
	}

	tag, err := r.conn.Exec(context.Background(), `UPDATE refresh_tokens SET "isRevoked" = TRUE WHERE "userId" = $1 AND "isRevoked" IS FALSE`, userId)
	return tag.RowsAffected(), err
}
