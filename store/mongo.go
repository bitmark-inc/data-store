package store

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	mongoLogPrefix = "mongo"
)

type DataStorePool interface {
	RegisterAccount(accountNumber string) error

	// Account will return a personal data store for a specific account number.
	Account(accountNumber string) PersonalDataStore

	// Community will return a community data store.
	Community() CommunityDataStore
}

type PersonalDataStore interface {
	SetPOIRating(ctx context.Context, poiID string, ratings map[string]float64) error
	GetPOIRating(ctx context.Context, poiID string) (map[string]float64, error)
	ExportData(ctx context.Context) ([]byte, error)
	DeleteData(ctx context.Context) error
}

type CommunityDataStore interface {
	SetPOIRating(ctx context.Context, accountNumber, poiID string, ratings map[string]float64) error
	GetPOISummarizedRatings(ctx context.Context, poiIDs []string) (map[string]POISummarizedRating, error)
	AddSymptomDailyReports(ctx context.Context, reports []SymptomDailyReport) error
	GetSymptomReportItems(ctx context.Context, start, end string) (map[string][]Bucket, error)
	ExportData(ctx context.Context, accountNumber string) ([]byte, error)
}

// mongodbDataPool is an implementation of DataStorePool.
type mongodbDataPool struct {
	client   *mongo.Client
	dbPrefix string
}

// NewMongodbDataPool returns a mongodbDataPool instance
func NewMongodbDataPool(client *mongo.Client, dbPrefix string) *mongodbDataPool {
	return &mongodbDataPool{
		client:   client,
		dbPrefix: dbPrefix,
	}
}

// Account returns a personal data store.
func (m mongodbDataPool) Account(accountNumber string) PersonalDataStore {
	if accountNumber == "" {
		return nil
	}

	dbName := fmt.Sprintf("%s%s", m.dbPrefix, accountNumber)
	db := m.client.Database(dbName)
	indexForPersonalAccountStore(db)
	return &mongoAccountStore{
		accountNumber: accountNumber,
		db:            db,
	}
}

// Community returns a community data store.
func (m mongodbDataPool) Community() CommunityDataStore {
	dbName := fmt.Sprintf("%scommunity", m.dbPrefix)
	return &mongoCommunityStore{
		db: m.client.Database(dbName),
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

type mongoCommunityStore struct {
	db *mongo.Database
}

// Resource returns the collection of the given resource from the database
// of the community store
func (m mongoCommunityStore) Resource(name string) *mongo.Collection {
	return m.db.Collection(name)
}
