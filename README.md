# Netgear Router Prometheus Exporter

A [Prometheus](https://prometheus.io) exporter for Netgear consumer routers. This exporter consumes the data obtained by [netgear_client](https://github.com/DRuggeri/netgear_client) and is based on the [node_exporter](https://github.com/prometheus/node_exporter) and [cf_exporter](https://github.com/bosh-prometheus/cf_exporter) projects.

## Installation

### Binaries

Download the already existing [binaries](https://github.com/DRuggeri/netgear_exporter/releases) for your platform:

```bash
$ ./netgear_exporter <flags>
```

### From source

Using the standard `go install` (you must have [Go](https://golang.org/) already installed in your local machine):

```bash
$ go install github.com/DRuggeri/netgear_exporter
$ netgear_exporter <flags>
```

### With Docker
```bash
docker build -t netgear_exporter .
docker run -d -p 9192:9192 -e NETGEAR_EXPORTER_PASSWORD=YOUR_PASSWORD netgear_exporter --insecure --url="http://YOUR_IP_ADDRESS" --username="YOUR_USERNAME"
```

### Cloud Foundry

The exporter can be deployed to an already existing [Cloud Foundry](https://www.cloudfoundry.org/) environment if, for some reason, that lives on a network with your Netgear consumer router.

```bash
$ git clone https://github.com/DRuggeri/netgear_exporter.git
$ cd cf_exporter
```

Modify the included [application manifest file](https://github.com/DRuggeri/netgear_exporter/blob/master/manifest.yml) to include your router properties. Then you can push the exporter to your Cloud Foundry environment:

```bash
$ cf push
```


## Usage

### Flags

```
  -h, --help                    Show context-sensitive help (also try --help-long and --help-man).
      --url="https://www.routerlogin.com"  
                                URL of the Netgear router. Defaults to 'https://www.routerlogin.com' ($NETGEAR_EXPORTER_URL)
      --username="admin"        Username to use. Defaults to 'admin' ($NETGEAR_EXPORTER_USERNAME)
      --password=PASSWORD       Password to use. ($NETGEAR_EXPORTER_PASSWORD)
      --insecure                Disable TLS validation of the router. This is needed if you are connecting by IP or a custom host name. Default: false
                                ($NETGEAR_EXPORTER_INSECURE)
      --timeout=2               Timeout in seconds for communication with the router. On LAN networks, this should be very small. Default: 2 ($NETGEAR_EXPORTER_TIMEOUT)
      --clientdebug             Print requests and responses on STDOUT. ($NETGEAR_EXPORTER_CLIENT_DEBUG)
      --filter.collectors=""    Comma separated collectors to filter (Traffic) ($NETGEAR_EXPORTER_FILTER_COLLECTORS)
      --traffic.calculatedelta  When enabled, calculates a delta value for in/out bytes. See README.md for warning. ($NETGEAR_EXPORTER_CACLULATE_DELTA)
      --metrics.namespace="netgear"  
                                Metrics Namespace ($NETGEAR_EXPORTER_METRICS_NAMESPACE)
      --web.listen-address=":9192"  
                                Address to listen on for web interface and telemetry ($NETGEAR_EXPORTER_WEB_LISTEN_ADDRESS)
      --web.telemetry-path="/metrics"  
                                Path under which to expose Prometheus metrics ($NETGEAR_EXPORTER_WEB_TELEMETRY_PATH)
      --web.auth.username=WEB.AUTH.USERNAME  
                                Username for web interface basic auth ($NETGEAR_EXPORTER_WEB_AUTH_USERNAME)
      --web.auth.password=WEB.AUTH.PASSWORD  
                                Password for web interface basic auth ($NETGEAR_EXPORTER_WEB_AUTH_PASSWORD)
      --web.tls.cert_file=WEB.TLS.CERT_FILE  
                                Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should
                                be the concatenation of the server's certificate, any intermediates, and the CA's certificate ($NETGEAR_EXPORTER_WEB_TLS_CERTFILE)
      --web.tls.key_file=WEB.TLS.KEY_FILE  
                                Path to a file that contains the TLS private key (PEM format) ($NETGEAR_EXPORTER_WEB_TLS_KEYFILE)
      --printMetrics            Print the metrics this exporter exposes and exits. Default: false ($NETGEAR_EXPORTER_PRINT_METRICS)
      --log.level="info"        Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]
      --log.format="logger:stderr"  
                                Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"
      --version                 Show application version.
```

## Metrics

### Traffic
This collector gathers the raw traffic data from the router. The time metrics are converted from `hh:mm` format to number of seconds.

**IMPORTANT NOTE:** Netgear implements these statistics as incrementing counters that reset after their prescribed duration (day, week, month). As such, these metrics should mostly show "sawtooth" style data when graphed.

```
  netgear_traffic_todayconnectiontime - Value of the 'TodayConnectionTime' traffic metric from the router
  netgear_traffic_todaydownload - Value of the 'TodayDownload' traffic metric from the router
  netgear_traffic_todayupload - Value of the 'TodayUpload' traffic metric from the router
  netgear_traffic_yesterdayconnectiontime - Value of the 'YesterdayConnectionTime' traffic metric from the router
  netgear_traffic_yesterdaydownload - Value of the 'YesterdayDownload' traffic metric from the router
  netgear_traffic_yesterdayupload - Value of the 'YesterdayUpload' traffic metric from the router
  netgear_traffic_weekconnectiontime - Value of the 'WeekConnectionTime' traffic metric from the router
  netgear_traffic_weekdownload - Value of the 'WeekDownload' traffic metric from the router
  netgear_traffic_weekdownloadaverage - Value of the 'WeekDownloadAverage' traffic metric from the router
  netgear_traffic_weekupload - Value of the 'WeekUpload' traffic metric from the router
  netgear_traffic_weekuploadaverage - Value of the 'WeekUploadAverage' traffic metric from the router
  netgear_traffic_monthconnectiontime - Value of the 'MonthConnectionTime' traffic metric from the router
  netgear_traffic_monthdownload - Value of the 'MonthDownload' traffic metric from the router
  netgear_traffic_monthdownloadaverage - Value of the 'MonthDownloadAverage' traffic metric from the router
  netgear_traffic_monthupload - Value of the 'MonthUpload' traffic metric from the router
  netgear_traffic_monthuploadaverage - Value of the 'MonthUploadAverage' traffic metric from the router
  netgear_traffic_lastmonthconnectiontime - Value of the 'LastMonthConnectionTime' traffic metric from the router
  netgear_traffic_lastmonthdownload - Value of the 'LastMonthDownload' traffic metric from the router
  netgear_traffic_lastmonthdownloadaverage - Value of the 'LastMonthDownloadAverage' traffic metric from the router
  netgear_traffic_lastmonthupload - Value of the 'LastMonthUpload' traffic metric from the router
  netgear_traffic_lastmonthuploadaverage - Value of the 'LastMonthUploadAverage' traffic metric from the router
  netgear_traffic_scrapes_total - Total number of scrapes for Netgear traffic stats.
  netgear_traffic_scrape_errors_total - Total number of scrapes errors for Netgear traffic stats.
  netgear_last_traffic_scrape_error - Whether the last scrape of Netgear traffic stats resulted in an error (1 for error, 0 for success).
  netgear_last_traffic_scrape_timestamp - Number of seconds since 1970 since last scrape of Netgear traffic metrics.
```

#### Additional metrics available with `traffic.calculatedelta`

**IMPORTANT NOTE:** This portion of the collector *IS NOT* capable of handling concurrent scrapes and scrapes from multiple clients. This is because the Netgear routers provide current traffic statistics as incrementing counters that reset each day.
In order to detect the amount of traffic that has been passed since the previous scrape, the collector keeps track of the previous result.
The following metrics therefore rely on having only one client scraping the exporter at a time.

```
  netgear_traffic_download - Value downloaded since previous check
  netgear_traffic_upload - Value uploaded since previous check
```

## Contributing

Refer to the [contributing guidelines](https://github.com/DRuggeri/netgear_exporter/blob/master/CONTRIBUTING.md).

## License

Apache License 2.0, see [LICENSE](https://github.com/DRuggeri/netgear_exporter/blob/master/LICENSE).
