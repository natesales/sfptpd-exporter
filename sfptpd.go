package main

import (
	"encoding/json"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Stats struct {
	Instance    string `json:"instance"`
	Time        string `json:"time"`
	ClockMaster struct {
		Name string `json:"name"`
		Time string `json:"time,omitempty"`
	} `json:"clock-master"`
	ClockSlave struct {
		Name             string `json:"name"`
		Time             string `json:"time"`
		PrimaryInterface string `json:"primary-interface"`
	} `json:"clock-slave"`
	IsDisciplining bool          `json:"is-disciplining"`
	InSync         bool          `json:"in-sync"`
	Alarms         []interface{} `json:"alarms"`
	Stats          struct {
		Offset  float64 `json:"offset"`
		FreqAdj float64 `json:"freq-adj"`
		PTerm   float64 `json:"p-term"`
		ITerm   float64 `json:"i-term"`
	} `json:"stats"`
}

func gaugeVec(gaugeVec *prometheus.GaugeVec, instance string) prometheus.Gauge {
	return gaugeVec.With(map[string]string{"instance": instance})
}

func setBool(gauge prometheus.Gauge, value bool) {
	if value {
		gauge.Set(1)
	} else {
		gauge.Set(0)
	}
}

// parseTime6 parses a time string in the format "2022-07-29 15:52:46.121677"
func parseTime6(timeStr string) (int64, error) {
	t, err := time.Parse("2006-01-02 15:04:05.999999", timeStr)
	if err != nil {
		return -1, err
	}
	return t.UnixNano() / 1000000, nil
}

// parseTime9 parses a time string in the format "2022-07-29 15:52:46.121677000"
func parseTime9(timeStr string) (int64, error) {
	t, err := time.Parse("2006-01-02 15:04:05.999999999", timeStr)
	if err != nil {
		return -1, err
	}
	return t.UnixNano() / 1000000, nil
}

func processLine(line string) {
	var stats Stats
	err := json.Unmarshal([]byte(line), &stats)
	if err != nil {
		log.Errorf("Error parsing JSON: %s", err)
		return
	}
	log.Debugf("Parsed stats: %+v", stats)

	gaugeVec(metricLastUpdate, stats.Instance).Set(float64(time.Now().UTC().UnixNano() / 1000000))

	t, err := parseTime6(stats.Time)
	if err != nil {
		log.Warnf("Error parsing time: %s", err)
	}
	masterTime, err := parseTime9(stats.ClockMaster.Time)
	if err != nil {
		log.Warnf("Error parsing master time %s: %s", stats.ClockMaster.Time, err)
	}
	slaveTime, err := parseTime9(stats.ClockSlave.Time)
	if err != nil {
		log.Warnf("Error parsing slave time %s: %s", stats.ClockSlave.Time, err)
	}
	gaugeVec(metricTime, stats.Instance).Set(float64(t))
	metricMaster.With(map[string]string{"instance": stats.Instance, "name": stats.ClockMaster.Name}).Set(float64(masterTime))
	metricSlave.With(map[string]string{
		"instance":          stats.Instance,
		"name":              stats.ClockSlave.Name,
		"primary_interface": stats.ClockSlave.PrimaryInterface,
	}).Set(float64(slaveTime))
	setBool(gaugeVec(metricIsDisciplining, stats.Instance), stats.IsDisciplining)
	setBool(gaugeVec(metricInSync, stats.Instance), stats.InSync)
	gaugeVec(metricAlarms, stats.Instance).Set(float64(len(stats.Alarms)))
	gaugeVec(metricOffset, stats.Instance).Set(stats.Stats.Offset)
	gaugeVec(metricFreqAdj, stats.Instance).Set(stats.Stats.FreqAdj)
	gaugeVec(metricPTerm, stats.Instance).Set(stats.Stats.PTerm)
	gaugeVec(metricITerm, stats.Instance).Set(stats.Stats.ITerm)
}
