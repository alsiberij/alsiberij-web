package srv

import "github.com/jackc/pgtype/pgxtype"

func IsBanned(q pgxtype.Querier, userId int64) (bool, error) {
	bansRep := PostgresAuth.Bans(q)

	ban, exists, err := bansRep.ByUserId(userId)
	if err != nil {
		return true, err
	}

	if !exists {
		return false, nil
	}
}
