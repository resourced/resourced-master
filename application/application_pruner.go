package application

import (
	"github.com/Sirupsen/logrus"
	"github.com/didip/stopwatch"

	"github.com/resourced/resourced-master/libtime"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shims"
)

func (app *Application) myClusters() (clusters []*pg.ClusterRow, err error) {
	daemons := make([]string, 0)
	allPeers := app.Peers.Items()

	if len(allPeers) > 0 {
		for hostAndPort, _ := range allPeers {
			daemons = append(daemons, hostAndPort)
		}

		groupedClustersByDaemon, err := shims.NewCluster(app.GetContext()).AllSplitToDaemons(daemons)
		if err != nil {
			return nil, err
		}

		clusters = groupedClustersByDaemon[app.FullAddr()]

	} else {
		clusters, err = shims.NewCluster(app.GetContext()).All()
	}

	return clusters, err
}

// PruneAll runs background job to prune all old timeseries data.
func (app *Application) PruneAll() {
	if !app.GeneralConfig.EnablePeriodicPruneJobs {
		return
	}

	for {
		clusters, err := app.myClusters()
		if err != nil {
			app.ErrLogger.WithFields(logrus.Fields{
				"Method": "Application.myClusters",
			}).Error(err)

			libtime.SleepString("24h")
			continue
		}

		for _, cluster := range clusters {
			go func(cluster *pg.ClusterRow) {
				app.PruneTSCheckOnce(cluster.ID)
			}(cluster)

			if app.GeneralConfig.GetMetricsDBType() == "pg" {
				go func(cluster *pg.ClusterRow) {
					app.PruneTSMetricOnce(cluster.ID)
				}(cluster)
			}

			if app.GeneralConfig.GetEventsDBType() == "pg" {
				go func(cluster *pg.ClusterRow) {
					app.PruneTSEventOnce(cluster.ID)
				}(cluster)
			}

			if app.GeneralConfig.GetLogsDBType() == "pg" {
				go func(cluster *pg.ClusterRow) {
					app.PruneTSLogOnce(cluster.ID)
				}(cluster)
			}
		}

		libtime.SleepString("24h")
	}
}

// PruneTSCheckOnce deletes old ts_checks data.
func (app *Application) PruneTSCheckOnce(clusterID int64) (err error) {
	f := func() {
		err = pg.NewTSCheck(app.GetContext(), clusterID).DeleteDeleted(nil, clusterID)
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
	if app.GeneralConfig.GetMetricsDBType() != "pg" {
		return nil
	}

	f := func() {
		err = pg.NewTSMetric(app.GetContext(), clusterID).DeleteDeleted(nil, clusterID)
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

// PruneTSEventOnce deletes old ts_events data.
func (app *Application) PruneTSEventOnce(clusterID int64) (err error) {
	f := func() {
		err = pg.NewTSEvent(app.GetContext(), clusterID).DeleteDeleted(nil, clusterID)
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
		err = pg.NewTSLog(app.GetContext(), clusterID).DeleteDeleted(nil, clusterID)
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
