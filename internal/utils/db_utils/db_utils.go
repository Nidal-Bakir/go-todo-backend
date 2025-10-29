package dbutils

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

func IsErrPgxNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func IsErrRedisNilNoRows(err error) bool {
	return errors.Is(err, redis.Nil)
}

func ToPgTypeText(str string) pgtype.Text {
	return pgtype.Text{String: str, Valid: len(str) != 0}
}

func ToPgTypeTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: !t.IsZero()}
}

func ToPgTypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: !t.IsZero()}
}

func ToPgTypeInt4(num *int32) pgtype.Int4 {
	if num == nil {
		return pgtype.Int4{Int32: -1, Valid: false}
	}
	return pgtype.Int4{Int32: int32(*num), Valid: true}
}
