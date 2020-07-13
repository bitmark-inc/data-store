package cds

import (
	"github.com/bitmark-inc/data-store/store"
)

type CDS struct {
	dataStorePool store.DataStorePool
}

func New(pool store.DataStorePool) *CDS {
	return &CDS{
		dataStorePool: pool,
	}
}
