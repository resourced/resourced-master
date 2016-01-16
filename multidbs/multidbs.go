package multidbs

import (
	"math"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
)

func New(dsns []string, replicationPercentage int) (*MultiDBs, error) {
	m := &MultiDBs{}
	m.currentIndex = 0
	m.replicationPercentage = replicationPercentage
	m.dsns = dsns
	m.dbs = make([]*sqlx.DB, len(dsns))

	for i, dsn := range dsns {
		db, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}
		m.dbs[i] = db
	}

	return m, nil
}

type MultiDBs struct {
	dbs                   []*sqlx.DB
	dsns                  []string
	currentIndex          int
	replicationPercentage int
}

func (mdb *MultiDBs) PickRandom() *sqlx.DB {
	rand.Seed(time.Now().Unix())
	index := rand.Intn(len(mdb.dbs))

	return mdb.dbs[index]
}

func (mdb *MultiDBs) PickNext() *sqlx.DB {
	mdb.currentIndex = mdb.currentIndex + 1
	if mdb.currentIndex >= len(mdb.dbs) {
		mdb.currentIndex = 0
	}

	return mdb.dbs[mdb.currentIndex]
}

func (mdb *MultiDBs) NumOfConnectionsByReplicationPercentage() int {
	return int(math.Ceil(float64(mdb.replicationPercentage / 100 * len(mdb.dbs))))
}

func (mdb *MultiDBs) PickMultipleForWrites() []*sqlx.DB {
	dbs := make([]*sqlx.DB, mdb.NumOfConnectionsByReplicationPercentage())

	for i := 0; i < len(dbs); i++ {
		dbs[i] = mdb.PickRandom()
	}

	return dbs
}
