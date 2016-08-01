package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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

type ClusterMember struct {
	ID      int64
	Email   string
	Level   string
	Enabled bool
}

type ClusterRow struct {
	ID            int64               `db:"id"`
	Name          string              `db:"name"`
	CreatorID     int64               `db:"creator_id"`
	CreatorEmail  string              `db:"creator_email"`
	DataRetention sqlx_types.JSONText `db:"data_retention"`
	Members       sqlx_types.JSONText `db:"members"`
}

// GetDataRetention returns DataRetention in map
func (cr *ClusterRow) GetDataRetention() map[string]int {
	retentions := make(map[string]int)
	cr.DataRetention.Unmarshal(&retentions)

	return retentions
}

// GetDeletedFromUNIXTimestampForSelect returns UNIX timestamp from which data should be queried.
func (cr *ClusterRow) GetDeletedFromUNIXTimestampForSelect(tableName string) int64 {
	retention := cr.GetDataRetention()[tableName]
	return time.Now().UTC().Unix() - int64(retention*86400) // 1 day = 86400 seconds
}

// GetDeletedFromUNIXTimestampForInsert returns UNIX timestamp for timeseries data on insert.
func (cr *ClusterRow) GetDeletedFromUNIXTimestampForInsert(tableName string) int64 {
	retention := cr.GetDataRetention()[tableName]
	return time.Now().UTC().Unix() + int64(retention*86400) // 1 day = 86400 seconds
}

// GetMembers returns Members in map
func (cr *ClusterRow) GetMembers() []ClusterMember {
	members := make([]ClusterMember, 0)
	cr.Members.Unmarshal(&members)

	return members
}

// GetLevelByUserID returns a specific cluster member level keyed by user id.
func (cr *ClusterRow) GetLevelByUserID(id int64) string {
	for _, member := range cr.GetMembers() {
		if member.ID == id {
			return member.Level
		}
	}

	return "read"
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

// Create a cluster row record with default settings.
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

	member := ClusterMember{}
	member.ID = creator.ID
	member.Email = creator.Email
	member.Level = "write"

	members := make([]ClusterMember{}, 1)
	members[0] = member

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

// UpdateMember adds or updates cluster member information.
func (c *Cluster) UpdateMember(tx *sqlx.Tx, id int64, user *UserRow, level string, enabled bool) error {
	clusterRow, err := c.GetByID(tx, id)
	if err != nil {
		return err
	}

	members := make([]ClusterMember, 0)
	err = clusterRow.Members.Unmarshal(&members)
	if err != nil {
		return err
	}

	foundExisting := false

	for i, member := range members {
		if member.ID == user.ID {
			memberReplacement := ClusterMember{}
			memberReplacement.ID = member.ID
			memberReplacement.Email = user.Email
			memberReplacement.Level = level
			memberReplacement.Enabled = enabled

			members[i] = memberReplacement
			foundExisting = true
			break
		}
	}

	if !foundExisting {
		member := ClusterMember{}
		member.ID = int64(user.ID)
		member.Email = user.Email
		member.Level = level
		member.Enabled = enabled

		members = append(members, member)
	}

	membersJSON, err := json.Marshal(members)
	if err != nil {
		return err
	}

	println(string(membersJSON))

	data := make(map[string]interface{})
	data["members"] = membersJSON

	_, err = c.UpdateByID(tx, data, id)

	return err
}

// RemoveMember from a cluster.
func (c *Cluster) RemoveMember(tx *sqlx.Tx, id int64, user *UserRow) error {
	clusterRow, err := c.GetByID(tx, id)
	if err != nil {
		return err
	}

	members := make([]ClusterMember, 0)
	err = clusterRow.Members.Unmarshal(&members)
	if err != nil {
		return err
	}

	newMembers := make([]ClusterMember, 0)

	for _, member := range members {
		if member.ID == user.ID {
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
