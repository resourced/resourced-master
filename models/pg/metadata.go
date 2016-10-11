package pg

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewMetadata(ctx context.Context) *Metadata {
	metadata := &Metadata{}
	metadata.AppContext = ctx
	metadata.table = "metadata"

	return metadata
}

type MetadataRow struct {
	ClusterID int64               `db:"cluster_id" json:"-"`
	Key       string              `db:"key"`
	Data      sqlx_types.JSONText `db:"data"`
}

func (metadataRow *MetadataRow) DataString() string {
	return string(metadataRow.Data)
}

type Metadata struct {
	Base
}

// AllByClusterID returns all metadata rows.
func (metadata *Metadata) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*MetadataRow, error) {
	pgdb, err := metadata.GetPGDB()
	if err != nil {
		return nil, err
	}

	metadataRows := []*MetadataRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 ORDER BY key ASC", metadata.table)
	err = pgdb.Select(&metadataRows, query, clusterID)

	return metadataRows, err
}

// GetByClusterIDAndKey returns record by cluster_id and key.
func (metadata *Metadata) GetByClusterIDAndKey(tx *sqlx.Tx, clusterID int64, key string) (*MetadataRow, error) {
	pgdb, err := metadata.GetPGDB()
	if err != nil {
		return nil, err
	}

	metadataRow := &MetadataRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND key=$2 ORDER BY key ASC", metadata.table)
	err = pgdb.Get(metadataRow, query, clusterID, key)

	return metadataRow, err
}

// UpdateByClusterIDAndKey updates record by cluster_id and key.
func (metadata *Metadata) UpdateByClusterIDAndKey(tx *sqlx.Tx, clusterID int64, key string, data []byte) (*MetadataRow, error) {
	pgdb, err := metadata.GetPGDB()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("UPDATE %v SET data=$3 WHERE cluster_id=$1 AND key=$2", metadata.table)

	_, err = pgdb.Exec(query, clusterID, key, data)
	if err != nil {
		return nil, err
	}

	return &MetadataRow{clusterID, key, data}, nil
}

// DeleteByClusterIDAndKey updates record by cluster_id and key.
func (metadata *Metadata) DeleteByClusterIDAndKey(tx *sqlx.Tx, clusterID int64, key string) (*MetadataRow, error) {
	pgdb, err := metadata.GetPGDB()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("DELETE FROM %v WHERE cluster_id=$1 AND key=$2", metadata.table)

	_, err = pgdb.Exec(query, clusterID, key)
	if err != nil {
		return nil, err
	}

	return &MetadataRow{clusterID, key, nil}, nil
}

// CreateOrUpdate performs insert/update for one metadata data.
func (metadata *Metadata) CreateOrUpdate(tx *sqlx.Tx, clusterID int64, key string, data []byte) (*MetadataRow, error) {
	metadataRow, err := metadata.GetByClusterIDAndKey(tx, clusterID, key)

	// Perform INSERT
	if metadataRow == nil || err != nil {
		saveData := make(map[string]interface{})
		saveData["cluster_id"] = clusterID
		saveData["key"] = key
		saveData["data"] = data

		_, err := metadata.InsertIntoTable(tx, saveData)
		if err != nil {
			return nil, err
		}

		return &MetadataRow{clusterID, key, data}, nil
	}

	// Perform UPDATE
	return metadata.UpdateByClusterIDAndKey(tx, clusterID, key, data)
}
