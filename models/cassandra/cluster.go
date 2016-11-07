package cassandra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
)

func NewCluster(ctx context.Context) *Cluster {
	c := &Cluster{}
	c.AppContext = ctx
	c.table = "clusters"

	return c
}

type ClusterMember struct {
	ID      int64
	Email   string
	Level   string
	Enabled bool
}

type ClusterRow struct {
	ID            int64          `db:"id"`
	Name          string         `db:"name"`
	CreatorID     int64          `db:"creator_id"`
	CreatorEmail  string         `db:"creator_email"`
	DataRetention map[string]int `db:"data_retention"`
	Members       []string       `db:"members"`
}

// GetDeletedFromUNIXTimestampForSelect returns UNIX timestamp from which data should be queried.
func (cr *ClusterRow) GetDeletedFromUNIXTimestampForSelect(tableName string) int64 {
	retention := cr.DataRetention[tableName]
	return time.Now().UTC().Unix() - int64(retention*86400) // 1 day = 86400 seconds
}

// GetDeletedFromUNIXTimestampForInsert returns UNIX timestamp for timeseries data on insert.
func (cr *ClusterRow) GetDeletedFromUNIXTimestampForInsert(tableName string) int64 {
	retention := cr.DataRetention[tableName]
	return time.Now().UTC().Unix() + int64(retention*86400) // 1 day = 86400 seconds
}

// GetTTLDurationForInsert returns TTL duration for timeseries data on insert.
func (cr *ClusterRow) GetTTLDurationForInsert(tableName string) time.Duration {
	retention := cr.DataRetention[tableName]
	return time.Duration(retention * 86400) // 1 day = 86400 seconds
}

// GetMembers returns Members in map
func (cr *ClusterRow) GetMembers() []ClusterMember {
	members := make([]ClusterMember, 0)

	for _, m := range cr.Members {
		member := ClusterMember{}

		err := json.Unmarshal([]byte(m), &member)
		if err == nil {
			members = append(members, member)
		}
	}

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

// GetByID returns one record by id.
func (c *Cluster) GetByID(id int64) (*ClusterRow, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, name, creator_id, creator_email, data_retention, members FROM %v WHERE id=?", c.table)

	var scannedID, scannedCreatorID int64
	var scannedName, scannedCreatorEmail string
	var scannedDataRetention map[string]int
	var scannedMembers []string

	err = session.Query(query, id).Scan(&scannedID, &scannedName, &scannedCreatorID, &scannedCreatorEmail, &scannedDataRetention, &scannedMembers)
	if err != nil {
		return nil, err
	}

	row := &ClusterRow{
		ID:            scannedID,
		Name:          scannedName,
		CreatorID:     scannedCreatorID,
		CreatorEmail:  scannedCreatorEmail,
		DataRetention: scannedDataRetention,
		Members:       scannedMembers,
	}

	return row, err
}

// Create a cluster row record with default settings.
func (c *Cluster) Create(creator *UserRow, name string) (*ClusterRow, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	dataRetention := make(map[string]int)
	dataRetention["ts_checks"] = 1
	dataRetention["ts_metrics"] = 1
	dataRetention["ts_events"] = 1
	dataRetention["ts_executor_logs"] = 1
	dataRetention["ts_logs"] = 1

	member := ClusterMember{}
	member.ID = creator.ID
	member.Email = creator.Email
	member.Level = "write"
	member.Enabled = true

	memberJSON, err := json.Marshal(member)
	if err != nil {
		return nil, err
	}

	id := NewExplicitID()

	query := fmt.Sprintf("INSERT INTO %v (id, name, creator_id, creator_email, data_retention, members) VALUES (?, ?, ?, ?, ?, ?)", c.table)

	err = session.Query(query, id, name, creator.ID, creator.Email, dataRetention, []string{string(memberJSON)}).Exec()
	if err != nil {
		return nil, err
	}

	return c.GetByID(id)
}

// AllByUserID returns all clusters rows by user ID.
func (c *Cluster) AllByUserID(userId int64) ([]*ClusterRow, error) {
	rows, err := c.All()
	if err != nil {
		return nil, err
	}

	newRows := make([]*ClusterRow, 0)

	for _, row := range rows {
		for _, member := range row.GetMembers() {
			if member.ID == userId {
				newRows = append(newRows, row)
				break
			}
		}
	}

	return newRows, err
}

// All returns all clusters rows.
func (c *Cluster) All() ([]*ClusterRow, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*ClusterRow{}

	query := fmt.Sprintf(`SELECT id, name, creator_id, creator_email, data_retention, members FROM %v`, c.table)

	var scannedID, scannedCreatorID int64
	var scannedName, scannedCreatorEmail string
	var scannedDataRetention map[string]int
	var scannedMembers []string

	iter := session.Query(query).Iter()
	for iter.Scan(&scannedID, &scannedName, &scannedCreatorID, &scannedCreatorEmail, &scannedDataRetention, &scannedMembers) {
		rows = append(rows, &ClusterRow{
			ID:            scannedID,
			Name:          scannedName,
			CreatorID:     scannedCreatorID,
			CreatorEmail:  scannedCreatorEmail,
			DataRetention: scannedDataRetention,
			Members:       scannedMembers,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "User.All"}).Error(err)

		return nil, err
	}
	return rows, err
}

// AllSplitToDaemons returns all rows divided into daemons equally.
func (c *Cluster) AllSplitToDaemons(daemons []string) (map[string][]*ClusterRow, error) {
	rows, err := c.All()
	if err != nil {
		return nil, err
	}

	buckets := make([][]*ClusterRow, len(daemons))
	for i, _ := range daemons {
		buckets[i] = make([]*ClusterRow, 0)
	}

	bucketsPointer := 0
	for _, row := range rows {
		if bucketsPointer >= len(buckets) {
			bucketsPointer = 0
		}

		buckets[bucketsPointer] = append(buckets[bucketsPointer], row)
		bucketsPointer = bucketsPointer + 1
	}

	result := make(map[string][]*ClusterRow)

	for i, checksInbucket := range buckets {
		result[daemons[i]] = checksInbucket
	}

	return result, err
}

// UpdateMember adds or updates cluster member information.
func (c *Cluster) UpdateMember(id int64, user *UserRow, level string, enabled bool) error {
	clusterRow, err := c.GetByID(id)
	if err != nil {
		return err
	}

	session, err := c.GetCassandraSession()
	if err != nil {
		return err
	}

	members := clusterRow.GetMembers()

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

	inListJSON := make([]string, 0)

	for _, member := range members {
		inJSON, err := json.Marshal(member)
		if err == nil {
			inListJSON = append(inListJSON, string(inJSON))
		}
	}

	query := fmt.Sprintf(`UPDATE %v SET members = ? WHERE id = ? IF EXISTS`, c.table)

	return session.Query(query, inListJSON, id).Exec()
}

// RemoveMember from a cluster.
func (c *Cluster) RemoveMember(id int64, user *UserRow) error {
	clusterRow, err := c.GetByID(id)
	if err != nil {
		return err
	}

	session, err := c.GetCassandraSession()
	if err != nil {
		return err
	}

	members := clusterRow.GetMembers()

	newMembers := make([]ClusterMember, 0)

	for _, member := range members {
		if member.ID != user.ID {
			newMembers = append(newMembers, member)
		}
	}

	inListJSON := make([]string, 0)

	for _, member := range newMembers {
		inJSON, err := json.Marshal(member)
		if err == nil {
			inListJSON = append(inListJSON, string(inJSON))
		}
	}

	query := fmt.Sprintf(`UPDATE %v SET members = ? WHERE id = ? IF EXISTS`, c.table)

	return session.Query(query, inListJSON, id).Exec()
}

// UpdateNameAndDataRetentionByID.
func (c *Cluster) UpdateNameAndDataRetentionByID(id int64, name string, dataRetention map[string]int) error {
	session, err := c.GetCassandraSession()
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`UPDATE %v SET name = ?, data_retention = ? WHERE id = ? IF EXISTS`, c.table)

	return session.Query(query, name, dataRetention, id).Exec()
}
