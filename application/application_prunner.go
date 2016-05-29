// Package application allows the creation of Application struct.
// There's only one Application per main().
package application

import (
	"math"

	"github.com/Sirupsen/logrus"

	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libtime"
)

func (app *Application) PruneAll() {
	for {
		clusters, err := dal.NewCluster(app.DBConfig.Core).All(nil)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Method": "Application.PruneAll",
			}).Error(err)

			libtime.SleepString("24h")
			continue
		}

		app.PruneTSCheckOnce(clusters)
		app.PruneTSMetricOnce(clusters)
		app.PruneTSMetricAggr15mOnce(clusters)
		app.PruneTSEventOnce(clusters)
		app.PruneTSExecutorLogOnce(clusters)
		app.PruneTSLogOnce(clusters)

		libtime.SleepString("24h")
	}
}

func (app *Application) PruneTSCheckOnce(clusters []*dal.ClusterRow) {
	for _, cluster := range clusters {
		go func(cluster *dal.ClusterRow) {
			clusterRetentions := cluster.GetDataRetention()
			clusterRetention, ok := clusterRetentions["ts_checks"]
			if !ok {
				clusterRetention = 1
			}

			err := dal.NewTSCheck(app.DBConfig.TSCheck).DeleteByDayInterval(
				nil,
				int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Checks.DataRetention))),
			)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method":                "TSCheck.DeleteByDayInterval",
					"cluster.DataRetention": clusterRetentions,
					"DataRetention":         app.GeneralConfig.Checks.DataRetention,
				}).Error(err)
			}
		}(cluster)
	}
}

func (app *Application) PruneTSMetricOnce(clusters []*dal.ClusterRow) {
	for _, cluster := range clusters {
		go func(cluster *dal.ClusterRow) {
			clusterRetentions := cluster.GetDataRetention()
			clusterRetention, ok := clusterRetentions["ts_metrics"]
			if !ok {
				clusterRetention = 1
			}

			err := dal.NewTSMetric(app.DBConfig.TSMetric).DeleteByDayInterval(
				nil,
				int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Metrics.DataRetention))),
			)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method":                "TSMetric.DeleteByDayInterval",
					"cluster.DataRetention": clusterRetentions,
					"DataRetention":         app.GeneralConfig.Metrics.DataRetention,
				}).Error(err)
			}
		}(cluster)
	}
}

func (app *Application) PruneTSMetricAggr15mOnce(clusters []*dal.ClusterRow) {
	for _, cluster := range clusters {
		go func(cluster *dal.ClusterRow) {
			clusterRetentions := cluster.GetDataRetention()
			clusterRetention, ok := clusterRetentions["ts_metrics_aggr_15m"]
			if !ok {
				clusterRetention = 1
			}

			err := dal.NewTSMetricAggr15m(app.DBConfig.TSMetric).DeleteByDayInterval(
				nil,
				int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Metrics.DataRetention))),
			)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method":                "TSMetricAggr15m.DeleteByDayInterval",
					"cluster.DataRetention": clusterRetentions,
					"DataRetention":         app.GeneralConfig.Metrics.DataRetention,
				}).Error(err)
			}
		}(cluster)
	}
}

func (app *Application) PruneTSEventOnce(clusters []*dal.ClusterRow) {
	for _, cluster := range clusters {
		go func(cluster *dal.ClusterRow) {
			clusterRetentions := cluster.GetDataRetention()
			clusterRetention, ok := clusterRetentions["ts_events"]
			if !ok {
				clusterRetention = 1
			}

			err := dal.NewTSEvent(app.DBConfig.TSEvent).DeleteByDayInterval(
				nil,
				int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Events.DataRetention))),
			)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method":                "TSEvent.DeleteByDayInterval",
					"cluster.DataRetention": clusterRetentions,
					"DataRetention":         app.GeneralConfig.Events.DataRetention,
				}).Error(err)
			}
		}(cluster)
	}
}

func (app *Application) PruneTSExecutorLogOnce(clusters []*dal.ClusterRow) {
	for _, cluster := range clusters {
		go func(cluster *dal.ClusterRow) {
			clusterRetentions := cluster.GetDataRetention()
			clusterRetention, ok := clusterRetentions["ts_executor_logs"]
			if !ok {
				clusterRetention = 1
			}

			err := dal.NewTSExecutorLog(app.DBConfig.TSExecutorLog).DeleteByDayInterval(
				nil,
				int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.ExecutorLogs.DataRetention))),
			)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method":                "TSExecutorLog.DeleteByDayInterval",
					"cluster.DataRetention": clusterRetentions,
					"DataRetention":         app.GeneralConfig.ExecutorLogs.DataRetention,
				}).Error(err)
			}
		}(cluster)
	}
}

func (app *Application) PruneTSLogOnce(clusters []*dal.ClusterRow) {
	for _, cluster := range clusters {
		go func(cluster *dal.ClusterRow) {
			clusterRetentions := cluster.GetDataRetention()
			clusterRetention, ok := clusterRetentions["ts_logs"]
			if !ok {
				clusterRetention = 1
			}

			err := dal.NewTSLog(app.DBConfig.TSLog).DeleteByDayInterval(
				nil,
				int(math.Max(float64(clusterRetention), float64(app.GeneralConfig.Logs.DataRetention))),
			)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Method":                "TSLog.DeleteByDayInterval",
					"cluster.DataRetention": clusterRetentions,
					"DataRetention":         app.GeneralConfig.Logs.DataRetention,
				}).Error(err)
			}
		}(cluster)
	}
}
