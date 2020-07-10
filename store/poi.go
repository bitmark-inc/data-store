package store

import (
	"context"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (m *mongoCommunityStore) AddPOIRating(ctx context.Context, accountNumber, poiID string, ratings map[string]float64) error {
	_, err := m.Resource("poi_ratings").UpdateOne(ctx,
		bson.M{"id": poiID, "account_number": accountNumber},
		bson.M{
			"$set":         bson.M{"ratings": ratings},
			"$setOnInsert": bson.M{"id": poiID, "account_number": accountNumber},
		},
		options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}

type POISummarizedRating struct {
	AverageRating float64            `bson:"rating_avg" json:"rating_avg"`
	Ratings       map[string]float64 `bson:"ratings" json:"ratings"`
}

func (m *mongoCommunityStore) GetPOISummarizedRating(ctx context.Context, poiID string) (POISummarizedRating, error) {
	log.WithField("id", poiID).Info("get poi rating summary")
	cursor, err := m.Resource("poi_ratings").Aggregate(ctx,
		mongo.Pipeline{
			AggregationMatch(bson.M{"id": poiID}),
			AggregationAddFields(bson.M{
				"rating": bson.M{"$objectToArray": "$ratings"},
			}),
			AggregationUnwind("$rating"),
			AggregationGroup(bson.M{
				"k":  "$rating.k",
				"id": "$id",
			}, bson.D{
				bson.E{"v", bson.M{"$avg": "$rating.v"}},
			}),
			AggregationGroup("$_id.id", bson.D{
				bson.E{"rating_avg", bson.M{"$avg": "$v"}},
				bson.E{"ratings", bson.M{"$push": bson.M{
					"k": "$_id.k",
					"v": "$v",
				}}},
			}),
			AggregationAddFields(bson.M{
				"ratings": bson.M{"$arrayToObject": "$ratings"},
			}),
		})
	if err != nil {
		return POISummarizedRating{}, err
	}

	var rating POISummarizedRating
	if cursor.Next(ctx) {
		err := cursor.Decode(&rating)
		if err != nil {
			return POISummarizedRating{}, err
		}
	} else {
		return POISummarizedRating{}, nil
	}

	return rating, nil
}
