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
	row, err := b.conn.Query(context.Background(),
		`SELECT "bannedUserId", reason, "activeUntil", "createdByUserId", "createdAt"
			FROM bans
			WHERE "bannedUserId" = $1 AND "activeUntil">CURRENT_TIMESTAMP`,
		userId)
	if err != nil {
		return models.Ban{}, false, err
	}

	var ban models.Ban
	var exists bool
	for row.Next() {
		exists = true
		err = row.Scan(&ban.BannedUserId, &ban.Reason, &ban.ActiveUntil, &ban.CreatedByUserId, &ban.CreatedAt)
	}

	return ban, exists, err
}

func (b *Bans) ExpireByUserId(userId int64) (bool, error) {
	tag, err := b.conn.Exec(context.Background(), `UPDATE bans SET "activeUntil" = CURRENT_TIMESTAMP WHERE "bannedUserId" = $1 AND "activeUntil">CURRENT_TIMESTAMP`,
		userId)
	return tag.RowsAffected() > 0, err
}
