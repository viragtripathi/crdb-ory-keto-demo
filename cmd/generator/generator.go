package generator

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/config"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/keto"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/metrics"
)

func RunGenerator(dryRun bool) {
	cfg := config.AppConfig.Workload
	var wg sync.WaitGroup
	tuplesPerWorker := cfg.TupleCount / cfg.Concurrency
	rate := time.Second / time.Duration(cfg.ChecksPerSecond)

	fmt.Printf("🚧 Generating %d tuples with %d workers and max %d checks/sec...\n",
		cfg.TupleCount, cfg.Concurrency, cfg.ChecksPerSecond)

	var allowedCount, deniedCount, failedWrites int64

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ticker := time.NewTicker(rate)
			defer ticker.Stop()

			for j := 0; j < tuplesPerWorker; j++ {
				objectUUID := uuid.New()
				subjectUUID := uuid.New()

				objectName := fmt.Sprintf("doc-%d", workerID*1000+j)
				subjectName := fmt.Sprintf("user-%d", workerID)
				subjectFull := fmt.Sprintf("user:%s", subjectUUID.String())

				if !dryRun {
					err := keto.WriteTuple("documents", objectUUID.String(), "viewer", subjectFull)
					if err != nil {
						log.Printf("❌ WriteTuple failed: %v", err)
						failedWrites++
					} else {
						log.Println("📤 Tuple mirrored to Keto successfully")
					}
				}

				<-ticker.C

				allowed := false
				if !dryRun {
					allowed = keto.CheckPermission("documents", objectUUID.String(), "viewer", subjectFull)
					log.Printf("🔒 Permission check result: subject=%s, object=%s, allowed=%v", subjectName, objectName, allowed)
				}

				if allowed {
					metrics.PermissionCheckCounter.WithLabelValues("allowed").Inc()
					allowedCount++
				} else {
					metrics.PermissionCheckCounter.WithLabelValues("denied").Inc()
					deniedCount++
				}
			}
		}(i)
	}

	wg.Wait()

	log.Println("✅ Tuple generation and permission checks complete")
	log.Printf("🔢 Total tuples: %d", cfg.TupleCount)
	log.Printf("⚙️  Concurrency: %d", cfg.Concurrency)
	log.Printf("🚦 Checks/sec:  %d", cfg.ChecksPerSecond)
	log.Printf("🧪 Mode:        %s", map[bool]string{true: "DRY RUN", false: "LIVE"}[dryRun])
	log.Printf("📈 Allowed:     %d", allowedCount)
	log.Printf("📉 Denied:      %d", deniedCount)
	log.Printf("🚨 Failed writes to Keto: %d", failedWrites)

	if dryRun {
		log.Println("⚠️  Dry-run mode: No tuples were written to Keto.")
	}
}
