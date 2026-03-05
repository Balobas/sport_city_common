package repositoryBaseEntityPostgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	uuid "github.com/satori/go.uuid"
)

func PgUidFromUUID(uid uuid.UUID) pgtype.UUID {
	if uuid.Equal(uid, uuid.UUID{}) {
		return pgtype.UUID{}
	}
	return pgtype.UUID{
		Bytes: uid,
		Valid: true,
	}
}

func PgUtcTimestampFromTime(t time.Time) pgtype.Timestamp {
	if t.IsZero() || t.Equal(time.Time{}) {
		return pgtype.Timestamp{}
	}
	return pgtype.Timestamp{
		Time:  t.UTC(),
		Valid: true,
	}
}

func PgUidsFromUUIDs(uids []uuid.UUID) []pgtype.UUID {
	res := make([]pgtype.UUID, len(uids))
	for i := 0; i < len(uids); i++ {
		res[i] = PgUidFromUUID(uids[i])
	}
	return res
}

func UidsFromPgUids(uids []pgtype.UUID) []uuid.UUID {
	res := make([]uuid.UUID, len(uids))
	for i := 0; i < len(uids); i++ {
		res[i] = uids[i].Bytes
	}
	return res
}
