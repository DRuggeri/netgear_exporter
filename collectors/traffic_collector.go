package collectors

import (
	"github.com/DRuggeri/netgear_client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"time"
	"fmt"
	"strings"
	"strconv"
)

type TrafficCollector struct {
	namespace        string
	client           *netgear_client.NetgearClient
	metrics		map[string]prometheus.Gauge

	trafficScrapesTotalMetric              prometheus.Counter
	trafficScrapeErrorsTotalMetric         prometheus.Counter
	lastTrafficScrapeErrorMetric           prometheus.Gauge
	lastTrafficScrapeTimestampMetric       prometheus.Gauge
	lastTrafficScrapeDurationSecondsMetric prometheus.Gauge
}

var TrafficCollectorFields = [...]string {
  "TodayConnectionTime",
  "TodayDownload",
  "TodayUpload",
  "YesterdayConnectionTime",
  "YesterdayDownload",
  "YesterdayUpload",
  "WeekConnectionTime",
  "WeekDownload",
  "WeekDownloadAverage",
  "WeekUpload",
  "WeekUploadAverage",
  "MonthConnectionTime",
  "MonthDownload",
  "MonthDownloadAverage",
  "MonthUpload",
  "MonthUploadAverage",
  "LastMonthConnectionTime",
  "LastMonthDownload",
  "LastMonthDownloadAverage",
  "LastMonthUpload",
  "LastMonthUploadAverage",
}

func NewTrafficCollector(namespace string, client *netgear_client.NetgearClient) *TrafficCollector {
	metrics := make(map[string]prometheus.Gauge)
	for _, name := range TrafficCollectorFields {
		metrics[name] = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "traffic",
				Name:      strings.ToLower(name),
				Help:      fmt.Sprintf("Value of the '%s' traffic metric from the router", name),
			},
		)
	}
	trafficScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "traffic_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Netgear traffic stats.",
		},
	)

	trafficScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "traffic_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrapes errors for Netgear traffic stats.",
		},
	)

	lastTrafficScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_traffic_scrape_error",
			Help:      "Whether the last scrape of Netgear traffic stats resulted in an error (1 for error, 0 for success).",
		},
	)

	lastTrafficScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_traffic_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Netgear traffic metrics.",
		},
	)

	lastTrafficScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_traffic_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Netgear traffic stats.",
		},
	)

	return &TrafficCollector{
		namespace:        namespace,
		client:           client,
		metrics:	metrics,

		trafficScrapesTotalMetric:              trafficScrapesTotalMetric,
		trafficScrapeErrorsTotalMetric:         trafficScrapeErrorsTotalMetric,
		lastTrafficScrapeErrorMetric:           lastTrafficScrapeErrorMetric,
		lastTrafficScrapeTimestampMetric:       lastTrafficScrapeTimestampMetric,
		lastTrafficScrapeDurationSecondsMetric: lastTrafficScrapeDurationSecondsMetric,
	}
}

func (c *TrafficCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	stats, err := c.client.GetTrafficMeterStatistics()
	if err != nil {
		log.Errorf("Error while collecting traffic statistics: %v", err)
		errorMetric = float64(1)
		c.trafficScrapeErrorsTotalMetric.Inc()
	} else {
		/* Loop through the names we expect */
		for _, name := range TrafficCollectorFields {
			/* Check first that we got what we expect */
			if val, ok := stats[name]; ok {
				var metric float64
				/* Convert time entries in H:M format to seconds */
				if strings.HasSuffix(name, "Time") {
					times := strings.Split(val, ":")
					tmp, _ := strconv.ParseFloat(times[0], 64)
					metric = tmp * float64(3600)
					tmp, _ = strconv.ParseFloat(times[0], 64)
					metric += tmp * float64(60)
				} else {
					metric, _ = strconv.ParseFloat(val, 64)
				}

				c.metrics[name].Set(metric)
				c.metrics[name].Collect(ch)
			} else {
				log.Warnf("Traffic stat named '%s' missing from results!", name)
			}
		}
	}
	c.trafficScrapeErrorsTotalMetric.Collect(ch)

	c.trafficScrapesTotalMetric.Inc()
	c.trafficScrapesTotalMetric.Collect(ch)

	c.lastTrafficScrapeErrorMetric.Set(errorMetric)
	c.lastTrafficScrapeErrorMetric.Collect(ch)

	c.lastTrafficScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastTrafficScrapeTimestampMetric.Collect(ch)

	c.lastTrafficScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastTrafficScrapeDurationSecondsMetric.Collect(ch)
}

func (c *TrafficCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, name := range TrafficCollectorFields {
		c.metrics[name].Describe(ch)
	}
	c.trafficScrapesTotalMetric.Describe(ch)
	c.trafficScrapeErrorsTotalMetric.Describe(ch)
	c.lastTrafficScrapeErrorMetric.Describe(ch)
	c.lastTrafficScrapeTimestampMetric.Describe(ch)
	c.lastTrafficScrapeDurationSecondsMetric.Describe(ch)
}
