Awesome — let’s start with **Step 1**: your comprehensive, production-grade `README.md`.

---

## 📘 `README.md`

```markdown
# crdb-ory-keto-demo

A workload simulator for [Ory Keto](https://www.ory.sh/keto) + [CockroachDB](https://www.cockroachlabs.com/), designed for performance benchmarking, access control validation, and large-scale POC demos.

This tool inserts relation tuples into CockroachDB, mirrors them to Ory Keto’s HTTP API, and performs real-time permission checks.

---

## 🚀 Features

- ✅ High-throughput **tuple ingestion** into CockroachDB
- ✅ **Keto HTTP API** integration (`PUT /admin/relation-tuples` and `POST /relation-tuples/check`)
- ✅ **Concurrency & throttling** controls (via CLI)
- ✅ **Prometheus metrics**: insert latency, check counts
- ✅ **Dry-run** mode: skip external calls, useful for perf/load planning
- ✅ **Log to file** for analysis/replay/debug
- ✅ **Flexible config** and CLI overrides
- ✅ Works with:
  - CockroachDB (local, on-prem, or cloud)
  - Ory Keto (v0.14.0+)
  - Docker or standalone installs

---

## 🧰 Folder Structure

```

crdb-ory-keto-demo/
├── cmd/
│   └── main.go
├── internal/
│   ├── config/           # YAML loader
│   ├── db/               # CockroachDB insert logic
│   ├── keto/             # HTTP client for checks
│   └── metrics/          # Prometheus integration
├── scripts/
│   └── build\_proto\_and\_run.sh  # full workflow script
├── config/
│   └── config.yaml
├── docker-compose.yml
├── run.sh
└── README.md

````

---

## ⚙️ Configuration

### 📁 `config/config.yaml`

```yaml
database:
  url: "postgresql://root@localhost:26257/defaultdb?sslmode=disable"

keto:
  base_url: "http://localhost:4466"

workload:
  tuple_count: 1000
  concurrency: 10
  checks_per_second: 5
````

---

## 🛠 CLI Flags

Override any config value via command-line:

| Flag                  | Description                          |
| --------------------- | ------------------------------------ |
| `--tuple-count`       | Number of tuples to insert           |
| `--concurrency`       | Number of parallel goroutines        |
| `--checks-per-second` | Max checks per second per worker     |
| `--dry-run`           | Skip DB/API calls, log only          |
| `--log-file=out.log`  | Write logs to file instead of stdout |

---

## 🧪 Examples

### Dry run 5k tuples with 20 workers:

```bash
go run cmd/main.go --tuple-count=5000 --concurrency=20 --dry-run
```

### Full test with cloud CockroachDB:

```bash
DATABASE_URL="postgresql://root@<host>:26257/defaultdb?sslmode=require" \
go run cmd/main.go --tuple-count=10000 --concurrency=50 --checks-per-second=10
```

### Log to file:

```bash
go run cmd/main.go --log-file run.log
```

---

## 🔍 Validation

### Logging output (live or in file):

```
✅ Tuple inserted into CockroachDB: user-1 -> doc-1001
📤 Tuple mirrored to Keto successfully
🔒 Permission check result: allowed=true
```

### Summary report:

```
✅ Tuple generation and permission checks complete
🔢 Total tuples: 1000
⚙️  Concurrency: 10
🚦 Checks/sec:  5
🧪 Mode:        LIVE
📈 Allowed:     999
📉 Denied:      0
🚨 Failed inserts: 0
```

---

## 📊 Prometheus Metrics

Available at [http://localhost:2112/metrics](http://localhost:2112/metrics)

* `tuple_insert_duration_seconds`
* `permission_check_total{result="allowed|denied"}`

Add to Grafana for real-time dashboards.

---

## 🧪 Run Everything with a Script

```bash
./scripts/run.sh
```

This script:

* Starts Docker containers
* Applies DB schema
* Runs the simulator

---

## 🧼 Clean Shutdown

```bash
docker compose down -v
rm run.log
```

---

Yes — it’s absolutely worth updating the README to include the binary build and usage instructions.

Providing precompiled binaries (or simple build instructions) is standard for open-source CLI tools. It:

* ✅ Makes onboarding faster for new users
* ✅ Avoids Go dependency setup
* ✅ Encourages adoption by developers and SREs in CI/CD or automation scripts

---

## ✅ Updated `README.md` Section to Add

Append this to your README after the **“Run Everything with a Script”** section:

---

### 🛠️ Build as a CLI Binary

You can build and distribute the simulator as a standalone CLI:

```bash
make build
```

The compiled binary `crdb-ory-keto-demo` can then be run directly:

```bash
./crdb-ory-keto-demo --help
```

### Cross-platform builds:

```bash
make build-linux     # For Linux x86_64
make build-mac       # For macOS ARM64 (Apple Silicon)
make build-windows   # For Windows x86_64
```

### Clean up all binaries:

```bash
make clean
```

---

### 🧪 Example usage with binary

```bash
./crdb-ory-keto-demo \
  --tuple-count=10000 \
  --concurrency=50 \
  --checks-per-second=20 \
  --log-file=sim-run.log
```

The binary will:

* Connect to the database using `DATABASE_URL` or config fallback
* Mirror tuples to Keto via HTTP
* Perform permission checks with real responses
* Output summary stats and logs

---
