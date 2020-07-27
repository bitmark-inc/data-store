package store

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
)

var ResourceToExport = []string{"poi_ratings"}

// ExportData prepares an archive file that contains all resources to be exported from personal data store
func (m *mongoAccountStore) ExportData(ctx context.Context) ([]byte, error) {
	// prepare temporary archive file
	fs := afero.NewOsFs()

	zipFile, err := afero.TempFile(fs, viper.GetString("archive.tempdir"), fmt.Sprintf("pds-data-export-%s-zip-", m.accountNumber))
	if err != nil {
		return nil, err
	}

	defer fs.Remove(zipFile.Name())
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)

	for _, resource := range ResourceToExport {
		cursor, err := m.Resource(resource).Find(ctx, bson.M{})
		if err != nil {
			return nil, err
		}

		var ratingData []interface{}
		if err := cursor.All(ctx, &ratingData); err != nil {
			return nil, err
		}

		w, err := archive.Create(fmt.Sprintf("pds/%s.json", resource))
		if err != nil {
			return nil, err
		}

		if err := json.NewEncoder(w).Encode(ratingData); err != nil {
			return nil, err
		}
	}

	if err := archive.Close(); err != nil {
		return nil, err
	}

	if _, err := zipFile.Seek(0, 0); err != nil {
		return nil, err
	}

	return ioutil.ReadAll(zipFile)
}

func (m *mongoAccountStore) DeleteData(ctx context.Context) error {
	collectionNames, err := m.db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	for _, name := range collectionNames {
		if err := m.db.Collection(name).Drop(ctx); err != nil {
			return err
		}
	}
	return nil
}

// ExportData prepares an archive file that contains all resources to be exported from community data store
func (m *mongoCommunityStore) ExportData(ctx context.Context, accountNumber string) ([]byte, error) {
	if accountNumber == "" {
		return nil, fmt.Errorf("empty account number error")
	}

	// prepare temporary archive file
	fs := afero.NewOsFs()

	zipFile, err := afero.TempFile(fs, viper.GetString("archive.tempdir"), fmt.Sprintf("cds-data-export-%s-zip-", accountNumber))
	if err != nil {
		return nil, err
	}

	defer fs.Remove(zipFile.Name())
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)

	for _, resource := range ResourceToExport {
		cursor, err := m.Resource(resource).Find(ctx, bson.M{"account_number": accountNumber})
		if err != nil {
			return nil, err
		}

		var ratingData []interface{}
		if err := cursor.All(ctx, &ratingData); err != nil {
			return nil, err
		}

		w, err := archive.Create(fmt.Sprintf("cds/%s.json", resource))
		if err != nil {
			return nil, err
		}

		if err := json.NewEncoder(w).Encode(ratingData); err != nil {
			return nil, err
		}
	}

	if err := archive.Close(); err != nil {
		return nil, err
	}

	if _, err := zipFile.Seek(0, 0); err != nil {
		return nil, err
	}

	return ioutil.ReadAll(zipFile)
}
