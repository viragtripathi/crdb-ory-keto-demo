package main

import (
	"flag"
	"log"
	"os"
	"io"

	"github.com/viragtripathi/crdb-ory-keto-demo/internal/config"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/db"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/metrics"
	"github.com/viragtripathi/crdb-ory-keto-demo/cmd/generator"
)

func main() {
    initSchema := flag.Bool("init-schema", false, "Create required tables in the database and exit")
	tupleCount := flag.Int("tuple-count", 0, "Override number of tuples to insert")
	concurrency := flag.Int("concurrency", 0, "Override number of concurrent workers")
	checksPerSecond := flag.Int("checks-per-second", 0, "Override checks per second")
	dryRun := flag.Bool("dry-run", false, "Simulate workload without DB or API calls")
	logFile := flag.String("log-file", "", "Path to log output file")
	verbose := flag.Bool("verbose", true, "Enable verbose logging (default: true)")

	flag.Parse()

    if *logFile != "" {
        f, err := os.Create(*logFile)
        if err != nil {
            log.Fatalf("❌ Failed to create log file: %v", err)
        }
        defer f.Close()

        if *verbose {
            mw := io.MultiWriter(os.Stdout, f)
            log.SetOutput(mw)
            os.Stdout = f
            os.Stderr = f
        } else {
            log.SetOutput(io.Discard)
        }
    } else if !*verbose {
        log.SetOutput(io.Discard)
    }

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("❌ Config error: %v", err)
	}

	if *tupleCount > 0 {
		config.AppConfig.Workload.TupleCount = *tupleCount
	}
	if *concurrency > 0 {
		config.AppConfig.Workload.Concurrency = *concurrency
	}
	if *checksPerSecond > 0 {
		config.AppConfig.Workload.ChecksPerSecond = *checksPerSecond
	}

    connStr := os.Getenv("DATABASE_URL")
    if connStr == "" {
        connStr = config.AppConfig.Database.URL
    }

    if !*dryRun || *initSchema {
    	if err := db.Connect(connStr); err != nil {
    		log.Fatalf("❌ DB connection failed: %v", err)
    	}
    	defer db.Close()
    }

    if *initSchema {
    	if err := db.InitSchema(); err != nil {
    		log.Fatalf("❌ Failed to initialize schema: %v", err)
    	}
    	log.Println("✅ Schema initialized successfully.")
    	return
    }

	metrics.Init()
	generator.RunGenerator(*dryRun)
}
