package cassandra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/models/shared"
)

func NewCluster(ctx context.Context) *Cluster {
	c := &Cluster{}
	c.AppContext = ctx
	c.table = "clusters"
	c.hasID = true

	return c
}

type ClusterRow struct {
	ID            int64  `db:"id"`
	Name          string `db:"name"`
	CreatorID     int64  `db:"creator_id"`
	CreatorEmail  string `db:"creator_email"`
	DataRetention string `db:"data_retention"`
	Members       string `db:"members"`
}

// GetDataRetention returns DataRetention in map
func (cr *ClusterRow) GetDataRetention() map[string]int {
	retentions := make(map[string]int)
	json.Unmarshal([]byte(cr.DataRetention), &retentions)

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

// GetTTLDurationForInsert returns TTL duration for timeseries data on insert.
func (cr *ClusterRow) GetTTLDurationForInsert(tableName string) time.Duration {
	retention := cr.GetDataRetention()[tableName]
	return time.Duration(retention * 86400) // 1 day = 86400 seconds
}

// GetMembers returns Members in map
func (cr *ClusterRow) GetMembers() []shared.ClusterMember {
	members := make([]shared.ClusterMember, 0)
	json.Unmarshal([]byte(cr.Members), &members)

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

func (c *Cluster) GetCassandraSession() (*gocql.Session, error) {
	cassandradbs, err := contexthelper.GetCassandraDBConfig(c.AppContext)
	if err != nil {
		return nil, err
	}

	return cassandradbs.CoreSession, nil
}

// GetByID returns one record by id.
func (c *Cluster) GetByID(id int64) (*ClusterRow, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	// id bigint primary key,
	// name text,
	// creator_id bigint,
	// creator_email text,
	// data_retention text,
	// members text

	query := fmt.Sprintf("SELECT id, name, creator_id, creator_email, data_retention, members FROM %v WHERE id=? LIMIT 1", c.table)

	var scannedID, scannedCreatorID int64
	var scannedName, scannedCreatorEmail, scannedDataRetention, scannedMembers string

	err = session.Query(query, id).Scan(&scannedID, &scannedName, &scannedCreatorID, &scannedCreatorEmail, &scannedDataRetention, &scannedMembers)
	if err != nil {
		return nil, err
	}

	user := &ClusterRow{
		ID:            scannedID,
		Name:          scannedName,
		CreatorID:     scannedCreatorID,
		CreatorEmail:  scannedCreatorEmail,
		DataRetention: scannedDataRetention,
		Members:       scannedMembers,
	}

	return user, nil
}

// Create a cluster row record with default settings.
func (c *Cluster) Create(creator *shared.UserRow, name string) (*ClusterRow, error) {
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

	dataRetentionJSON, err := json.Marshal(dataRetention)
	if err != nil {
		return nil, err
	}

	member := shared.ClusterMember{}
	member.ID = creator.ID
	member.Email = creator.Email
	member.Level = "write"
	member.Enabled = true

	members := make([]shared.ClusterMember, 1)
	members[0] = member

	membersJSON, err := json.Marshal(members)
	if err != nil {
		return nil, err
	}

	id := NewExplicitID()

	query := fmt.Sprintf("INSERT INTO %v (id, name, creator_id, creator_email, data_retention, members) VALUES (?, ?, ?, ?, ?, ?)", c.table)

	err = session.Query(query, id, name, creator.ID, creator.Email, string(dataRetentionJSON), string(membersJSON)).Exec()
	if err != nil {
		return nil, err
	}

	return c.GetByID(id)
}

// AllByUserID returns all clusters rows by user ID.
func (c *Cluster) AllByUserID(userId int64) ([]*ClusterRow, error) {
	// pgdb, err := c.GetPGDB()
	// if err != nil {
	// 	return nil, err
	// }

	// rows := []*ClusterRow{}

	// query := fmt.Sprintf(`SELECT * from %v WHERE members @> '[{"ID" : %v}]'`, c.table, userId)
	// err = pgdb.Select(&rows, query)
	// if err != nil {
	// 	logrus.WithFields(logrus.Fields{"Query": query}).Error(err)
	// }

	// return rows, err
	return nil, nil
}

// All returns all clusters rows.
func (c *Cluster) All() ([]*ClusterRow, error) {
	session, err := c.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	users := []*ClusterRow{}

	query := fmt.Sprintf(`SELECT id, name, creator_id, creator_email, data_retention, members FROM %v`, c.table)

	var scannedID, scannedCreatorID int64
	var scannedName, scannedCreatorEmail, scannedDataRetention, scannedMembers string

	iter := session.Query(query).Iter()
	for iter.Scan(&scannedID, &scannedName, &scannedCreatorID, &scannedCreatorEmail, &scannedDataRetention, &scannedMembers) {
		users = append(users, &ClusterRow{
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
		logrus.WithFields(logrus.Fields{"Method": "Cluster.All"}).Error(err)

		return nil, err
	}
	return users, err
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
func (c *Cluster) UpdateMember(id int64, user *shared.UserRow, level string, enabled bool) error {
	clusterRow, err := c.GetByID(id)
	if err != nil {
		return err
	}

	foundExisting := false

	members := clusterRow.GetMembers()

	for i, member := range members {
		if member.ID == user.ID {
			memberReplacement := shared.ClusterMember{}
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
		member := shared.ClusterMember{}
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

	data := make(map[string]interface{})
	data["members"] = membersJSON

	// TODO: implement this
	// _, err = c.UpdateByID(tx, data, id)

	return err
}

// RemoveMember from a cluster.
func (c *Cluster) RemoveMember(id int64, user *shared.UserRow) error {
	clusterRow, err := c.GetByID(id)
	if err != nil {
		return err
	}

	members := clusterRow.GetMembers()

	newMembers := make([]shared.ClusterMember, 0)

	for _, member := range members {
		if member.ID != user.ID {
			newMembers = append(newMembers, member)
		}
	}

	newMembersJSON, err := json.Marshal(newMembers)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["members"] = newMembersJSON

	// TODO: implement this
	// _, err = c.UpdateByID(tx, data, id)

	return err
}
