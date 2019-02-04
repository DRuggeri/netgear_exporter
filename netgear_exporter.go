package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/DRuggeri/netgear_client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/DRuggeri/netgear_exporter/collectors"
	"github.com/DRuggeri/netgear_exporter/filters"
)

var (
	netgearUrl = kingpin.Flag(
		"url", "URL of the Netgear router. Defaults to 'https://www.routerlogin.com' ($NETGEAR_EXPORTER_URL)",
	).Envar("NETGEAR_EXPORTER_URL").Default("https://www.routerlogin.com").String()

	netgearUsername = kingpin.Flag(
		"username", "Username to use. Defaults to 'admin' ($NETGEAR_EXPORTER_USERNAME)",
	).Envar("NETGEAR_EXPORTER_USERNAME").Default("admin").String()

	/* Per Prometheus project, it is unacceptable to accept password on command line. See: https://github.com/prometheus/docs/pull/1275#issuecomment-460187844
	netgearPassword = kingpin.Flag(
		"password", "Password to use. ($NETGEAR_EXPORTER_PASSWORD)",
	).Envar("NETGEAR_EXPORTER_PASSWORD").Required().String()
	*/

	netgearInsecure = kingpin.Flag(
		"insecure", "Disable TLS validation of the router. This is needed if you are connecting by IP or a custom host name. Default: false ($NETGEAR_EXPORTER_INSECURE)",
	).Envar("NETGEAR_EXPORTER_INSECURE").Default("false").Bool()

	netgearTimeout = kingpin.Flag(
		"timeout", "Timeout in seconds for communication with the router. On LAN networks, this should be very small. Default: 2 ($NETGEAR_EXPORTER_TIMEOUT)",
	).Envar("NETGEAR_EXPORTER_TIMEOUT").Default("2").Int()

	netgearClientDebug = kingpin.Flag(
		"clientdebug", "Print requests and responses on STDOUT. ($NETGEAR_EXPORTER_CLIENT_DEBUG)",
	).Envar("NETGEAR_EXPORTER_CLIENT_DEBUG").Default("false").Bool()

	filterCollectors = kingpin.Flag(
		"filter.collectors", "Comma separated collectors to filter (Traffic) ($NETGEAR_EXPORTER_FILTER_COLLECTORS)",
	).Envar("NETGEAR_EXPORTER_FILTER_COLLECTORS").Default("").String()

	netgearCalculateTrafficDelta = kingpin.Flag(
		"traffic.calculatedelta", "When enabled, calculates a delta value for in/out bytes. See README.md for warning. ($NETGEAR_EXPORTER_CACLULATE_DELTA)",
	).Envar("NETGEAR_EXPORTER_CALCULATE_DELTA").Default("false").Bool()

	metricsNamespace = kingpin.Flag(
		"metrics.namespace", "Metrics Namespace ($NETGEAR_EXPORTER_METRICS_NAMESPACE)",
	).Envar("NETGEAR_EXPORTER_METRICS_NAMESPACE").Default("netgear").String()

	listenAddress = kingpin.Flag(
		"web.listen-address", "Address to listen on for web interface and telemetry ($NETGEAR_EXPORTER_WEB_LISTEN_ADDRESS)",
	).Envar("NETGEAR_EXPORTER_WEB_LISTEN_ADDRESS").Default(":9192").String()

	metricsPath = kingpin.Flag(
		"web.telemetry-path", "Path under which to expose Prometheus metrics ($NETGEAR_EXPORTER_WEB_TELEMETRY_PATH)",
	).Envar("NETGEAR_EXPORTER_WEB_TELEMETRY_PATH").Default("/metrics").String()

	authUsername = kingpin.Flag(
		"web.auth.username", "Username for web interface basic auth ($NETGEAR_EXPORTER_WEB_AUTH_USERNAME)",
	).Envar("NETGEAR_EXPORTER_WEB_AUTH_USERNAME").String()

	authPassword = kingpin.Flag(
		"web.auth.password", "Password for web interface basic auth ($NETGEAR_EXPORTER_WEB_AUTH_PASSWORD)",
	).Envar("NETGEAR_EXPORTER_WEB_AUTH_PASSWORD").String()

	tlsCertFile = kingpin.Flag(
		"web.tls.cert_file", "Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate ($NETGEAR_EXPORTER_WEB_TLS_CERTFILE)",
	).Envar("NETGEAR_EXPORTER_WEB_TLS_KEYFILE").ExistingFile()

	tlsKeyFile = kingpin.Flag(
		"web.tls.key_file", "Path to a file that contains the TLS private key (PEM format) ($NETGEAR_EXPORTER_WEB_TLS_KEYFILE)",
	).Envar("NETGEAR_EXPORTER_WEB_TLS_KEYFILE").ExistingFile()

	netgearPrintMetrics = kingpin.Flag(
		"printMetrics", "Print the metrics this exporter exposes and exits. Default: false ($NETGEAR_EXPORTER_PRINT_METRICS)",
	).Envar("NETGEAR_EXPORTER_PRINT_METRICS").Default("false").Bool()
)

func init() {
	prometheus.MustRegister(version.NewCollector(*metricsNamespace))
}

type basicAuthHandler struct {
	handler  http.HandlerFunc
	username string
	password string
}

func (h *basicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || username != h.username || password != h.password {
		log.Errorf("Invalid HTTP auth from `%s`", r.RemoteAddr)
		w.Header().Set("WWW-Authenticate", "Basic realm=\"metrics\"")
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	h.handler(w, r)
	return
}

func prometheusHandler() http.Handler {
	handler := prometheus.Handler()

	if *authUsername != "" && *authPassword != "" {
		handler = &basicAuthHandler{
			handler:  prometheus.Handler().ServeHTTP,
			username: *authUsername,
			password: *authPassword,
		}
	}

	return handler
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("netgear_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	if *netgearPrintMetrics {
		/* Make a channel and function to send output along */
		var out chan *prometheus.Desc
		eatOutput := func(in <-chan *prometheus.Desc) {
			for desc := range in {
				/* Weaksauce... no direct access to the variables */
				//Desc{fqName: "the_name", help: "help text", constLabels: {}, variableLabels: []}
				tmp := desc.String()
				vals := strings.Split(tmp, `"`)
				fmt.Printf("  %s - %s\n", vals[1], vals[3])
			}
		}

		/* Interesting juggle here...
		   - Make a channel the describe function can send output to
		   - Start the printing function that consumes the output in the background
		   - Call the describe function to feed the channel (which blocks until the consume function eats a message)
		   - When the describe function exits after returning the last item, close the channel to end the background consume function
		*/
		fmt.Println("Traffic")
		trafficCollector := collectors.NewTrafficCollector(*metricsNamespace, nil, true)
		out = make(chan *prometheus.Desc)
		go eatOutput(out)
		trafficCollector.Describe(out)
		close(out)

		os.Exit(0)
	}

	password := os.Getenv("NETGEAR_EXPORTER_PASSWORD")
	if password == "" {
		os.Stderr.WriteString("ERROR: The password for the SOAP API must be set in the environment variable NETGEAR_EXPORTER_PASSWORD\n")
		os.Exit(1)
	}

	log.Infoln("Starting node_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	netgearClient, err := netgear_client.NewNetgearClient(*netgearUrl, *netgearInsecure, *netgearUsername, password, *netgearTimeout, *netgearClientDebug)
	if err != nil {
		log.Errorf("Error creating Netgear client: %s", err.Error())
		os.Exit(1)
	}

	var collectorsFilters []string
	if *filterCollectors != "" {
		collectorsFilters = strings.Split(*filterCollectors, ",")
	}
	collectorsFilter, err := filters.NewCollectorsFilter(collectorsFilters)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if collectorsFilter.Enabled(filters.TrafficCollector) {
		trafficCollector := collectors.NewTrafficCollector(*metricsNamespace, netgearClient, *netgearCalculateTrafficDelta)
		prometheus.MustRegister(trafficCollector)
	}

	handler := prometheusHandler()
	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Netgear Exporter</title></head>
             <body>
             <h1>Netgear Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	if *tlsCertFile != "" && *tlsKeyFile != "" {
		log.Infoln("Listening TLS on", *listenAddress)
		log.Fatal(http.ListenAndServeTLS(*listenAddress, *tlsCertFile, *tlsKeyFile, nil))
	} else {
		log.Infoln("Listening on", *listenAddress)
		log.Fatal(http.ListenAndServe(*listenAddress, nil))
	}
}
