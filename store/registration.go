package store

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func indexForPersonalAccountStore(db *mongo.Database) error {
	_, err := db.Collection("poi_ratings").Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{"id", 1},
			},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
	})
	return err
}

func indexForCommunityStore(db *mongo.Database) error {
	_, err := db.Collection("poi_ratings").Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{"account_number", 1},
				{"id", 1},
			},
			Options: options.Index().SetUnique(true).SetName("id_account_unique"),
		},
	})
	return err
}

// Account returns a personal data store.
func (m mongodbDataPool) RegisterAccount(accountNumber string) error {
	dbName := fmt.Sprintf("%s%s", m.dbPrefix, accountNumber)
	return indexForPersonalAccountStore(m.client.Database(dbName))
}

func (m mongodbDataPool) InitCommunityStore() error {
	dbName := fmt.Sprintf("%scommunity", m.dbPrefix)
	return indexForCommunityStore(m.client.Database(dbName))
}
