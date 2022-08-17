package database

import (
	"auth/models"
	"context"
	"github.com/jackc/pgtype/pgxtype"
	"time"
)

type (
	Bans struct {
		conn pgxtype.Querier
	}
)

func NewBans(conn pgxtype.Querier) Bans {
	return Bans{conn: conn}
}

func (r *Bans) Create(bannedUserId int64, reason string, activeUntil time.Time, createdByUserId int64) error {
	_, err := r.conn.Exec(context.Background(),
		`INSERT INTO bans("bannedUserId", reason, "activeUntil", "createdByUserId") VALUES ($1, $2, $3, $4)`,
		bannedUserId, reason, activeUntil, createdByUserId)
	return err
}

func (r *Bans) ActiveByUserId(userId int64) (models.Ban, bool, error) {
	row, err := r.conn.Query(context.Background(),
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

func (r *Bans) ActiveExistsByUserId(userId int64) (bool, error) {
	var exists bool
	err := r.conn.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT FROM bans WHERE "bannedUserId" = $1 AND "activeUntil">CURRENT_TIMESTAMP)`,
		userId).Scan(&exists)
	return exists, err
}
