package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	recentLogCountVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dockerx",
		Name:      "container_log_count_last_12_hours",
		Help:      "Number of log entries for htis container in the last 12 hours",
	}, []string{"container"})
)

func main() {
	http.Handle("/metrics", promhttp.Handler())
	go containerLogs()

	log.Printf("Starting server at port 2112\n")
	if err := http.ListenAndServe(":2112", nil); err != nil {
		log.Printf("error: %v", err)
	}
}

func containerLogs() {
	for {
		cli, err := client.NewClientWithOpts(client.FromEnv)
		defer cli.Close()
		if err != nil {
			panic(err)
		}

		logOpts := types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Since:      "12h", // new logs in the last 12 hours
		}

		containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
		for _, container := range containers {
			logs, err := cli.ContainerLogs(context.Background(), container.ID, logOpts)
			if err != nil {
				log.Fatal(err)
			}
			logEntryCount, err := countLogLines(logs)
			if err != nil {
				log.Println("error counting logs: ", err)
			}
			recentLogCountVec.WithLabelValues(container.Names[0]).Set(logEntryCount)
		}
		time.Sleep(3 * time.Second)
	}
}

func countLogLines(logs io.ReadCloser) (float64, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, logs)
	if err != nil {
		return 0, err
	}
	logData := buf.Bytes()
	newLine := []byte{'\n'}
	lines := bytes.Count(logData, newLine)
	if len(logData) > 0 && !bytes.HasSuffix(logData, newLine) {
		lines++
	}
	return float64(lines), nil
}
