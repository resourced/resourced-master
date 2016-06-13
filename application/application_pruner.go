package application

import (
	"math"

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
				app.PruneTSCheckOnce(cluster)
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				app.PruneTSMetricOnce(cluster)
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				app.PruneTSMetricAggr15mOnce(cluster)
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				app.PruneTSEventOnce(cluster)
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				app.PruneTSExecutorLogOnce(cluster)
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				app.PruneTSLogOnce(cluster)
			}(cluster)
		}

		libtime.SleepString("24h")
	}
}

// PruneTSCheckOnce deletes old ts_checks data.
func (app *Application) PruneTSCheckOnce(cluster *dal.ClusterRow) (err error) {
	clusterRetention, ok := cluster.GetDataRetention()["ts_checks"]
	if !ok {
		clusterRetention = 1
	}

	f := func() {
		err = dal.NewTSCheck(app.DBConfig.TSCheck).DeleteByDayInterval(
			nil,
			int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Checks.DataRetention))),
		)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSCheckOnce",
		"DataRetention":       clusterRetention,
		"ClusterID":           cluster.ID,
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
func (app *Application) PruneTSMetricOnce(cluster *dal.ClusterRow) (err error) {
	clusterRetention, ok := cluster.GetDataRetention()["ts_metrics"]
	if !ok {
		clusterRetention = 1
	}

	f := func() {
		err = dal.NewTSMetric(app.DBConfig.TSMetric).DeleteByDayInterval(
			nil,
			int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Metrics.DataRetention))),
		)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSMetricOnce",
		"DataRetention":       clusterRetention,
		"ClusterID":           cluster.ID,
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
func (app *Application) PruneTSMetricAggr15mOnce(cluster *dal.ClusterRow) (err error) {
	clusterRetention, ok := cluster.GetDataRetention()["ts_metrics_aggr_15m"]
	if !ok {
		clusterRetention = 1
	}

	f := func() {
		err = dal.NewTSMetricAggr15m(app.DBConfig.TSMetricAggr15m).DeleteByDayInterval(
			nil,
			int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.MetricsAggr15m.DataRetention))),
		)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSMetricAggr15mOnce",
		"DataRetention":       clusterRetention,
		"ClusterID":           cluster.ID,
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
func (app *Application) PruneTSEventOnce(cluster *dal.ClusterRow) (err error) {
	clusterRetention, ok := cluster.GetDataRetention()["ts_events"]
	if !ok {
		clusterRetention = 1
	}

	f := func() {
		err = dal.NewTSEvent(app.DBConfig.TSEvent).DeleteByDayInterval(
			nil,
			int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Events.DataRetention))),
		)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSEventOnce",
		"DataRetention":       clusterRetention,
		"ClusterID":           cluster.ID,
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
func (app *Application) PruneTSExecutorLogOnce(cluster *dal.ClusterRow) (err error) {
	clusterRetention, ok := cluster.GetDataRetention()["ts_executor_logs"]
	if !ok {
		clusterRetention = 1
	}

	f := func() {
		err = dal.NewTSExecutorLog(app.DBConfig.TSExecutorLog).DeleteByDayInterval(
			nil,
			int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.ExecutorLogs.DataRetention))),
		)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSExecutorLogOnce",
		"DataRetention":       clusterRetention,
		"ClusterID":           cluster.ID,
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
func (app *Application) PruneTSLogOnce(cluster *dal.ClusterRow) (err error) {
	clusterRetention, ok := cluster.GetDataRetention()["ts_logs"]
	if !ok {
		clusterRetention = 1
	}

	f := func() {
		err = dal.NewTSLog(app.DBConfig.TSLog).DeleteByDayInterval(
			nil,
			int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Logs.DataRetention))),
		)
	}

	latency := stopwatch.Measure(f)

	logrusEntry := logrus.WithFields(logrus.Fields{
		"Method":              "Application.PruneTSLogOnce",
		"DataRetention":       clusterRetention,
		"ClusterID":           cluster.ID,
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
