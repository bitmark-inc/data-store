package store

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	testDataExportAccount = "data-export-account"
	testDataDeleteAccount = "data-delete-account"
	testDataExport1ID     = "data-export-1"
	testDataExport2ID     = "data-export-2"
	testDataExport3ID     = "data-export-3"
)

var (
	testDataExportRating1 = map[string]interface{}{
		"id":      testDataExport1ID,
		"ratings": map[string]float64{"a": 1},
	}
	testDataExportRating2 = map[string]interface{}{
		"id":      testDataExport2ID,
		"ratings": map[string]float64{"b": 2},
	}

	testDataExportCommunityRating1 = map[string]interface{}{
		"id":             testDataExport2ID,
		"account_number": testDataExportAccount,
		"ratings":        map[string]float64{"a": 2},
	}
	testDataExportCommunityRating2 = map[string]interface{}{
		"id":             testDataExport2ID,
		"account_number": testDataExportAccount,
		"ratings":        map[string]float64{"b": 3},
	}
	testDataExportCommunityRating3 = map[string]interface{}{
		"id":             testDataExport3ID,
		"account_number": testDataExportAccount,
		"ratings":        map[string]float64{"b": 4},
	}
)

type DataManagementTestSuite struct {
	suite.Suite
	connURI     string
	mongoClient *mongo.Client
}

func NewDataManagementTestSuite(connURI string) *DataManagementTestSuite {
	return &DataManagementTestSuite{
		connURI: connURI,
	}
}

func (s *DataManagementTestSuite) SetupSuite() {
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

	dbNames, err := mongoClient.ListDatabaseNames(ctx, bson.M{"name": primitive.Regex{Pattern: fmt.Sprintf("^%s", TestDBPrefix)}})
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

func (s *DataManagementTestSuite) SetupTest() {}

func (s *DataManagementTestSuite) LoadFixtures() error {
	ctx := context.Background()
	if _, err := s.mongoClient.Database(TestDBPrefix+testDataExportAccount).Collection("poi_ratings").InsertMany(ctx, bson.A{
		testDataExportRating1,
		testDataExportRating2,
	}); err != nil {
		return err
	}

	if _, err := s.mongoClient.Database(TestDBPrefix+"community").Collection("poi_ratings").InsertMany(ctx, bson.A{
		testDataExportCommunityRating1,
		testDataExportCommunityRating2,
		testDataExportCommunityRating3,
	}); err != nil {
		return err
	}

	if _, err := s.mongoClient.Database(TestDBPrefix+testDataDeleteAccount).Collection("poi_ratings").InsertMany(ctx, bson.A{
		testDataExportRating1,
		testDataExportRating2,
	}); err != nil {
		return err
	}

	return nil
}

// TestPDSExport checks archive file extraction and files for a PDS exporting file
func (s *DataManagementTestSuite) TestPDSExport() {
	ctx := context.Background()
	data, err := NewMongodbDataPool(s.mongoClient, TestDBPrefix).Account(testDataExportAccount).ExportData(ctx)
	s.NoError(err)
	s.NotZero(len(data))

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return
	}

	fileNames := map[string]struct{}{}
	for _, file := range reader.File {
		fileNames[file.Name] = struct{}{}

		r, err := file.Open()
		s.NoError(err)
		rating := []interface{}{}

		s.NoError(json.NewDecoder(r).Decode(&rating))
		s.Len(rating, 2)
	}

	s.Contains(fileNames, "pds/poi_ratings.json")
}

// TestCDSExport checks archive file extraction and files for a CDS exporting file
func (s *DataManagementTestSuite) TestCDSExport() {
	ctx := context.Background()
	data, err := NewMongodbDataPool(s.mongoClient, TestDBPrefix).Community().ExportData(ctx, testDataExportAccount)
	s.NoError(err)
	s.NotZero(len(data))

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return
	}

	fileNames := map[string]struct{}{}
	for _, file := range reader.File {
		fileNames[file.Name] = struct{}{}

		r, err := file.Open()
		s.NoError(err)
		rating := []interface{}{}

		s.NoError(json.NewDecoder(r).Decode(&rating))
		s.Len(rating, 3)
	}

	s.Contains(fileNames, "cds/poi_ratings.json")
}

// TestPDSDataDelete validates whether an account database disappears after it is removed
func (s *DataManagementTestSuite) TestPDSDataDelete() {
	ctx := context.Background()
	err := NewMongodbDataPool(s.mongoClient, TestDBPrefix).Account(testDataDeleteAccount).DeleteData(ctx)
	s.NoError(err)

	dbNames, err := s.mongoClient.ListDatabaseNames(ctx, bson.M{"name": primitive.Regex{Pattern: fmt.Sprintf("^%s", TestDBPrefix)}})
	s.NoError(err)

	for _, n := range dbNames {
		if n == TestDBPrefix+testDataDeleteAccount {
			s.T().Fatalf("database should not be existed")
		}
	}
}

func TestDataManagement(t *testing.T) {
	suite.Run(t, NewDataManagementTestSuite("mongodb://127.0.0.1:27017/?compressors=disabled"))
}
