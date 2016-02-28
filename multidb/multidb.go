package multidb

import (
	"math"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
)

func New(dsns []string, replicationPercentage int) (*MultiDB, error) {
	m := &MultiDB{}
	m.currentIndex = 0
	m.replicationPercentage = replicationPercentage
	m.dsns = dsns
	m.DBs = make([]*sqlx.DB, len(dsns))

	for i, dsn := range dsns {
		db, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}
		m.DBs[i] = db
	}

	return m, nil
}

type MultiDB struct {
	DBs                   []*sqlx.DB
	dsns                  []string
	currentIndex          int
	replicationPercentage int
}

func (mdb *MultiDB) PickRandom() *sqlx.DB {
	rand.Seed(time.Now().Unix())
	index := rand.Intn(len(mdb.DBs))

	return mdb.DBs[index]
}

func (mdb *MultiDB) PickNext() *sqlx.DB {
	mdb.currentIndex = mdb.currentIndex + 1
	if mdb.currentIndex >= len(mdb.DBs) {
		mdb.currentIndex = 0
	}

	return mdb.DBs[mdb.currentIndex]
}

func (mdb *MultiDB) NumOfConnectionsByReplicationPercentage() int {
	return int(
		math.Ceil(
			float64(mdb.replicationPercentage) / float64(100) * float64(len(mdb.DBs))))
}

func (mdb *MultiDB) PickMultipleForWrites() []*sqlx.DB {
	maxDbs := mdb.NumOfConnectionsByReplicationPercentage()
	dbs := make([]*sqlx.DB, len(mdb.DBs))
	permutation := rand.Perm(len(mdb.DBs))

	for i, v := range permutation {
		dbs[v] = mdb.DBs[i]
	}

	return dbs[0:maxDbs]
}
