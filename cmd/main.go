package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/viragtripathi/crdb-ory-keto-demo/cmd/generator"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/config"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/metrics"
)

func main() {
	ketoWriteAPI := flag.String("keto-api", "http://localhost:4467", "Base URL for Keto Write API")
	concurrency := flag.Int("concurrency", 0, "Override number of concurrent workers")
	checksPerSecond := flag.Int("checks-per-second", 0, "Override checks per second")
	duration := flag.Int("duration-sec", 0, "Override duration in seconds")
    readRatio := flag.Int("read-ratio", 0, "Override read/write ratio (e.g. 100 = 100:1)")
	dryRun := flag.Bool("dry-run", false, "Simulate workload without API calls")
	workloadConfig := flag.String("workload-config", "config/config.yaml", "Path to workload config")
	logFile := flag.String("log-file", "", "Path to log output file")
	serveMetrics := flag.Bool("serve-metrics", false, "Keep Prometheus metrics endpoint alive after run")
	verbose := flag.Bool("verbose", true, "Enable verbose logging")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `
📦 crdb-ory-keto-demo: Workload simulator for Ory Keto + CockroachDB

Usage:
  ./crdb-ory-keto-demo [flags]

Options:
  -concurrency         Number of concurrent workers (overrides config file)
  -checks-per-second   Max permission checks per second (overrides config file)
  -duration-sec        Run for this many seconds (default from config file)
  -read-ratio          Read-to-write ratio (e.g. 100 means 100 reads per 1 write)
  -keto-api            Base URL for Keto Write API (default: http://localhost:4467)
  -workload-config     Path to workload config file (default: config/config.yaml)
  -log-file            Path to write logs to (default: stdout only)
  -serve-metrics       Keep Prometheus metrics endpoint alive after run (default: false)
  -dry-run             Skip actual writes and permission checks
  -help                Show this help message

🔒 This tool assumes Ory Keto is running and reachable.
📖 See install docs: https://www.ory.sh/docs/keto/install
`)
	}

	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	if err := config.LoadConfig(*workloadConfig); err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	config.AppConfig.Keto.WriteAPI = *ketoWriteAPI
	if config.AppConfig.Keto.ReadAPI == "" {
		config.AppConfig.Keto.ReadAPI = strings.Replace(config.AppConfig.Keto.WriteAPI, ":4467", ":4466", 1)
	}

	if *concurrency > 0 {
		config.AppConfig.Workload.Concurrency = *concurrency
	}
	if *checksPerSecond > 0 {
		config.AppConfig.Workload.ChecksPerSecond = *checksPerSecond
	}
    if *duration > 0 {
        config.AppConfig.Workload.DurationSec = *duration
    }
	if *readRatio > 0 {
		config.AppConfig.Workload.ReadRatio = *readRatio
	}

	if *logFile != "" {
		f, err := os.Create(*logFile)
		if err != nil {
			log.Fatalf("❌ Failed to create log file: %v", err)
		}
		defer f.Close()

		if *verbose {
			log.SetOutput(io.MultiWriter(os.Stdout, f))
		} else {
			log.SetOutput(f)
		}
	} else if !*verbose {
		log.SetOutput(io.Discard)
	}

	if !*dryRun {
		healthURL := config.AppConfig.Keto.ReadAPI + "/health/alive"
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(healthURL)
		if err != nil || resp.StatusCode != 200 {
			log.Fatalf(`❌ Unable to reach Ory Keto at %s.

Make sure Ory Keto is running and reachable.
Refer to: https://www.ory.sh/docs/keto/install

Details:
- Error: %v
- HTTP Status: %v
`, config.AppConfig.Keto.ReadAPI, err, resp.StatusCode)
		}
	}

	metrics.Init()
	generator.RunGenerator(*dryRun)

	if *serveMetrics {
		fmt.Println("📊 Prometheus metrics available at http://localhost:2112/metrics")
		fmt.Println("🔁 Waiting indefinitely for Prometheus to scrape. Ctrl+C to exit.")
		select {}
	}
}
