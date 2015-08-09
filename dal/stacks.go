package dal

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewStacks(db *sqlx.DB) *Stacks {
	stacks := &Stacks{}
	stacks.db = db
	stacks.table = "stacks"

	return stacks
}

type StacksRow struct {
	ClusterID int64               `db:"cluster_id" json:"-"`
	Data      sqlx_types.JsonText `db:"data"`
}

func (stacksRow *StacksRow) DataString() string {
	return string(stacksRow.Data)
}

type Stacks struct {
	Base
}

// GetByClusterID returns record by cluster_id.
func (stacks *Stacks) GetByClusterID(tx *sqlx.Tx, clusterID int64) (*StacksRow, error) {
	stacksRow := &StacksRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1", stacks.table)
	err := stacks.db.Get(stacksRow, query, clusterID)

	return stacksRow, err
}

// UpdateByClusterID updates record by cluster_id.
func (stacks *Stacks) UpdateByClusterID(tx *sqlx.Tx, clusterID int64, data []byte) (*StacksRow, error) {
	query := fmt.Sprintf("UPDATE %v SET data=$3 WHERE cluster_id=$1", stacks.table)

	_, err := stacks.db.Exec(query, clusterID, data)
	if err != nil {
		return nil, err
	}

	return &StacksRow{clusterID, data}, nil
}

// CreateOrUpdate performs insert/update for one stacks data.
func (stacks *Stacks) CreateOrUpdate(tx *sqlx.Tx, clusterID int64, data []byte) (*StacksRow, error) {
	stacksRow, err := stacks.GetByClusterID(tx, clusterID)

	// Perform INSERT
	if stacksRow == nil || err != nil {
		saveData := make(map[string]interface{})
		saveData["cluster_id"] = clusterID
		saveData["data"] = data

		_, err := stacks.InsertIntoTable(tx, saveData)
		if err != nil {
			return nil, err
		}

		return &StacksRow{clusterID, data}, nil
	}

	// Perform UPDATE
	return stacks.UpdateByClusterID(tx, clusterID, data)
}
