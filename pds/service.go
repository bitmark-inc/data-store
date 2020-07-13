package pds

import (
	"github.com/bitmark-inc/data-store/store"
)

type PDS struct {
	dataStorePool store.DataStorePool
}

func New(pool store.DataStorePool) *PDS {
	return &PDS{
		dataStorePool: pool,
	}
}
