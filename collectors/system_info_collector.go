package collectors

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/DRuggeri/netgear_client"
	"github.com/prometheus/client_golang/prometheus"
)

type SystemInfo struct {
	namespace string
	client    *netgear_client.NetgearClient
	metrics   map[string]prometheus.Gauge

	scrapesTotalMetric              prometheus.Counter
	scrapeErrorsTotalMetric         prometheus.Counter
	lastScrapeErrorMetric           prometheus.Gauge
	lastScrapeTimestampMetric       prometheus.Gauge
	lastScrapeDurationSecondsMetric prometheus.Gauge
}

var SystemInfoFields = [...]string{
	"CPUUtilization",
	"PhysicalMemory",
	"MemoryUtilization",
	"PhysicalFlash",
	"AvailableFlash",
}

func NewSystemInfoCollector(namespace string, client *netgear_client.NetgearClient) *SystemInfo {
	metrics := make(map[string]prometheus.Gauge)
	for _, name := range SystemInfoFields {
		metrics[name] = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "system_info",
				Name:      strings.ToLower(name),
				Help:      fmt.Sprintf("Value of the '%s' system info metric from the router", name),
			},
		)
	}

	scrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "system_info_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Netgear system info stats.",
		},
	)

	scrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "system_info_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrapes errors for Netgear system info stats.",
		},
	)

	lastScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_system_info_scrape_error",
			Help:      "Whether the last scrape of Netgear system info stats resulted in an error (1 for error, 0 for success).",
		},
	)

	lastScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_system_info_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Netgear system info metrics.",
		},
	)

	lastScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_info_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Netgear system info stats.",
		},
	)

	return &SystemInfo{
		namespace: namespace,
		client:    client,
		metrics:   metrics,

		scrapesTotalMetric:              scrapesTotalMetric,
		scrapeErrorsTotalMetric:         scrapeErrorsTotalMetric,
		lastScrapeErrorMetric:           lastScrapeErrorMetric,
		lastScrapeTimestampMetric:       lastScrapeTimestampMetric,
		lastScrapeDurationSecondsMetric: lastScrapeDurationSecondsMetric,
	}
}

func (c *SystemInfo) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	stats, err := c.client.GetSystemInfo()
	if err != nil {
		slog.Error("error while collecting system info: %v", slog.String("error", err.Error()))
		errorMetric = float64(1)
		c.scrapeErrorsTotalMetric.Inc()
	} else {
		/* Loop through the names we expect */
		for _, name := range SystemInfoFields {
			/* Check first that we got what we expect */
			if val, ok := stats[name]; ok {
				var metric float64
				metric, _ = strconv.ParseFloat(val, 64)

				c.metrics[name].Set(metric)
				c.metrics[name].Collect(ch)
			} else {
				slog.Warn(fmt.Sprintf("system info stat named '%s' missing from results!", name))
			}
		}
	}

	c.scrapeErrorsTotalMetric.Collect(ch)

	c.scrapesTotalMetric.Inc()
	c.scrapesTotalMetric.Collect(ch)

	c.lastScrapeErrorMetric.Set(errorMetric)
	c.lastScrapeErrorMetric.Collect(ch)

	c.lastScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastScrapeTimestampMetric.Collect(ch)

	c.lastScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastScrapeDurationSecondsMetric.Collect(ch)
}

func (c *SystemInfo) Describe(ch chan<- *prometheus.Desc) {
	for _, name := range SystemInfoFields {
		c.metrics[name].Describe(ch)
	}

	c.scrapesTotalMetric.Describe(ch)
	c.scrapeErrorsTotalMetric.Describe(ch)
	c.lastScrapeErrorMetric.Describe(ch)
	c.lastScrapeTimestampMetric.Describe(ch)
	c.lastScrapeDurationSecondsMetric.Describe(ch)
}
