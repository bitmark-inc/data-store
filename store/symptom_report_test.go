package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	reports = []interface{}{
		SymptomDailyReport{
			Date: "2020-07-19",
			Symptoms: []SymptomStats{
				{Name: "Cough", Count: 10},
				{Name: "Fatigue", Count: 18},
			},
			CheckinsNumPastThreeDays: 0,
		},
		SymptomDailyReport{
			Date: "2020-07-20",
			Symptoms: []SymptomStats{
				{Name: "Cough", Count: 7},
				{Name: "Fatigue", Count: 28},
			},
			CheckinsNumPastThreeDays: 0,
		},
		SymptomDailyReport{
			Date: "2020-07-21",
			Symptoms: []SymptomStats{
				{Name: "Cough", Count: 8},
				{Name: "Fatigue", Count: 29},
			},
			CheckinsNumPastThreeDays: 1003,
		}}
)

type SymptomReportTestSuite struct {
	suite.Suite
	connURI     string
	mongoClient *mongo.Client
}

func NewSymptomReportTestSuite(connURI string) *SymptomReportTestSuite {
	return &SymptomReportTestSuite{
		connURI: connURI,
	}
}

func (s *SymptomReportTestSuite) SetupSuite() {
	if s.connURI == "" {
		s.T().Fatal("invalid test suite configuration")
	}

	opts := options.Client().ApplyURI(s.connURI)
	mongoClient, err := mongo.NewClient(opts)
	if nil != err {
		s.T().Fatalf("create mongo client with error: %s", err)
	}

	ctx := context.Background()
	if err = mongoClient.Connect(ctx); nil != err {
		s.T().Fatalf("connect mongo database with error: %s", err.Error())
	}

	s.mongoClient = mongoClient

	dbNames, err := mongoClient.ListDatabaseNames(ctx, bson.M{"name": primitive.Regex{Pattern: "^testcase_"}})
	if err != nil {
		s.T().Fatalf("list all databases with error: %s", err.Error())
	}

	for _, name := range dbNames {
		if err := mongoClient.Database(name).Drop(ctx); err != nil {
			s.T().Fatalf("drop database with error: %s", err.Error())
		}
	}

	if err := s.LoadFixtures(); err != nil {
		s.T().Fatalf("load fixtures with error: %s", err.Error())
	}
}

func (s *SymptomReportTestSuite) SetupTest() {}

func (s *SymptomReportTestSuite) LoadFixtures() error {
	ctx := context.Background()
	if _, err := s.mongoClient.Database("testcase_community").Collection("symptom_reports").InsertMany(ctx, reports); err != nil {
		return err
	}
	return nil
}

func (s *SymptomReportTestSuite) TestCommunityGetSymptomReportItems() {
	dataPool := NewMongodbDataPool(s.mongoClient, TestDBPrefix)

	ctx := context.Background()
	items, err := dataPool.Community().GetSymptomReportItems(ctx, "2020-07-21", 7)
	s.NoError(err)
	s.Equal(map[string][]Bucket{
		"Cough": {
			{"2020-07-21", 8},
			{"2020-07-20", 7},
			{"2020-07-19", 10},
		},
		"Fatigue": {
			{"2020-07-21", 29},
			{"2020-07-20", 28},
			{"2020-07-19", 18},
		},
	}, items)

	items, err = dataPool.Community().GetSymptomReportItems(ctx, "2020-07-18", 7)
	s.NoError(err)
	s.Equal(0, len(items))
}

func (s *SymptomReportTestSuite) TestCommunityFindLatestDailyReport() {
	ctx := context.Background()
	report, err := NewMongodbDataPool(s.mongoClient, TestDBPrefix).Community().FindLatestDailyReport(ctx)
	s.NoError(err)
	s.Equal(&SymptomDailyReport{
		Date: "2020-07-21",
		Symptoms: []SymptomStats{
			{Name: "Cough", Count: 8},
			{Name: "Fatigue", Count: 29},
		},
		CheckinsNumPastThreeDays: 1003,
	}, report)
}

func TestSymptomReport(t *testing.T) {
	suite.Run(t, NewSymptomReportTestSuite("mongodb://127.0.0.1:27017/?compressors=disabled"))
}
