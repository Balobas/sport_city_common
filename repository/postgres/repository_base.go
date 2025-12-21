package repositoryBasePostgres

import (
	clientDB "github.com/balobas/sport_city_common/clients/database"
)

type BasePgRepository struct {
	clientDB.QueryExecer
}

func New(client clientDB.ClientDB) *BasePgRepository {
	return &BasePgRepository{
		QueryExecer: client,
	}
}
