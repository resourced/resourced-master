package application

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/rcrowley/go-metrics"
)

func (app *Application) NewMetricsRegistry(handlerInstruments map[string]chan int64, latencyGauges map[string]metrics.Gauge) metrics.Registry {
	r := metrics.NewRegistry()
	metrics.RegisterDebugGCStats(r)
	metrics.RegisterRuntimeMemStats(r)

	for handlerName, _ := range handlerInstruments {
		latencyGauges[handlerName] = metrics.NewGauge()
		r.Register("requests."+handlerName+".ms", latencyGauges[handlerName])
	}

	go metrics.CaptureDebugGCStats(r, time.Second*60)
	go metrics.CaptureRuntimeMemStats(r, time.Second*60)

	// Capture request handlers latency
	for handlerName, latencyChan := range handlerInstruments {
		go func(handlerName string, latencyChan chan int64) {
			for latency := range latencyChan {
				app.OutLogger.WithFields(logrus.Fields{
					"Handler":      handlerName,
					"NanoSeconds":  latency,
					"MicroSeconds": latency / 1000,
					"MilliSeconds": latency / 1000 / 1000,
				}).Info("Capturing latency data")

				app.RLock()
				latencyGauges[handlerName].Update(latency / 1000 / 1000)
				app.RUnlock()
			}
		}(handlerName, latencyChan)
	}

	return r
}
