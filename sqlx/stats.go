package sqlx

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type dbStats struct {
	db                 *sql.DB
	MaxOpenConnections *prometheus.Desc
	OpenConnections    *prometheus.Desc
	InUse              *prometheus.Desc
	Idle               *prometheus.Desc
	WaitCount          *prometheus.Desc
	WaitDuration       *prometheus.Desc
	MaxIdleClosed      *prometheus.Desc
	MaxLifetimeClosed  *prometheus.Desc
}

// NewCollector NewCollector
func NewCollector(db *sql.DB) prometheus.Collector {
	return &dbStats{
		db:                 db,
		MaxOpenConnections: prometheus.NewDesc("db_max_open_connections", "Maximum number of open connections to the database", nil, nil),
		OpenConnections:    prometheus.NewDesc("db_open_connections", "The number of established connections both in use and idle.", nil, nil),
		InUse:              prometheus.NewDesc("db_in_use", "The number of connections currently in use.", nil, nil),
		Idle:               prometheus.NewDesc("db_idle", "The number of idle connections.", nil, nil),
		WaitCount:          prometheus.NewDesc("db_wait_count", "The total number of connections waited for.", nil, nil),
		WaitDuration:       prometheus.NewDesc("db_wait_duration", "The total time blocked waiting for a new connection.", nil, nil),
		MaxIdleClosed:      prometheus.NewDesc("db_max_idle_closed", "The total number of connections closed due to SetMaxIdleConns.", nil, nil),
		MaxLifetimeClosed:  prometheus.NewDesc("db_max_lifetime_closed", "The total number of connections closed due to SetConnMaxLifetime.", nil, nil),
	}
}

func (ds *dbStats) Collect(ch chan<- prometheus.Metric) {
	dbStats := ds.db.Stats()
	ch <- prometheus.MustNewConstMetric(ds.MaxOpenConnections, prometheus.GaugeValue, float64(dbStats.MaxOpenConnections))
	ch <- prometheus.MustNewConstMetric(ds.OpenConnections, prometheus.GaugeValue, float64(dbStats.OpenConnections))
	ch <- prometheus.MustNewConstMetric(ds.InUse, prometheus.GaugeValue, float64(dbStats.InUse))
	ch <- prometheus.MustNewConstMetric(ds.Idle, prometheus.GaugeValue, float64(dbStats.Idle))
	ch <- prometheus.MustNewConstMetric(ds.WaitCount, prometheus.GaugeValue, float64(dbStats.WaitCount))
	ch <- prometheus.MustNewConstMetric(ds.WaitDuration, prometheus.GaugeValue, float64(dbStats.WaitDuration.Microseconds()))
	ch <- prometheus.MustNewConstMetric(ds.MaxIdleClosed, prometheus.GaugeValue, float64(dbStats.MaxIdleClosed))
	ch <- prometheus.MustNewConstMetric(ds.MaxLifetimeClosed, prometheus.GaugeValue, float64(dbStats.MaxLifetimeClosed))
}

func (ds *dbStats) Describe(ch chan<- *prometheus.Desc) {
	ch <- ds.MaxOpenConnections
	ch <- ds.OpenConnections
	ch <- ds.InUse
	ch <- ds.Idle
	ch <- ds.WaitCount
	ch <- ds.WaitDuration
	ch <- ds.MaxIdleClosed
	ch <- ds.MaxLifetimeClosed
}
