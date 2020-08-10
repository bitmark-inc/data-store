package cds

import (
	"github.com/bitmark-inc/data-store/notification"
	"github.com/bitmark-inc/data-store/store"
)

type CDS struct {
	dataStorePool      store.DataStorePool
	notificationClient *notification.Client
}

func New(pool store.DataStorePool, client *notification.Client) *CDS {
	return &CDS{
		dataStorePool:      pool,
		notificationClient: client,
	}
}
