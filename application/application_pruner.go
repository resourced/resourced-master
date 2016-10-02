package application

import (
	"github.com/Sirupsen/logrus"
	"github.com/didip/stopwatch"

	"github.com/resourced/resourced-master/libtime"
	"github.com/resourced/resourced-master/models/pg"
)

// PruneAll runs background job to prune all old timeseries data.
func (app *Application) PruneAll() {
	if !app.GeneralConfig.EnablePeriodicPruneJobs {
		return
	}

	for {
		var clusters []*pg.ClusterRow
		var err error

		daemons := make([]string, 0)
		allPeers := app.Peers.Items()

		if len(allPeers) > 0 {
			for hostAndPort, _ := range allPeers {
				daemons = append(daemons, hostAndPort)
			}

			groupedClustersByDaemon, err := pg.NewCluster(app.DBConfig.Core).AllSplitToDaemons(nil, daemons)
			if err != nil {
				app.ErrLogger.WithFields(logrus.Fields{
					"Method": "Cluster.AllSplitToDaemons",
				}).Error(err)

				libtime.SleepString("24h")
				continue
			}

			clusters = groupedClustersByDaemon[app.FullAddr()]

		} else {
			clusters, err = pg.NewCluster(app.DBConfig.Core).All(nil)
		}

		if err != nil {
			app.ErrLogger.WithFields(logrus.Fields{
				"Method": "Application.PruneAll",
			}).Error(err)

			libtime.SleepString("24h")
			continue
		}

		for _, cluster := range clusters {
			go func(cluster *pg.ClusterRow) {
				app.PruneTSCheckOnce(cluster.ID)
			}(cluster)

			go func(cluster *pg.ClusterRow) {
				app.PruneTSMetricOnce(cluster.ID)
			}(cluster)

			go func(cluster *pg.ClusterRow) {
				app.PruneTSMetricAggr15mOnce(cluster.ID)
			}(cluster)

			go func(cluster *pg.ClusterRow) {
				app.PruneTSEventOnce(cluster.ID)
			}(cluster)

			go func(cluster *pg.ClusterRow) {
				app.PruneTSLogOnce(cluster.ID)
			}(cluster)
		}

		libtime.SleepString("24h")
	}
}

// PruneTSCheckOnce deletes old ts_checks data.
func (app *Application) PruneTSCheckOnce(clusterID int64) (err error) {
	f := func() {
		err = pg.NewTSCheck(app.DBConfig.GetTSCheck(clusterID)).DeleteDeleted(nil, clusterID)
	}

	latency := stopwatch.Measure(f)

	logFields := logrus.Fields{
		"Method":       "Application.PruneTSCheckOnce",
		"NanoSeconds":  latency,
		"MicroSeconds": latency / 1000,
		"MilliSeconds": latency / 1000 / 1000,
	}
	if err != nil {
		app.ErrLogger.WithFields(logFields).Error(err)
	} else {
		app.OutLogger.WithFields(logFields).Info("Latency measurement")
	}

	return err
}

// PruneTSMetricOnce deletes old ts_metrics data.
func (app *Application) PruneTSMetricOnce(clusterID int64) (err error) {
	f := func() {
		err = pg.NewTSMetric(app.DBConfig.TSMetric).DeleteDeleted(nil, clusterID)
	}

	latency := stopwatch.Measure(f)

	logFields := logrus.Fields{
		"Method":       "Application.PruneTSMetricOnce",
		"NanoSeconds":  latency,
		"MicroSeconds": latency / 1000,
		"MilliSeconds": latency / 1000 / 1000,
	}
	if err != nil {
		app.ErrLogger.WithFields(logFields).Error(err)
	} else {
		app.OutLogger.WithFields(logFields).Info("Latency measurement")
	}

	return err
}

// PruneTSMetricAggr15mOnce deletes old ts_metrics_aggr_15m data.
func (app *Application) PruneTSMetricAggr15mOnce(clusterID int64) (err error) {
	f := func() {
		err = pg.NewTSMetricAggr15m(app.DBConfig.TSMetricAggr15m).DeleteDeleted(nil, clusterID)
	}

	latency := stopwatch.Measure(f)

	logFields := logrus.Fields{
		"Method":       "Application.PruneTSMetricAggr15mOnce",
		"NanoSeconds":  latency,
		"MicroSeconds": latency / 1000,
		"MilliSeconds": latency / 1000 / 1000,
	}
	if err != nil {
		app.ErrLogger.WithFields(logFields).Error(err)
	} else {
		app.OutLogger.WithFields(logFields).Info("Latency measurement")
	}

	return err
}

// PruneTSEventOnce deletes old ts_events data.
func (app *Application) PruneTSEventOnce(clusterID int64) (err error) {
	f := func() {
		err = pg.NewTSEvent(app.DBConfig.TSEvent).DeleteDeleted(nil, clusterID)
	}

	latency := stopwatch.Measure(f)

	logFields := logrus.Fields{
		"Method":       "Application.PruneTSEventOnce",
		"NanoSeconds":  latency,
		"MicroSeconds": latency / 1000,
		"MilliSeconds": latency / 1000 / 1000,
	}
	if err != nil {
		app.ErrLogger.WithFields(logFields).Error(err)
	} else {
		app.OutLogger.WithFields(logFields).Info("Latency measurement")
	}

	return err
}

// PruneTSLogOnce deletes old ts_logs data.
func (app *Application) PruneTSLogOnce(clusterID int64) (err error) {
	f := func() {
		err = pg.NewTSLog(app.DBConfig.TSLog).DeleteDeleted(nil, clusterID)
	}

	latency := stopwatch.Measure(f)

	logFields := logrus.Fields{
		"Method":       "Application.PruneTSLogOnce",
		"NanoSeconds":  latency,
		"MicroSeconds": latency / 1000,
		"MilliSeconds": latency / 1000 / 1000,
	}
	if err != nil {
		app.ErrLogger.WithFields(logFields).Error(err)
	} else {
		app.OutLogger.WithFields(logFields).Info("Latency measurement")
	}

	return err
}
