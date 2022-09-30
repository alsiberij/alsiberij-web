package storages

import (
	"auth/internal/models"
	"auth/pkg/pgs"
	"context"
	"github.com/jackc/pgtype/pgxtype"
	"time"
)

//TODO context

type (
	RefreshTokenStorage struct {
		querier pgxtype.Querier
	}
)

func NewRefreshTokenStorage(q pgxtype.Querier) models.RefreshTokenStorage {
	return &RefreshTokenStorage{querier: q}
}

func (r *RefreshTokenStorage) CreateAndStore(userId int64, tokenValue string) error {
	if r.querier == nil {
		return pgs.ErrNotInitialized
	}

	_, err := r.querier.Exec(context.Background(), `INSERT INTO refresh_tokens("userId", token) VALUES ($1, $2)`,
		userId, tokenValue)

	return err
}

func (r *RefreshTokenStorage) Get(tokenValue string, lifePeriod time.Duration) (*models.RefreshToken, error) {
	if r.querier == nil {
		return nil, pgs.ErrNotInitialized
	}

	lifePeriod = lifePeriod / time.Second
	rows, err := r.querier.Query(context.Background(),
		`SELECT u.id, u.email, u.login, u.password, u.role, u."createdAt",
       			t.token, t."issuedAt", t."lastUsedAt", t."isRevoked"
				FROM refresh_tokens AS t JOIN users AS u ON t."userId" = u.id
				WHERE t.token = $1 AND EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - t."lastUsedAt")) < $2 AND t."isRevoked" IS FALSE`,
		tokenValue, lifePeriod)
	if err != nil {
		return nil, err
	}

	var refreshToken *models.RefreshToken
	for rows.Next() {
		refreshToken = &models.RefreshToken{}
		err = rows.Scan(&refreshToken.User.Id, &refreshToken.User.Email, &refreshToken.User.Login,
			&refreshToken.User.Password, &refreshToken.User.Role, &refreshToken.User.CreatedAt, &refreshToken.Token,
			&refreshToken.IssuedAt, &refreshToken.LastUsedAt, &refreshToken.IsRevoked)
		if err != nil {
			return nil, err
		}
	}

	_, err = r.querier.Exec(context.Background(),
		`UPDATE refresh_tokens SET "lastUsedAt" = CURRENT_TIMESTAMP WHERE token = $1`, tokenValue)
	if err != nil {
		return nil, err
	}

	return refreshToken, err
}

func (r *RefreshTokenStorage) Revoke(tokenValue string) error {
	if r.querier == nil {
		return pgs.ErrNotInitialized
	}

	_, err := r.querier.Exec(context.Background(),
		`UPDATE refresh_tokens SET "isRevoked" = TRUE WHERE token = $1 AND "isRevoked" IS FALSE`, tokenValue)
	return err
}

func (r *RefreshTokenStorage) RevokeAll(tokenValue string) error {
	if r.querier == nil {
		return pgs.ErrNotInitialized
	}

	_, err := r.querier.Exec(context.Background(),
		`UPDATE refresh_tokens SET "isRevoked" = TRUE WHERE "userId" = (SELECT "userId" FROM refresh_tokens WHERE token = $1) AND "isRevoked" IS FALSE`, tokenValue)
	return err
}

func (r *RefreshTokenStorage) RevokeAllExceptCurrent(tokenValue string) error {
	if r.querier == nil {
		return pgs.ErrNotInitialized
	}

	_, err := r.querier.Exec(context.Background(),
		`UPDATE refresh_tokens SET "isRevoked" = TRUE WHERE "userId" = (SELECT "userId" FROM refresh_tokens WHERE token = $1) AND token != $1 AND "isRevoked" IS FALSE`, tokenValue)
	return err
}

func (r *RefreshTokenStorage) RevokeAllByUserId(userId int64) error {
	if r.querier == nil {
		return pgs.ErrNotInitialized
	}

	_, err := r.querier.Exec(context.Background(), `UPDATE refresh_tokens SET "isRevoked" = TRUE WHERE "userId" = $1 AND "isRevoked" IS FALSE`, userId)
	return err
}
