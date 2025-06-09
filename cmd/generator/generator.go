package generator

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/config"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/keto"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/metrics"
)

type tuple struct {
	Subject string
	Object  string
}

func RunGenerator(dryRun bool) {
	cfg := config.AppConfig.Workload
	duration := time.Duration(cfg.DurationSec) * time.Second
	endTime := time.Now().Add(duration)

	writeWorkers := 1
	readWorkers := cfg.ReadRatio
	totalWorkers := writeWorkers + readWorkers

	var wg sync.WaitGroup
	tupleChan := make(chan tuple, 10000)

	var allowedCount, deniedCount, failedWrites, readCount, writeCount int64

	log.Printf("🚧 Load generation for %v with %d total workers (%d writers, %d readers)...",
		duration, totalWorkers, writeWorkers, readWorkers)

	// Phase 1: Start write worker(s)
	for i := 0; i < writeWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for time.Now().Before(endTime) {
				objectID := uuid.New().String()
				subjectID := uuid.New().String()
				subjectFull := "user:" + subjectID

				if !dryRun {
					err := keto.WriteTuple("documents", objectID, "viewer", subjectFull)
					if err != nil {
						log.Printf("❌ WriteTuple failed: %v", err)
						failedWrites++
					} else {
						// Push the same tuple read_ratio times
						for j := 0; j < cfg.ReadRatio; j++ {
							tupleChan <- tuple{Subject: subjectFull, Object: objectID}
						}
						writeCount++
					}
				}
			}
		}(i)
	}

	// Phase 2: Start read workers
	for i := 0; i < readWorkers; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for time.Now().Before(endTime) {
				select {
				case t := <-tupleChan:
					allowed := false
					if !dryRun {
						allowed = keto.CheckPermission("documents", t.Object, "viewer", t.Subject)
						log.Printf("🔒 Permission check result: subject=%s, object=%s, allowed=%v", t.Subject, t.Object, allowed)
					}

					if allowed {
						metrics.PermissionCheckCounter.WithLabelValues("allowed").Inc()
						allowedCount++
					} else {
						metrics.PermissionCheckCounter.WithLabelValues("denied").Inc()
						deniedCount++
					}
					readCount++
				default:
					time.Sleep(5 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()

	log.Println("✅ Load generation and permission checks complete")
	log.Printf("⏱️ Duration: %v", duration)
	log.Printf("⚙️  Concurrency: %d", totalWorkers)
	log.Printf("🚦 Checks/sec:  %d", cfg.ChecksPerSecond)
	log.Printf("🧪 Mode:        %s", map[bool]string{true: "DRY RUN", false: "LIVE"}[dryRun])
	log.Printf("📈 Allowed:     %d", allowedCount)
	log.Printf("📉 Denied:      %d", deniedCount)
	log.Printf("📤 Writes:      %d", writeCount)
	log.Printf("👁️  Reads:       %d", readCount)
	if writeCount > 0 {
		log.Printf("📊 Read/Write ratio: %.1f:1", float64(readCount)/float64(writeCount))
	}
	log.Printf("🚨 Failed writes to Keto: %d", failedWrites)

	if dryRun {
		log.Println("⚠️  Dry-run mode: No tuples were written to Keto.")
	}
}
