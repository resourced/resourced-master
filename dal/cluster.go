package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewCluster(db *sqlx.DB) *Cluster {
	c := &Cluster{}
	c.db = db
	c.table = "clusters"
	c.hasID = true

	return c
}

type ClusterRow struct {
	ID            int64               `db:"id"`
	Name          string              `db:"name"`
	CreatorID     int64               `db:"creator_id"`
	CreatorEmail  string              `db:"creator_email"`
	DataRetention sqlx_types.JSONText `db:"data_retention"`
	Members       sqlx_types.JSONText `db:"members"`
}

func (cr *ClusterRow) GetDataRetention() map[string]int {
	retentions := make(map[string]int)
	cr.DataRetention.Unmarshal(&retentions)

	return retentions
}

type Cluster struct {
	Base
}

func (c *Cluster) clusterRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*ClusterRow, error) {
	id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return c.GetByID(tx, id)
}

// GetByID returns one record by id.
func (c *Cluster) GetByID(tx *sqlx.Tx, id int64) (*ClusterRow, error) {
	row := &ClusterRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", c.table)
	err := c.db.Get(row, query, id)

	return row, err
}

func (c *Cluster) Create(tx *sqlx.Tx, creator *UserRow, name string) (*ClusterRow, error) {
	dataRetention := make(map[string]int)
	dataRetention["ts_checks"] = 1
	dataRetention["ts_metrics"] = 1
	dataRetention["ts_metrics_aggr_15m"] = 1
	dataRetention["ts_events"] = 1
	dataRetention["ts_executor_logs"] = 1
	dataRetention["ts_logs"] = 1

	dataRetentionJSON, err := json.Marshal(dataRetention)
	if err != nil {
		return nil, err
	}

	members := make([]map[string]interface{}, 1)
	members[0] = make(map[string]interface{})
	members[0]["ID"] = creator.ID
	members[0]["Email"] = creator.Email
	members[0]["Permission"] = "execute"

	membersJSON, err := json.Marshal(members)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["name"] = name
	data["creator_id"] = creator.ID
	data["creator_email"] = creator.Email
	data["data_retention"] = dataRetentionJSON
	data["members"] = membersJSON

	sqlResult, err := c.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return c.clusterRowFromSqlResult(tx, sqlResult)
}

// AllByUserID returns all clusters rows by user ID.
func (c *Cluster) AllByUserID(tx *sqlx.Tx, userId int64) ([]*ClusterRow, error) {
	rows := []*ClusterRow{}

	query := fmt.Sprintf(`SELECT * from %v WHERE members @> '[{"ID" : %v}]'`, c.table, userId)
	err := c.db.Select(&rows, query)
	if err != nil {
		logrus.WithFields(logrus.Fields{"Query": query}).Error(err)
	}

	return rows, err
}

// All returns all clusters rows.
func (c *Cluster) All(tx *sqlx.Tx) ([]*ClusterRow, error) {
	rows := []*ClusterRow{}
	query := fmt.Sprintf("SELECT * FROM %v", c.table)
	err := c.db.Select(&rows, query)

	return rows, err
}

// AllSplitToDaemons returns all rows divided into daemons equally.
func (c *Cluster) AllSplitToDaemons(tx *sqlx.Tx, daemons []string) (map[string][]*ClusterRow, error) {
	rows, err := c.All(tx)
	if err != nil {
		return nil, err
	}

	buckets := make([][]*ClusterRow, len(daemons))
	for i, _ := range daemons {
		buckets[i] = make([]*ClusterRow, 0)
	}

	bucketsPointer := 0
	for _, row := range rows {
		buckets[bucketsPointer] = append(buckets[bucketsPointer], row)
		bucketsPointer = bucketsPointer + 1

		if bucketsPointer >= len(buckets) {
			bucketsPointer = 0
		}
	}

	result := make(map[string][]*ClusterRow)

	for i, checksInbucket := range buckets {
		result[daemons[i]] = checksInbucket
	}

	return result, err
}

func (c *Cluster) UpdateMember(tx *sqlx.Tx, id int64, user *UserRow, permission string) error {
	clusterRow, err := c.GetByID(tx, id)
	if err != nil {
		return err
	}

	members := make([]map[string]interface{}, 0)
	err = clusterRow.Members.Unmarshal(&members)
	if err != nil {
		return err
	}

	foundExisting := false

	for _, member := range members {
		if member["ID"].(int64) == user.ID {
			member["Email"] = user.Email
			member["Permission"] = permission

			foundExisting = true
			break
		}
	}

	if !foundExisting {
		member := make(map[string]interface{})
		member["ID"] = user.ID
		member["Email"] = user.Email
		member["Permission"] = permission

		members = append(members, member)
	}

	membersJSON, err := json.Marshal(members)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["members"] = membersJSON

	_, err = c.UpdateByID(tx, data, id)

	return err
}

func (c *Cluster) RemoveMember(tx *sqlx.Tx, id int64, user *UserRow) error {
	clusterRow, err := c.GetByID(tx, id)
	if err != nil {
		return err
	}

	members := make([]map[string]interface{}, 0)
	err = clusterRow.Members.Unmarshal(&members)
	if err != nil {
		return err
	}

	newMembers := make([]map[string]interface{}, 0)

	for _, member := range members {
		if member["ID"].(int64) != user.ID {
			newMembers = append(newMembers, member)
		}
	}

	newMembersJSON, err := json.Marshal(members)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["members"] = newMembersJSON

	_, err = c.UpdateByID(tx, data, id)

	return err
}
