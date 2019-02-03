package collectors

import (
	"github.com/DRuggeri/netgear_client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"strconv"
	"time"
)

type TrafficDeltaCollector struct {
	namespace        string
	client           *netgear_client.NetgearClient
	previousIn       float64
	previousOut      float64
	trafficDeltaInMetric  prometheus.Gauge
	trafficDeltaOutMetric prometheus.Gauge

	trafficDeltaScrapesTotalMetric              prometheus.Counter
	trafficDeltaScrapeErrorsTotalMetric         prometheus.Counter
	lastTrafficDeltaScrapeErrorMetric           prometheus.Gauge
	lastTrafficDeltaScrapeTimestampMetric       prometheus.Gauge
	lastTrafficDeltaScrapeDurationSecondsMetric prometheus.Gauge
}

func NewTrafficDeltaCollector(namespace string, client *netgear_client.NetgearClient) *TrafficDeltaCollector {
	trafficDeltaInMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "trafficdelta",
			Name:      "download",
			Help:      "Value downloaded since previous check",
		},
	)
	trafficDeltaOutMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "trafficdelta",
			Name:      "upload",
			Help:      "Value uploaded since previous check",
		},
	)

	trafficDeltaScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "trafficdelta_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Netgear traffic delta stats.",
		},
	)

	trafficDeltaScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "trafficdelta_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrapes errors for Netgear traffic delta stats.",
		},
	)

	lastTrafficDeltaScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_trafficdelta_scrape_error",
			Help:      "Whether the last scrape of Netgear traffic delta stats resulted in an error (1 for error, 0 for success).",
		},
	)

	lastTrafficDeltaScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_trafficdelta_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Netgear traffic delta metrics.",
		},
	)

	lastTrafficDeltaScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_trafficdelta_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Netgear traffic delta stats.",
		},
	)

	return &TrafficDeltaCollector{
		namespace:        namespace,
		client:           client,
		trafficDeltaInMetric:  trafficDeltaInMetric,
		previousIn:       float64(-1),
		trafficDeltaOutMetric: trafficDeltaOutMetric,
		previousOut:      float64(-1),

		trafficDeltaScrapesTotalMetric:              trafficDeltaScrapesTotalMetric,
		trafficDeltaScrapeErrorsTotalMetric:         trafficDeltaScrapeErrorsTotalMetric,
		lastTrafficDeltaScrapeErrorMetric:           lastTrafficDeltaScrapeErrorMetric,
		lastTrafficDeltaScrapeTimestampMetric:       lastTrafficDeltaScrapeTimestampMetric,
		lastTrafficDeltaScrapeDurationSecondsMetric: lastTrafficDeltaScrapeDurationSecondsMetric,
	}
}

func (c *TrafficDeltaCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	stats, err := c.client.GetTrafficMeterStatistics()
	if err != nil {
		log.Errorf("Error while collecting traffic statistics: %v", err)
		errorMetric = float64(1)
		c.trafficDeltaScrapeErrorsTotalMetric.Inc()
	}
	c.trafficDeltaScrapeErrorsTotalMetric.Collect(ch)

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

	c.trafficDeltaInMetric.Set(newIn)
	c.trafficDeltaInMetric.Collect(ch)

	c.trafficDeltaOutMetric.Set(newOut)
	c.trafficDeltaOutMetric.Collect(ch)

	c.trafficDeltaScrapesTotalMetric.Inc()
	c.trafficDeltaScrapesTotalMetric.Collect(ch)

	c.lastTrafficDeltaScrapeErrorMetric.Set(errorMetric)
	c.lastTrafficDeltaScrapeErrorMetric.Collect(ch)

	c.lastTrafficDeltaScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastTrafficDeltaScrapeTimestampMetric.Collect(ch)

	c.lastTrafficDeltaScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastTrafficDeltaScrapeDurationSecondsMetric.Collect(ch)

}

func (c *TrafficDeltaCollector) Describe(ch chan<- *prometheus.Desc) {
	c.trafficDeltaInMetric.Describe(ch)
	c.trafficDeltaOutMetric.Describe(ch)
	c.trafficDeltaScrapesTotalMetric.Describe(ch)
	c.trafficDeltaScrapeErrorsTotalMetric.Describe(ch)
	c.lastTrafficDeltaScrapeErrorMetric.Describe(ch)
	c.lastTrafficDeltaScrapeTimestampMetric.Describe(ch)
	c.lastTrafficDeltaScrapeDurationSecondsMetric.Describe(ch)
}
