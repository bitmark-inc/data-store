package store

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SymptomDailyReport struct {
	Date                     string         `bson:"date"`
	Symptoms                 []SymptomStats `bson:"symptoms"`
	CheckinsNumPastThreeDays int            `bson:"checkins_num_past_three_days"`
}

type SymptomStats struct {
	Name  string `bson:"name"`
	Count int    `bson:"count"`
}

func (m *mongoCommunityStore) AddSymptomDailyReports(ctx context.Context, reports []SymptomDailyReport) error {
	opts := options.Update().SetUpsert(true)
	for _, report := range reports {
		filter := bson.M{"date": report.Date}
		update := bson.M{
			"$set": bson.M{
				"symptoms":                     report.Symptoms,
				"checkins_num_past_three_days": report.CheckinsNumPastThreeDays,
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

// GetSymptomReportItems returns report items in the date range which starts at `limit` days ago, and ends at `end`.
func (m *mongoCommunityStore) GetSymptomReportItems(ctx context.Context, end string, limit int64) (map[string][]Bucket, error) {
	pipeline := mongo.Pipeline{
		AggregationMatch(bson.M{
			"date": bson.M{
				"$lte": end,
			},
		}),
		AggregationSort("date", -1),
		AggregationLimit(limit),
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

func (m *mongoCommunityStore) FindLatestDailyReport(ctx context.Context) (*SymptomDailyReport, error) {
	var report SymptomDailyReport

	opts := options.FindOne().SetSort(bson.D{{"date", -1}})
	err := m.Resource("symptom_reports").FindOne(ctx, bson.M{}, opts).Decode(&report)
	return &report, err
}
