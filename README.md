# crdb-ory-keto-demo

A workload simulator and benchmarking tool for evaluating [Ory Keto](https://www.ory.sh/docs/keto) with [CockroachDB](https://www.cockroachlabs.com/).

It simulates permission tuple insertions and checks at scale — useful for load testing and benchmarking on local or CockroachDB Cloud.

---

## 🚀 Ways to Run

### 🟢 1. Quick Demo with Docker

This mode runs Keto (with 3 replicas) + Prometheus + HAProxy in Docker. You can connect to CockroachDB **locally** or use **CockroachDB Cloud**.

```bash
./scripts/run.sh --mode local       # for local CockroachDB
./scripts/run.sh --mode cloud       # for CockroachDB Cloud
````

It will:

* Start 3 Keto containers behind HAProxy
* Run `keto-migrate` once
* Wait for health checks
* Launch the workload simulator

---

### 📊 2. Benchmarking Mode

Runs a full matrix of duration, concurrency, checks/sec, and read/write ratios:

```bash
./scripts/benchmark.sh --mode local
./scripts/benchmark.sh --mode cloud
```

Results are saved to:

```
benchmark_results.csv
```

Sample row:

```
timestamp,duration_sec,concurrency,checks_per_sec,read_ratio,allowed,denied,writes,reads,failed
2025-06-09T18:00:00Z,30,10,1000,100,34900,0,350,34900,0
```

---

### 🛠️ 3. Manual Binary Run

```bash
./crdb-ory-keto-demo \
  --duration-sec=60 \
  --concurrency=10 \
  --checks-per-second=1000 \
  --read-ratio=100 \
  --workload-config=config/stress.yaml \
  --keto-api=http://localhost:4467 \
  --log-file=run.log
```

---

## ⚙️ Configuration Modes

Two config files let you switch between environments:

* `keto/config.local.yaml` → for Docker with `host.docker.internal`
* `keto/config.cloud.yaml` → for Cockroach Cloud

Your Docker Compose respects this:

```yaml
volumes:
  - ${KETO_CONFIG_PATH:-./keto/config.local.yaml}:/config/config.yaml
```

To switch modes:

```bash
KETO_CONFIG_PATH=./keto/config.cloud.yaml ./scripts/run.sh --mode cloud
```

Or:

```bash
KETO_CONFIG_PATH=./keto/config.local.yaml ./scripts/run.sh --mode local
```

---

## 📁 Workload Config Profiles

You can define profiles like:

```yaml
keto:
  write_api: "http://localhost:4467"
  read_api: "http://localhost:4466"

workload:
  concurrency: 10
  checks_per_second: 1000
  read_ratio: 100
  duration_sec: 60
```

Run with:

```bash
./crdb-ory-keto-demo --workload-config=config/stress.yaml
```

---

## 📊 Understanding `read_ratio`

This option controls **how many reads per write**:

```yaml
read_ratio: 100
```

Means: for every 1 write, the workload will perform approximately 100 permission checks.

This simulates **real-world workloads**, where reads vastly outnumber writes.

Results include detailed breakdowns:

```
📤 Writes:      345
👁️  Reads:       34396
📊 Read/Write ratio: 99.7:1
```

---

## 🔁 Scaled Keto + Load Balancing (HAProxy)

The setup includes:

* 3 Keto containers: `keto-1`, `keto-2`, `keto-3`
* HAProxy fronting both APIs
* Used by both `run.sh` and `benchmark.sh`

### 📡 API Endpoints

| Purpose   | Endpoint                |
|-----------|-------------------------|
| Read API  | `http://localhost:4466` |
| Write API | `http://localhost:4467` |

---

### 🔍 Manual Verification of Load Distribution

```bash
docker logs --tail=3 keto-1 && echo "---"
docker logs --tail=3 keto-2 && echo "---"
docker logs --tail=3 keto-3
```

Look for:

```
method:PUT path:/admin/relation-tuples
method:POST path:/relation-tuples/check
...
```

You should see requests across all nodes.

---

## 🧪 Debugging & Troubleshooting

### ✅ Confirm APIs:

```bash
curl -s http://localhost:4466/health/alive
curl -s http://localhost:4467/health/alive
```

### ✅ See Final Config:

```bash
docker exec -it keto-1 cat /config/config.yaml
```

### ✅ Run Keto Standalone:

```bash
docker run --rm -v "$(pwd)/keto:/config" oryd/keto:v0.14.0 serve --config /config/config.yaml
```

### ✅ Show Migration Logs:

```bash
docker logs keto-migrate
```

### ✅ Manual Tuple + Check:

```bash
curl -i -X PUT http://localhost:4467/admin/relation-tuples \
  -H "Content-Type: application/json" \
  -d '{"namespace":"documents","object":"doc-123","relation":"viewer","subject_id":"user:alice"}'

curl -s -X POST http://localhost:4466/relation-tuples/check \
  -H "Content-Type: application/json" \
  -d '{"namespace":"documents","object":"doc-123","relation":"viewer","subject_id":"user:alice"}'
```

---

## 📦 Build

To build the binary locally:

```bash
make build
```

---

## ❓ Why Use the Keto API?

This simulator doesn't just benchmark the DB — it mimics **real application behavior** by calling Ory Keto’s HTTP APIs.

---

### ✅ 1. Realistic Workload

Inserts tuples via:

```http
PUT /admin/relation-tuples
```

---

### ✅ 2. Authorization Checks

Every inserted tuple is checked with:

```http
POST /relation-tuples/check
```

---

### ✅ 3. API Load Is the Benchmark

This simulates:

* Permission graph resolution
* Query path performance
* Internal caching/indexing behavior

---

### 🚫 Why Not Use Direct DB Writes?

Because:

* It skips validation/indexing
* Fails to reflect production reality
* Produces misleading benchmarks

---

## ✅ TL;DR

> Benchmark the actual access control API, not just the database behind it.

---

## 📖 References

* [Ory Keto Docs](https://www.ory.sh/docs/keto)
* [CockroachDB Docs](https://www.cockroachlabs.com/docs/)
