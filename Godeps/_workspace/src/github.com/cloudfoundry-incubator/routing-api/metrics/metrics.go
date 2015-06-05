package metrics

import (
	"os"
	"time"

	"github.com/cloudfoundry-incubator/routing-api/db"
	"github.com/cloudfoundry/storeadapter"
)

type PartialStatsdClient interface {
	GaugeDelta(stat string, value int64) error
	Gauge(stat string, value int64) error
}

type MetricsReporter struct {
	db       db.DB
	stats    PartialStatsdClient
	ticker   *time.Ticker
	doneChan chan bool
}

func NewMetricsReporter(database db.DB, stats PartialStatsdClient, ticker *time.Ticker) *MetricsReporter {
	return &MetricsReporter{db: database, stats: stats, ticker: ticker}
}

func (r *MetricsReporter) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	eventChan, _, errChan := r.db.WatchRouteChanges()
	close(ready)
	ready = nil

	for {
		select {
		case event := <-eventChan:
			statsDelta := getStatsEventType(event)
			r.stats.GaugeDelta("total_routes", statsDelta)
		case <-r.ticker.C:
			r.stats.Gauge("total_routes", r.getTotalRoutes())
		case <-signals:
			return nil
		case err := <-errChan:
			return err
		}
	}
}

func (r MetricsReporter) getTotalRoutes() int64 {
	routes, _ := r.db.ReadRoutes()
	return int64(len(routes))
}

func getStatsEventType(event storeadapter.WatchEvent) int64 {
	if event.PrevNode == nil && event.Type == storeadapter.UpdateEvent {
		return 1
	} else if event.Type == storeadapter.ExpireEvent || event.Type == storeadapter.DeleteEvent {
		return -1
	} else {
		return 0
	}
}
