package store

import (
	"context"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	mongoLogPrefix = "mongo"
)

type DataStorePool interface {
	// Account will return a personal data store for a specific account number.
	Account(accountNumber string) PersonalDataStore
}

type PersonalDataStore interface {
	RatePOIResource(ctx context.Context, poiID string, ratings map[string]float64) error
	GetPOIResource(ctx context.Context, poiID string) (map[string]float64, error)
}

// mongodbDataPool is an implementation of DataStorePool.
type mongodbDataPool struct {
	client *mongo.Client
}

// NewMongodbDataPool returns a mongodbDataPool instance
func NewMongodbDataPool(client *mongo.Client) *mongodbDataPool {
	return &mongodbDataPool{
		client: client,
	}
}

// Account returns a personal data store.
func (m mongodbDataPool) Account(accountNumber string) PersonalDataStore {
	return &mongoAccountStore{
		accountNumber: accountNumber,
		db:            m.client.Database(accountNumber),
	}
}

// Ping is to make a ping call to mongodb
func (m mongodbDataPool) Ping() error {
	return m.client.Ping(context.Background(), nil)
}

// Close it to close the mongodb connection for the instance
func (m mongodbDataPool) Close() {
	log.WithField("prefix", mongoLogPrefix).Info("closing mongo db connections")
	_ = m.client.Disconnect(context.Background())
}

type mongoAccountStore struct {
	accountNumber string
	db            *mongo.Database
}

// Resource returns the collection of the given resource from the database
// of the linked account number
func (m mongoAccountStore) Resource(name string) *mongo.Collection {
	return m.db.Collection(name)
}
