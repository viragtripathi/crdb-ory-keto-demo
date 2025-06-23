package keto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/viragtripathi/crdb-ory-keto-demo/internal/config"
	"github.com/viragtripathi/crdb-ory-keto-demo/internal/metrics"
)

type CheckRequest struct {
	Namespace string `json:"namespace"`
	Object    string `json:"object"`
	Relation  string `json:"relation"`
	SubjectID string `json:"subject_id"`
}

type CheckResponse struct {
	Allowed bool `json:"allowed"`
}

type RelationTuple struct {
	Namespace string `json:"namespace"`
	Object    string `json:"object"`
	Relation  string `json:"relation"`
	SubjectID string `json:"subject_id"`
}

// Shared HTTP client with pooling and connection reuse
var (
	sharedClient *http.Client
	clientOnce   sync.Once
)

func initClient() {
	clientOnce.Do(func() {
		cfg := config.AppConfig.Workload
		log.Printf("🔧 HTTP Pool: MaxIdleConns=%d, MaxIdleConnsPerHost=%d, MaxConnsPerHost=%d\n",
        	cfg.MaxIdleConns, cfg.MaxIdleConns, cfg.MaxOpenConns)
		tr := &http.Transport{
			MaxIdleConns:        cfg.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.MaxIdleConns,
			MaxConnsPerHost:     cfg.MaxOpenConns,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
		}
		sharedClient = &http.Client{
			Transport: tr,
			Timeout:   time.Duration(cfg.RequestTimeoutSec) * time.Second,
		}
	})
}

func CheckPermission(namespace, object, relation, subjectID string) bool {
	cfg := config.AppConfig.Workload
	initClient()

	reqBody := CheckRequest{
		Namespace: namespace,
		Object:    object,
		Relation:  relation,
		SubjectID: subjectID,
	}

	url := config.AppConfig.Keto.ReadAPI + "/relation-tuples/check"
	var resp *http.Response
	var err error

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		start := time.Now()

		jsonData, marshalErr := json.Marshal(reqBody)
		if marshalErr != nil {
			log.Printf("❌ Error marshaling check request: %v\n", marshalErr)
			return false
		}

		resp, err = sharedClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
		metrics.RetryAttempts.Inc()

		if err == nil && resp != nil && resp.StatusCode == 200 {
			metrics.RetrySuccess.Inc()
			metrics.RetryDuration.Observe(time.Since(start).Seconds())
			break
		}

		if attempt < cfg.MaxRetries {
			log.Printf("🔁 Retry %d: Keto check failed (status=%v, error=%v)\n", attempt, getStatus(resp), err)
			time.Sleep(time.Duration(cfg.RetryDelayMillis) * time.Millisecond)
		}
	}

	if err != nil || resp == nil {
		log.Printf("❌ Final failure: Keto check failed after %d attempts. Error: %v\n", cfg.MaxRetries, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("⚠️ Unexpected status from Keto: %d\nResponse body: %s\n", resp.StatusCode, string(body))
		return false
	}

	var checkResp CheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&checkResp); err != nil {
		log.Printf("❌ Error decoding Keto check response: %v\n", err)
		return false
	}

	return checkResp.Allowed
}

func WriteTuple(namespace, object, relation, subjectID string) error {
	cfg := config.AppConfig.Workload
	initClient()

	tuple := RelationTuple{
		Namespace: namespace,
		Object:    object,
		Relation:  relation,
		SubjectID: subjectID,
	}

	jsonData, err := json.Marshal(tuple)
	if err != nil {
		log.Printf("failed to marshal tuple: %v", err)
        return fmt.Errorf("failed to marshal tuple: %w", err)

	}

	url := config.AppConfig.Keto.WriteAPI + "/admin/relation-tuples"
	var resp *http.Response

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		start := time.Now()

		// Create a new request and buffer on every retry
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
		if err != nil {
		    log.Printf("failed to build request: %v", err)
            return fmt.Errorf("failed to build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = sharedClient.Do(req)
		metrics.RetryAttempts.Inc()

		if err == nil && resp.StatusCode < 300 {
			metrics.RetrySuccess.Inc()
			metrics.RetryDuration.Observe(time.Since(start).Seconds())
			break
		}

		if attempt < cfg.MaxRetries {
			log.Printf("🔁 Retry %d: Keto write failed (status=%v, error=%v)\n", attempt, getStatus(resp), err)
			time.Sleep(time.Duration(cfg.RetryDelayMillis) * time.Millisecond)
		}
	}

	if err != nil || resp == nil {
		log.Printf("❌ Final failure: WriteTuple failed after %d attempts: %v", cfg.MaxRetries, err)
        return fmt.Errorf("❌ Final failure: WriteTuple failed after %d attempts: %w", cfg.MaxRetries, err)

	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("PUT failed: status=%v body=%s", resp.StatusCode, string(body))
        return fmt.Errorf("PUT failed: status=%v body=%s", resp.StatusCode, string(body))

	}

	return nil
}

func getStatus(resp *http.Response) int {
	if resp != nil {
		return resp.StatusCode
	}
	return 0
}
