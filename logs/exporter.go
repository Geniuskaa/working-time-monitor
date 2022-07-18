package logs

import (
	"bufio"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"regexp"
	"strings"
)

const namespace = "sovcom"

var (
	// Metrics
	logsSent = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "go_logs"),
		"are written by app.",
		[]string{"level", "date", "msg"}, nil,
	)
)

type Exporter struct {
	logFilePath string
}

func NewExporter(logFilePath string) *Exporter {
	return &Exporter{logFilePath: logFilePath}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- logsSent
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	logs, err := e.logsReader()
	if err != nil {
		return
	}

	for _, log := range logs {
		switch log[0] {
		case "DEBUG":
			ch <- prometheus.MustNewConstMetric(logsSent, prometheus.UntypedValue, 1, log[0], log[1], log[2])
		case "INFO":
			ch <- prometheus.MustNewConstMetric(logsSent, prometheus.UntypedValue, 2, log[0], log[1], log[2])
		case "WARN":
			ch <- prometheus.MustNewConstMetric(logsSent, prometheus.UntypedValue, 3, log[0], log[1], log[2])
		case "ERROR":
			ch <- prometheus.MustNewConstMetric(logsSent, prometheus.UntypedValue, 4, log[0], log[1], log[2])
		case "DPANIC":
			ch <- prometheus.MustNewConstMetric(logsSent, prometheus.UntypedValue, 5, log[0], log[1], log[2])
		case "PANIC":
			ch <- prometheus.MustNewConstMetric(logsSent, prometheus.UntypedValue, 6, log[0], log[1], log[2])
		case "FATAL":
			ch <- prometheus.MustNewConstMetric(logsSent, prometheus.UntypedValue, 7, log[0], log[1], log[2])
		default:
			ch <- prometheus.MustNewConstMetric(logsSent, prometheus.UntypedValue, 10, "Unsupported level", log[1], log[2])
		}
	}
}

func (e *Exporter) logsReader() ([][3]string, error) {
	file, err := os.Open(e.logFilePath)
	if err != nil {
		return nil, err
	}

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	logs := make([][3]string, 0)

	for fileScanner.Scan() {
		s := fileScanner.Text()
		arr := strings.Split(s, ",")
		tempArr := [3]string{}

		for j := 0; j < 3; j++ {
			regexedString := regexp.MustCompile(":.+").FindString(arr[j])
			t := strings.Replace(regexedString, ":", "", 1)
			tt := strings.ReplaceAll(t, "}", "")
			tempArr[j] = strings.ReplaceAll(tt, "\"", "")
		}
		logs = append(logs, tempArr)

	}

	return logs, nil
}
