package store

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type POIResourceRating struct {
	Ratings map[string]float64 `bson:"ratings"`
}

func (m *mongoAccountStore) RatePOIResource(ctx context.Context, poiID string, ratings map[string]float64) error {
	_, err := m.Resource("poi_ratings").UpdateOne(ctx,
		bson.M{"id": poiID},
		bson.M{
			"$set":         bson.M{"ratings": ratings},
			"$setOnInsert": bson.M{"id": poiID},
		},
		options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}

func (m *mongoAccountStore) GetPOIResource(ctx context.Context, poiID string) (map[string]float64, error) {
	var rating POIResourceRating

	if err := m.Resource("poi_ratings").FindOne(ctx, bson.M{"id": poiID}).Decode(&rating); err != nil {
		return nil, err
	}

	return rating.Ratings, nil
}
