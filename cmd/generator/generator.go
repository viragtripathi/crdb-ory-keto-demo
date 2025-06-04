package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/config"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/db"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/keto"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/metrics"
)

type KetoTupleWrite struct {
	Namespace string `json:"namespace"`
	Object    string `json:"object"`
	Relation  string `json:"relation"`
	SubjectID string `json:"subject_id"`
}

func RunGenerator(dryRun bool) {
	cfg := config.AppConfig.Workload
	var wg sync.WaitGroup
	tuplesPerWorker := cfg.TupleCount / cfg.Concurrency
	rate := time.Second / time.Duration(cfg.ChecksPerSecond)

	fmt.Printf("🚧 Generating %d tuples with %d workers and max %d checks/sec...\n",
		cfg.TupleCount, cfg.Concurrency, cfg.ChecksPerSecond)

	var allowedCount, deniedCount, failedInserts int64

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ticker := time.NewTicker(rate)
			defer ticker.Stop()

			for j := 0; j < tuplesPerWorker; j++ {
				shardID := uuid.New()
				nid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
				objectUUID := uuid.New()
				subjectUUID := uuid.New()

				objectName := fmt.Sprintf("doc-%d", workerID*1000+j)
				subjectName := fmt.Sprintf("user-%d", workerID)
				subjectFull := fmt.Sprintf("user:%s", subjectUUID.String())

				if !dryRun {
					_ = db.InsertUUIDMapping(context.Background(), objectUUID, objectName)
					_ = db.InsertUUIDMapping(context.Background(), subjectUUID, subjectName)

					tuple := db.KetoTuple{
						ShardID:             shardID,
						NetworkID:           nid,
						Namespace:           "documents",
						Object:              objectUUID,
						Relation:            "viewer",
						SubjectID:           subjectUUID,
						CommitTime:          time.Now().UTC(),
						SubjectSetNamespace: nil,
						SubjectSetObject:    nil,
						SubjectSetRelation:  nil,
					}

					start := time.Now()
					err := db.InsertKetoTuple(context.Background(), tuple)
					duration := time.Since(start).Seconds()

					if err != nil {
						log.Printf("❌ Insert failed for %s -> %s: %v", subjectName, objectName, err)
						failedInserts++
						continue
					}

					log.Printf("✅ Tuple inserted into CockroachDB: %s (subject) -> %s (object)", subjectName, objectName)
					metrics.TupleInsertDuration.Observe(duration)

					ketoWrite := KetoTupleWrite{
						Namespace: "documents",
						Object:    objectUUID.String(),
						Relation:  "viewer",
						SubjectID: subjectFull,
					}

					body, _ := json.Marshal(ketoWrite)
					req, err := http.NewRequest(http.MethodPut, "http://localhost:4467/admin/relation-tuples", bytes.NewBuffer(body))
					if err != nil {
						log.Printf("❌ Failed to build PUT request to Keto: %v", err)
					} else {
						req.Header.Set("Content-Type", "application/json")
						client := &http.Client{}
						resp, err := client.Do(req)
						if err != nil || resp.StatusCode >= 300 {
							respBody, _ := io.ReadAll(resp.Body)
							log.Printf("⚠️  Failed to PUT to Keto: status=%v body=%s error=%v", resp.StatusCode, string(respBody), err)
						} else {
							log.Printf("📤 Tuple mirrored to Keto successfully")
						}
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

	fmt.Println("✅ Tuple generation and permission checks complete")
	fmt.Printf("🔢 Total tuples: %d\n", cfg.TupleCount)
	fmt.Printf("⚙️  Concurrency: %d\n", cfg.Concurrency)
	fmt.Printf("🚦 Checks/sec:  %d\n", cfg.ChecksPerSecond)
	fmt.Printf("🧪 Mode:        %s\n", map[bool]string{true: "DRY RUN", false: "LIVE"}[dryRun])
	if !dryRun {
		fmt.Printf("📈 Allowed:     %d\n", allowedCount)
		fmt.Printf("📉 Denied:      %d\n", deniedCount)
		fmt.Printf("🚨 Failed inserts: %d\n", failedInserts)
	}
}
