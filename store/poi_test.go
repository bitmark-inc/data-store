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

const (
	TestDBPrefix = "testcase_"
)

var (
	defaultRatingAccount = "account_default"
	testGetPOIRatingID   = "test1234"
	defaultRating        = map[string]interface{}{
		"id":      testGetPOIRatingID,
		"ratings": map[string]float64{"a": 1},
	}

	testGetCommunityRatingID1 = "testCommunity1234"
	testGetCommunityRatingID2 = "testCommunity5678"
	defaultCommunityRatings   = bson.A{
		map[string]interface{}{
			"id":             testGetCommunityRatingID1,
			"account_number": "user1",
			"ratings":        map[string]float64{"a": 3, "b": 4},
		},
		map[string]interface{}{
			"id":             testGetCommunityRatingID1,
			"account_number": "user2",
			"ratings":        map[string]float64{"a": 1, "b": 2},
		},
		map[string]interface{}{
			"id":             testGetCommunityRatingID2,
			"account_number": "user1",
			"ratings":        map[string]float64{"a": 1, "b": 2, "c": 3},
		},
		map[string]interface{}{
			"id":             testGetCommunityRatingID2,
			"account_number": "user2",
			"ratings":        map[string]float64{"a": 5, "b": 4, "c": 3},
		},
	}
)

type AccountPOITestSuite struct {
	suite.Suite
	connURI     string
	mongoClient *mongo.Client
}

func NewAccountPOITestSuite(connURI string) *AccountPOITestSuite {
	return &AccountPOITestSuite{
		connURI: connURI,
	}
}

func (s *AccountPOITestSuite) SetupSuite() {
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

func (s *AccountPOITestSuite) SetupTest() {}

func (s *AccountPOITestSuite) LoadFixtures() error {
	ctx := context.Background()
	if _, err := s.mongoClient.Database(TestDBPrefix+defaultRatingAccount).Collection("poi_ratings").InsertOne(ctx, defaultRating); err != nil {
		return err
	}
	if _, err := s.mongoClient.Database("testcase_community").Collection("poi_ratings").InsertMany(ctx, defaultCommunityRatings); err != nil {
		return err
	}
	return nil
}

func (s *AccountPOITestSuite) TestAccountSetPOIRating() {
	ctx := context.Background()
	testAccount := "testcase_account1"
	testPOIID := "abcd"
	err := NewMongodbDataPool(s.mongoClient, TestDBPrefix).Account(testAccount).SetPOIRating(ctx, testPOIID, map[string]float64{
		"a": 1,
		"b": 2,
		"c": 3,
		"d": 4,
	})
	s.NoError(err)

	cursor, err := s.mongoClient.Database(TestDBPrefix+testAccount).Collection("poi_ratings").Find(ctx, bson.M{"id": testPOIID})
	s.NoError(err)

	var results []struct {
		Ratings map[string]float64 `bson:"ratings"`
	}
	err = cursor.All(ctx, &results)
	s.NoError(err)
	s.Len(results, 1)
	s.Equal(results[0].Ratings["a"], 1.0)
	s.Equal(results[0].Ratings["b"], 2.0)
	s.Equal(results[0].Ratings["c"], 3.0)
	s.Equal(results[0].Ratings["d"], 4.0)
}

func (s *AccountPOITestSuite) TestAccountGetPOIRating() {
	ctx := context.Background()
	ratings, err := NewMongodbDataPool(s.mongoClient, TestDBPrefix).Account(defaultRatingAccount).GetPOIRating(ctx, testGetPOIRatingID)
	s.NoError(err)
	s.Equal(ratings["a"], 1.0)
}

func (s *AccountPOITestSuite) TestCommunitySetPOIRating() {
	ctx := context.Background()
	testAccount := "testcase_account1"
	testPOIID := "abcd"
	err := NewMongodbDataPool(s.mongoClient, TestDBPrefix).Community().SetPOIRating(ctx, testAccount, testPOIID, map[string]float64{
		"a": 5,
		"b": 4,
		"c": 3,
		"d": 2,
	})
	s.NoError(err)

	cursor, err := s.mongoClient.Database("testcase_community").Collection("poi_ratings").Find(ctx, bson.M{"id": testPOIID})
	s.NoError(err)

	var results []struct {
		Ratings map[string]float64 `bson:"ratings"`
	}
	err = cursor.All(ctx, &results)
	s.NoError(err)
	s.Len(results, 1)
	s.Equal(results[0].Ratings["a"], 5.0)
	s.Equal(results[0].Ratings["b"], 4.0)
	s.Equal(results[0].Ratings["c"], 3.0)
	s.Equal(results[0].Ratings["d"], 2.0)
}

func (s *AccountPOITestSuite) TestCommunityGetPOIRating() {
	ctx := context.Background()
	ratings, err := NewMongodbDataPool(s.mongoClient, TestDBPrefix).Community().GetPOISummarizedRatings(ctx, []string{testGetCommunityRatingID1, testGetCommunityRatingID2})
	s.NoError(err)
	s.Len(ratings, 2)
	s.Equal(2.5, ratings[testGetCommunityRatingID1].AverageRating)
	s.Equal(2.0, ratings[testGetCommunityRatingID1].Ratings["a"].Score)
	s.Equal(3.0, ratings[testGetCommunityRatingID1].Ratings["b"].Score)
	s.Equal(2, ratings[testGetCommunityRatingID1].Ratings["a"].Counts)
	s.Equal(2, ratings[testGetCommunityRatingID1].Ratings["b"].Counts)
	s.Equal(int64(2), ratings[testGetCommunityRatingID1].RatingCount)

	s.Equal(3.0, ratings[testGetCommunityRatingID2].AverageRating)
	s.Equal(3.0, ratings[testGetCommunityRatingID2].Ratings["a"].Score)
	s.Equal(3.0, ratings[testGetCommunityRatingID2].Ratings["b"].Score)
	s.Equal(3.0, ratings[testGetCommunityRatingID2].Ratings["c"].Score)
	s.Equal(2, ratings[testGetCommunityRatingID2].Ratings["a"].Counts)
	s.Equal(2, ratings[testGetCommunityRatingID2].Ratings["b"].Counts)
	s.Equal(2, ratings[testGetCommunityRatingID2].Ratings["b"].Counts)
	s.Equal(int64(2), ratings[testGetCommunityRatingID2].RatingCount)
}

func TestAccountPOI(t *testing.T) {
	suite.Run(t, NewAccountPOITestSuite("mongodb://127.0.0.1:27017/?compressors=disabled"))
}
