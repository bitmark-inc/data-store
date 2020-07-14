package store

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SymptomDailyReport struct {
	Date     string `json:"date"`
	Symptoms []struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"symptoms"`
}

func (m *mongoCommunityStore) AddSymptomDailyReports(ctx context.Context, reports []*SymptomDailyReport) error {
	opts := options.Update().SetUpsert(true)
	for _, report := range reports {
		filter := bson.M{"date": report.Date}
		update := bson.M{
			"$set": bson.M{
				"symptoms": report.Symptoms,
			},
		}
		_, err := m.Resource("symptom_reports").UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

type BucketAggregation struct {
	ID      string   `bson:"_id"`
	Buckets []Bucket `bson:"buckets"`
}

type Bucket struct {
	Name  string `bson:"name" json:"name"`
	Value int    `bson:"value" json:"value"`
}

func (m *mongoCommunityStore) GetSymptomReportItems(ctx context.Context, start, end string) (map[string][]Bucket, error) {
	pipeline := mongo.Pipeline{
		AggregationMatch(bson.M{
			"date": bson.M{
				"$gte": start,
				"$lt":  end,
			},
		}),
		AggregationUnwind("$symptoms"),
		AggregationGroup("$symptoms.name", bson.D{
			bson.E{
				Key: "buckets",
				Value: bson.M{"$push": bson.M{
					"name":  "$date",
					"value": "$symptoms.count",
				}},
			},
		}),
	}
	cursor, err := m.Resource("symptom_reports").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	results := make(map[string][]Bucket)
	for cursor.Next(ctx) {
		var aggItem BucketAggregation
		if err := cursor.Decode(&aggItem); err != nil {
			return nil, err
		}
		results[aggItem.ID] = aggItem.Buckets
	}
	return results, nil
}
