package main

import (
	"bufio"
	"flag"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var version = "dev"

var (
	statsFile     = flag.String("f", "/tmp/sfptpd_stats.jsonl", "sfptpd stats JSONL file")
	metricsListen = flag.String("l", ":9979", "metrics listen address")
	verbose       = flag.Bool("v", false, "Enable verbose logging")
	trace         = flag.Bool("vv", false, "Enable extra verbose logging")
)

var (
	metricLastUpdate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfptpd_last_update",
		Help: "Last time we got an update from sfptpd",
	}, []string{"instance"})
	metricTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfptpd_time",
	}, []string{"instance"})
	metricMaster = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfptpd_master",
	}, []string{"instance", "name"})
	metricSlave = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfptpd_slave",
	}, []string{"instance", "name", "primary_interface"})
	metricIsDisciplining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfptpd_is_disciplining",
	}, []string{"instance"})
	metricInSync = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfptpd_in_sync",
	}, []string{"instance"})
	metricAlarms = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfptpd_alarms",
	}, []string{"instance"})
	metricOffset = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfptpd_offset",
	}, []string{"instance"})
	metricFreqAdj = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfptpd_freq_adj",
	}, []string{"instance"})
	metricPTerm = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfptpd_pterm",
	}, []string{"instance"})
	metricITerm = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sfptpd_iterm",
	}, []string{"instance"})
)

func main() {
	flag.Parse()
	if *verbose {
		log.SetLevel(log.DebugLevel)
		log.Debug("Running in verbose mode")
	}
	if *trace {
		log.SetLevel(log.TraceLevel)
		log.Debug("Running in trace mode")
	}

	log.Infof("Starting sfptpd-exporter version %s stats from %s", version, *statsFile)

	// Create a new reader from the JSONL file
	file, err := os.Open(*statsFile)
	if err != nil {
		log.Fatalf("Error opening JSONL file: %s", err)
	}
	reader := bufio.NewReader(file)

	go func() {
		for {
			scanner := bufio.NewScanner(reader)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				processLine(scanner.Text())
			}
		}
		
		if err := scanner.Err(); err != nil {
			log.Errorf("Error reading from file: %s", err)
			time.Sleep(1 * time.Second)
		} else {
			time.Sleep(200 * time.Millisecond)
		}
	}()

	// Metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	log.Infof("Starting metrics exporter on %s/metrics", *metricsListen)
	log.Fatal(http.ListenAndServe(*metricsListen, metricsMux))
}
