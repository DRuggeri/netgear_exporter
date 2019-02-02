package collectors

import (
	"github.com/DRuggeri/netgear_client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"strconv"
	"time"
)

type TrafficCollector struct {
	namespace        string
	client           *netgear_client.NetgearClient
	previousIn       float64
	previousOut      float64
	trafficInMetric  prometheus.Gauge
	trafficOutMetric prometheus.Gauge

	trafficScrapesTotalMetric              prometheus.Counter
	trafficScrapeErrorsTotalMetric         prometheus.Counter
	lastTrafficScrapeErrorMetric           prometheus.Gauge
	lastTrafficScrapeTimestampMetric       prometheus.Gauge
	lastTrafficScrapeDurationSecondsMetric prometheus.Gauge
}

func NewTrafficCollector(namespace string, client *netgear_client.NetgearClient) *TrafficCollector {
	trafficInMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "traffic",
			Name:      "download",
			Help:      "Value downloaded since previous check",
		},
	)
	trafficOutMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "traffic",
			Name:      "upload",
			Help:      "Value uploaded since previous check",
		},
	)

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
		trafficInMetric:  trafficInMetric,
		previousIn:       float64(-1),
		trafficOutMetric: trafficOutMetric,
		previousOut:      float64(-1),

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
	}
	c.trafficScrapeErrorsTotalMetric.Collect(ch)

	currentIn, _ := strconv.ParseFloat(stats["TodayDownload"], 64)
	currentOut, _ := strconv.ParseFloat(stats["TodayUpload"], 64)

	/* On the first scrape, the previous values are -1. Since we are calcluating
	   a delta only, pretend this one was zero and update our previous value */
	if c.previousIn < 0 {
		c.previousIn = currentIn
	}
	if c.previousOut < 0 {
		c.previousOut = currentOut
	}

	newIn := currentIn - c.previousIn
	newOut := currentOut - c.previousOut

	log.Infof("In - previous: %v, current: %v, new: %v", c.previousIn, currentIn, newIn)
	log.Infof("Out - previous: %v, current: %v, new: %v", c.previousOut, currentOut, newOut)
	log.Debugf("Raw stats returned:\n")
	for k, v := range stats {
		log.Debugf("  %v => %v", k, v)
	}

	c.previousIn = currentIn
	c.previousOut = currentOut

	/* Metric rolled to next day or this collector started. Assume 0 */
	if newIn < 0 {
		newIn = 0
	}
	if newOut < 0 {
		newOut = 0
	}

	c.trafficInMetric.Set(newIn)
	c.trafficInMetric.Collect(ch)

	c.trafficOutMetric.Set(newOut)
	c.trafficOutMetric.Collect(ch)

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
	c.trafficInMetric.Describe(ch)
	c.trafficOutMetric.Describe(ch)
	c.trafficScrapesTotalMetric.Describe(ch)
	c.trafficScrapeErrorsTotalMetric.Describe(ch)
	c.lastTrafficScrapeErrorMetric.Describe(ch)
	c.lastTrafficScrapeTimestampMetric.Describe(ch)
	c.lastTrafficScrapeDurationSecondsMetric.Describe(ch)
}
