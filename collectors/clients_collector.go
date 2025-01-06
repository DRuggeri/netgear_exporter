package collectors

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/DRuggeri/netgear_client"
	"github.com/prometheus/client_golang/prometheus"
)

type ClientCollector struct {
	namespace              string
	client                 *netgear_client.NetgearClient
	clientsMetric          *prometheus.GaugeVec
	wirelessSpeedMetric    *prometheus.GaugeVec
	wirelessStrengthMetric *prometheus.GaugeVec

	scrapesTotalMetric              prometheus.Counter
	scrapeErrorsTotalMetric         prometheus.Counter
	lastScrapeErrorMetric           prometheus.Gauge
	lastScrapeTimestampMetric       prometheus.Gauge
	lastScrapeDurationSecondsMetric prometheus.Gauge
}

func NewClientCollector(namespace string, client *netgear_client.NetgearClient) *ClientCollector {
	clientsMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "client",
			Name:      "info",
			Help:      "Client information with ip, name, MAC address and connection type labels",
		},
		[]string{"ip", "name", "mac", "connection_type"},
	)

	wirelessSpeedMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "client",
			Name:      "wireless_speed",
			Help:      "Wireless speed of clients connected to the network",
		},
		[]string{"mac"},
	)

	wirelessStrengthMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "client",
			Name:      "wireless_strength",
			Help:      "Wireless strength of clients connected to the network",
		},
		[]string{"mac"},
	)

	scrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "client_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Netgear client stats.",
		},
	)

	scrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "client_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrapes errors for Netgear client stats.",
		},
	)

	lastScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_client_scrape_error",
			Help:      "Whether the last scrape of Netgear client stats resulted in an error (1 for error, 0 for success).",
		},
	)

	lastScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_client_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Netgear client metrics.",
		},
	)

	lastScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_client_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Netgear client stats.",
		},
	)

	return &ClientCollector{
		namespace:              namespace,
		client:                 client,
		clientsMetric:          clientsMetric,
		wirelessSpeedMetric:    wirelessSpeedMetric,
		wirelessStrengthMetric: wirelessStrengthMetric,

		scrapesTotalMetric:              scrapesTotalMetric,
		scrapeErrorsTotalMetric:         scrapeErrorsTotalMetric,
		lastScrapeErrorMetric:           lastScrapeErrorMetric,
		lastScrapeTimestampMetric:       lastScrapeTimestampMetric,
		lastScrapeDurationSecondsMetric: lastScrapeDurationSecondsMetric,
	}
}

func (c *ClientCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	clients, err := c.client.GetAttachDevice()
	if err != nil {
		slog.Error("error while collecting client statistics: %v", slog.String("error", err.Error()))
		errorMetric = float64(1)
		c.scrapeErrorsTotalMetric.Inc()
	} else {
		for _, client := range clients {
			c.clientsMetric.WithLabelValues(
				client["IPAddress"],
				client["Name"],
				client["MACAddress"],
				client["ConnectionType"],
			).Set(float64(1))

			if client["ConnectionType"] != "wired" {
				tmp, _ := strconv.ParseFloat(client["WirelessLinkSpeed"], 64)
				c.wirelessSpeedMetric.WithLabelValues(client["MACAddress"]).Set(tmp)

				tmp, _ = strconv.ParseFloat(client["WirelessSignalStrength"], 64)
				c.wirelessStrengthMetric.WithLabelValues(client["MACAddress"]).Set(tmp)
			}
		}
	}
	c.clientsMetric.Collect(ch)
	c.wirelessSpeedMetric.Collect(ch)
	c.wirelessStrengthMetric.Collect(ch)

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

func (c *ClientCollector) Describe(ch chan<- *prometheus.Desc) {
	c.clientsMetric.Describe(ch)
	c.wirelessSpeedMetric.Describe(ch)
	c.wirelessStrengthMetric.Describe(ch)
	c.scrapesTotalMetric.Describe(ch)
	c.scrapeErrorsTotalMetric.Describe(ch)
	c.lastScrapeErrorMetric.Describe(ch)
	c.lastScrapeTimestampMetric.Describe(ch)
	c.lastScrapeDurationSecondsMetric.Describe(ch)
}
