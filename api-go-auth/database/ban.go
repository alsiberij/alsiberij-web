package database

import (
	"auth/models"
	"context"
	"github.com/jackc/pgtype/pgxtype"
)

type (
	Bans struct {
		conn pgxtype.Querier
	}
)

func NewBans(conn pgxtype.Querier) Bans {
	return Bans{conn: conn}
}

func (b *Bans) ByUserId(userId int64) (models.Ban, bool, error) {
	row, err := b.conn.Query(context.Background(), `SELECT "userId", "isActive", "activeUntil", "createdAt", "createdByUserId", reason FROM bans WHERE "userId" = $1 AND "isActive" IS TRUE`,
		userId)
	if err != nil {
		return models.Ban{}, false, err
	}

	var ban models.Ban
	var exists bool
	for row.Next() {
		exists = true
		err = row.Scan(&ban.UserId, &ban.IsActive, &ban.ActiveUntil, &ban.CreatedAt, &ban.CreatedByUserId, &ban.Reason)
	}

	return ban, exists, err
}

func (b *Bans) Expire(userId int64) (bool, error) {
	tag, err := b.conn.Exec(context.Background(), `UPDATE bans SET "isActive" = FALSE WHERE "userId" = $1 AND "isActive" IS TRUE`,
		userId)
	return tag.RowsAffected() > 0, err
}
