package dal

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewStacks(db *sqlx.DB) *Stacks {
	metadata := &Stacks{}
	metadata.db = db
	metadata.table = "metadata"

	return metadata
}

type StacksRow struct {
	ClusterID int64               `db:"cluster_id" json:"-"`
	Data      sqlx_types.JsonText `db:"data"`
}

func (metadataRow *StacksRow) DataString() string {
	return string(metadataRow.Data)
}

type Stacks struct {
	Base
}

// GetByClusterID returns record by cluster_id.
func (metadata *Stacks) GetByClusterID(tx *sqlx.Tx, clusterID int64) (*StacksRow, error) {
	metadataRow := &StacksRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1", metadata.table)
	err := metadata.db.Get(metadataRow, query, clusterID)

	return metadataRow, err
}

// UpdateByClusterID updates record by cluster_id.
func (metadata *Stacks) UpdateByClusterID(tx *sqlx.Tx, clusterID int64, data []byte) (*StacksRow, error) {
	query := fmt.Sprintf("UPDATE %v SET data=$3 WHERE cluster_id=$1", metadata.table)

	_, err := metadata.db.Exec(query, clusterID, data)
	if err != nil {
		return nil, err
	}

	return &StacksRow{clusterID, data}, nil
}

// CreateOrUpdate performs insert/update for one metadata data.
func (metadata *Stacks) CreateOrUpdate(tx *sqlx.Tx, clusterID int64, data []byte) (*StacksRow, error) {
	metadataRow, err := metadata.GetByClusterID(tx, clusterID)

	// Perform INSERT
	if metadataRow == nil || err != nil {
		saveData := make(map[string]interface{})
		saveData["cluster_id"] = clusterID
		saveData["data"] = data

		_, err := metadata.InsertIntoTable(tx, saveData)
		if err != nil {
			return nil, err
		}

		return &StacksRow{clusterID, data}, nil
	}

	// Perform UPDATE
	return metadata.UpdateByClusterID(tx, clusterID, data)
}
