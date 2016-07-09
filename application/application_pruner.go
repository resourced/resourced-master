package application

import (
	"github.com/Sirupsen/logrus"
	"github.com/didip/stopwatch"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libtime"
)

// PruneAll runs background job to prune all old timeseries data.
func (app *Application) PruneAll() {
	if !app.GeneralConfig.EnablePeriodicPruneJobs {
		return
	}

	for {
		var clusters []*dal.ClusterRow
		var err error

		daemons := make([]string, 0)
		allPeers := app.Peers.All()

		if len(allPeers) > 0 {
			for hostAndPort, _ := range allPeers {
				daemons = append(daemons, hostAndPort)
			}

			groupedClustersByDaemon, err := dal.NewCluster(app.DBConfig.Core).AllSplitToDaemons(nil, daemons)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method": "Cluster.AllSplitToDaemons",
				}).Error(err)

				libtime.SleepString("24h")
				continue
			}

			clusters = groupedClustersByDaemon[app.FullAddr()]

		} else {
			clusters, err = dal.NewCluster(app.DBConfig.Core).All(nil)
		}

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method": "Application.PruneAll",
			}).Error(err)

			libtime.SleepString("24h")
			continue
		}

		for _, cluster := range clusters {
			go func(cluster *dal.ClusterRow) {
				app.PruneTSCheckOnce(cluster.ID)
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				app.PruneTSMetricOnce(cluster.ID)
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				app.PruneTSMetricAggr15mOnce(cluster.ID)
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				app.PruneTSEventOnce(cluster.ID)
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				app.PruneTSExecutorLogOnce(cluster.ID)
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				app.PruneTSLogOnce(cluster.ID)
			}(cluster)
		}

		libtime.SleepString("24h")
	}
}

// PruneTSCheckOnce deletes old ts_checks data.
func (app *Application) PruneTSCheckOnce(clusterID int64) (err error) {
	f := func() {
		err = dal.NewTSCheck(app.DBConfig.GetTSCheck(clusterID)).DeleteDeleted(nil, clusterID)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSCheckOnce",
		"LatencyNanoSeconds":  latency,
		"LatencyMicroSeconds": latency / 1000,
		"LatencyMilliSeconds": latency / 1000 / 1000,
	})
	if err != nil {
		logrusEntry.Error(err)
	} else {
		logrusEntry.Info("Latency measurement")
	}

	return err
}

// PruneTSMetricOnce deletes old ts_metrics data.
func (app *Application) PruneTSMetricOnce(clusterID int64) (err error) {
	f := func() {
		err = dal.NewTSMetric(app.DBConfig.TSMetric).DeleteDeleted(nil, clusterID)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSMetricOnce",
		"LatencyNanoSeconds":  latency,
		"LatencyMicroSeconds": latency / 1000,
		"LatencyMilliSeconds": latency / 1000 / 1000,
	})
	if err != nil {
		logrusEntry.Error(err)
	} else {
		logrusEntry.Info("Latency measurement")
	}

	return err
}

// PruneTSMetricAggr15mOnce deletes old ts_metrics_aggr_15m data.
func (app *Application) PruneTSMetricAggr15mOnce(clusterID int64) (err error) {
	f := func() {
		err = dal.NewTSMetricAggr15m(app.DBConfig.TSMetricAggr15m).DeleteDeleted(nil, clusterID)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSMetricAggr15mOnce",
		"LatencyNanoSeconds":  latency,
		"LatencyMicroSeconds": latency / 1000,
		"LatencyMilliSeconds": latency / 1000 / 1000,
	})
	if err != nil {
		logrusEntry.Error(err)
	} else {
		logrusEntry.Info("Latency measurement")
	}

	return err
}

// PruneTSEventOnce deletes old ts_events data.
func (app *Application) PruneTSEventOnce(clusterID int64) (err error) {
	f := func() {
		err = dal.NewTSEvent(app.DBConfig.TSEvent).DeleteDeleted(nil, clusterID)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSEventOnce",
		"LatencyNanoSeconds":  latency,
		"LatencyMicroSeconds": latency / 1000,
		"LatencyMilliSeconds": latency / 1000 / 1000,
	})
	if err != nil {
		logrusEntry.Error(err)
	} else {
		logrusEntry.Info("Latency measurement")
	}

	return err
}

// PruneTSExecutorLogOnce deletes old ts_executor_logs data.
func (app *Application) PruneTSExecutorLogOnce(clusterID int64) (err error) {
	f := func() {
		err = dal.NewTSExecutorLog(app.DBConfig.TSExecutorLog).DeleteDeleted(nil, clusterID)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSExecutorLogOnce",
		"LatencyNanoSeconds":  latency,
		"LatencyMicroSeconds": latency / 1000,
		"LatencyMilliSeconds": latency / 1000 / 1000,
	})
	if err != nil {
		logrusEntry.Error(err)
	} else {
		logrusEntry.Info("Latency measurement")
	}

	return err
}

// PruneTSLogOnce deletes old ts_logs data.
func (app *Application) PruneTSLogOnce(clusterID int64) (err error) {
	f := func() {
		err = dal.NewTSLog(app.DBConfig.TSLog).DeleteDeleted(nil, clusterID)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSLogOnce",
		"LatencyNanoSeconds":  latency,
		"LatencyMicroSeconds": latency / 1000,
		"LatencyMilliSeconds": latency / 1000 / 1000,
	})
	if err != nil {
		logrusEntry.Error(err)
	} else {
		logrusEntry.Info("Latency measurement")
	}

	return err
}
