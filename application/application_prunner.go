package application

import (
	"math"

	"github.com/Sirupsen/logrus"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libtime"
)

func (app *Application) PruneAll() {
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
				err := app.PruneTSCheckOnce(cluster)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":               "Application.PruneTSCheckOnce",
						"DefaultDataRetention": app.GeneralConfig.Checks.DataRetention,
					}).Error(err)
				}
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				err := app.PruneTSMetricOnce(cluster)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":               "Application.PruneTSMetricOnce",
						"DefaultDataRetention": app.GeneralConfig.Metrics.DataRetention,
					}).Error(err)
				}
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				err := app.PruneTSMetricAggr15mOnce(cluster)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":               "Application.PruneTSMetricAggr15mOnce",
						"DefaultDataRetention": app.GeneralConfig.Metrics.DataRetention,
					}).Error(err)
				}
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				err := app.PruneTSEventOnce(cluster)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":               "Application.PruneTSEventOnce",
						"DefaultDataRetention": app.GeneralConfig.Events.DataRetention,
					}).Error(err)
				}
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				err := app.PruneTSExecutorLogOnce(cluster)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":               "Application.PruneTSExecutorLogOnce",
						"DefaultDataRetention": app.GeneralConfig.ExecutorLogs.DataRetention,
					}).Error(err)
				}
			}(cluster)

			go func(cluster *dal.ClusterRow) {
				err := app.PruneTSLogOnce(cluster)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"Method":               "Application.PruneTSLogOnce",
						"DefaultDataRetention": app.GeneralConfig.Logs.DataRetention,
					}).Error(err)
				}
			}(cluster)
		}

		libtime.SleepString("24h")
	}
}

func (app *Application) PruneTSCheckOnce(cluster *dal.ClusterRow) error {
	clusterRetention, ok := cluster.GetDataRetention()["ts_checks"]
	if !ok {
		clusterRetention = 1
	}

	return dal.NewTSCheck(app.DBConfig.TSCheck).DeleteByDayInterval(
		nil,
		int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Checks.DataRetention))),
	)
}

func (app *Application) PruneTSMetricOnce(cluster *dal.ClusterRow) error {
	clusterRetention, ok := cluster.GetDataRetention()["ts_metrics"]
	if !ok {
		clusterRetention = 1
	}

	return dal.NewTSMetric(app.DBConfig.TSMetric).DeleteByDayInterval(
		nil,
		int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Metrics.DataRetention))),
	)
}

func (app *Application) PruneTSMetricAggr15mOnce(cluster *dal.ClusterRow) error {
	clusterRetention, ok := cluster.GetDataRetention()["ts_metrics_aggr_15m"]
	if !ok {
		clusterRetention = 1
	}

	return dal.NewTSMetricAggr15m(app.DBConfig.TSMetric).DeleteByDayInterval(
		nil,
		int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Metrics.DataRetention))),
	)
}

func (app *Application) PruneTSEventOnce(cluster *dal.ClusterRow) error {
	clusterRetention, ok := cluster.GetDataRetention()["ts_events"]
	if !ok {
		clusterRetention = 1
	}

	return dal.NewTSEvent(app.DBConfig.TSEvent).DeleteByDayInterval(
		nil,
		int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Events.DataRetention))),
	)
}

func (app *Application) PruneTSExecutorLogOnce(cluster *dal.ClusterRow) error {
	clusterRetention, ok := cluster.GetDataRetention()["ts_executor_logs"]
	if !ok {
		clusterRetention = 1
	}

	return dal.NewTSExecutorLog(app.DBConfig.TSExecutorLog).DeleteByDayInterval(
		nil,
		int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.ExecutorLogs.DataRetention))),
	)
}

func (app *Application) PruneTSLogOnce(cluster *dal.ClusterRow) error {
	clusterRetention, ok := cluster.GetDataRetention()["ts_logs"]
	if !ok {
		clusterRetention = 1
	}

	return dal.NewTSLog(app.DBConfig.TSLog).DeleteByDayInterval(
		nil,
		int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Logs.DataRetention))),
	)
}
