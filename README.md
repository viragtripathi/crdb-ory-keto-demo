# crdb-ory-keto-demo

A workload simulator and benchmarking tool for evaluating [Ory Keto](https://www.ory.sh/docs/keto) with [CockroachDB](https://www.cockroachlabs.com/).
It simulates permission tuple insertions and checks at scale — useful for load testing and benchmarking on local or CockroachDB Cloud.

---

## 🚀 Ways to Run

### 🟢 1. Quick Demo with Docker

This mode runs Keto + Prometheus in Docker. You run CockroachDB **locally** or use Cockroach Cloud.

```bash
./scripts/run.sh --mode local       # for local CockroachDB
./scripts/run.sh --mode cloud       # for CockroachDB Cloud
```

It:

* Starts Keto and Prometheus
* Runs `keto-migrate`
* Waits for API readiness (migration + health)
* Launches the simulator binary

---

### 📊 2. Benchmarking Mode

Runs a full test matrix of tuple volumes, concurrency, and check rates:

```bash
./scripts/benchmark.sh --mode local
./scripts/benchmark.sh --mode cloud
```

Results are saved to:

```
benchmark_results.csv
```

---

### 🛠️ 3. Manual Binary Run

```bash
./crdb-ory-keto-demo   --tuple-count=1000   --concurrency=10   --checks-per-second=5   --workload-config=config/stress.yaml   --keto-api=http://localhost:4467   --log-file=run.log
```

You can override any value from the workload config using flags.

---

## ⚙️ Configuration Modes

Two config files let you switch between environments:

* `keto/config.local.yaml` → points to `host.docker.internal`
* `keto/config.cloud.yaml` → points to CockroachDB Cloud

Your `docker-compose.yml` is wired to:

```yaml
volumes:
  - ${KETO_CONFIG_PATH:-./keto/config.local.yaml}:/config/config.yaml
```

So switching is as easy as:

```bash
KETO_CONFIG_PATH=./keto/config.cloud.yaml ./scripts/run.sh --mode cloud
```

Or:

```bash
KETO_CONFIG_PATH=./keto/config.local.yaml ./scripts/run.sh --mode local
```

---

## 📁 Workload Config Profiles

You can define workload profiles like:

```yaml
# config/small.yaml
keto:
  write_api: "http://localhost:4467"
  read_api: "http://localhost:4466"
workload:
  tuple_count: 500
  concurrency: 5
  checks_per_second: 5
```

Run with:

```bash
./crdb-ory-keto-demo --workload-config=config/small.yaml
```

---

## 🔁 Keto API Scaling with Built-in Load Balancer

This project supports horizontally scaling **Ory Keto** using Docker and **HAProxy**.

### ✅ What’s Included

* 3 Keto containers: `keto-1`, `keto-2`, `keto-3`
* HAProxy load balancing across all instances
* Built-in integration in both:

    * `./scripts/run.sh`
    * `./scripts/benchmark.sh`

### 📡 API Endpoints via HAProxy

| Purpose   | Endpoint                |
|-----------|-------------------------|
| Read API  | `http://localhost:4466` |
| Write API | `http://localhost:4467` |

These are used by the workload simulator behind the scenes.

---

### 🔍 Manual Verification (Optional)

To confirm load is distributed across all 3 nodes:

```bash
docker logs --tail=3 keto-1 && echo "---" && \
docker logs --tail=3 keto-2 && echo "---" && \
docker logs --tail=3 keto-3
```

You should see traffic like:

```
method:PUT path:/admin/relation-tuples
method:POST path:/relation-tuples/check
...
```

This confirms that the simulator is routing traffic evenly through HAProxy.

---

## 🧪 Debugging & Troubleshooting

### ✅ Confirm Keto Write API is Healthy:

```bash
curl -s http://localhost:4467/health/alive
```

### ✅ Confirm Keto Read API is Healthy:

```bash
curl -s http://localhost:4466/health/alive
```

### ✅ Dump final config mounted in Docker:

```bash
docker exec -it keto cat /config/config.yaml
```

### ✅ Run Keto manually to test config:

```bash
docker run --rm -v "$(pwd)/keto:/config" oryd/keto:v0.14.0 serve --config /config/config.yaml
```

### ✅ Show all migrations:

```bash
docker logs keto-migrate
```

### ✅ Full verify check for write+read API:

```bash
curl -i -X PUT http://localhost:4467/admin/relation-tuples \
  -H "Content-Type: application/json" \
  -d '{"namespace":"documents","object":"doc-123","relation":"viewer","subject_id":"user:alice"}'

curl -s -X POST http://localhost:4466/relation-tuples/check \
  -H "Content-Type: application/json" \
  -d '{"namespace":"documents","object":"doc-123","relation":"viewer","subject_id":"user:alice"}'
```

---

### 🛑 Unreachable Keto Detection

If Ory Keto is not reachable, you will see:

```
❌ Failed to reach Ory Keto at http://localhost:4467
- Error: dial tcp [::1]:4467: connect: connection refused
```

The tool will exit cleanly.

---

## 📦 Build

To build the binary locally:

```bash
make build
```

---

## ❓ Why Use the Ory Keto API in This Project?

This simulator doesn't just benchmark CockroachDB — it mimics **real-world access control workflows** by calling Ory Keto's REST APIs.

Here’s why:

### ✅ 1. Realistic Tuple Ingestion

Instead of just writing to the database, the simulator **mirrors every relation tuple** to:

```http
PUT /admin/relation-tuples
```

This mimics how production apps interact with Keto.

---

### ✅ 2. Access Control Validation

After inserting a tuple, the simulator performs a permission check via:

```http
POST /relation-tuples/check
```

This:

* Validates tuple registration
* Measures API response under load
* Tests authorization correctness

---

### ✅ 3. API Load Is the Real Benchmark

By using the API, the simulator:

* Benchmarks **Keto’s REST pipeline**, not just the DB
* Simulates real-world usage patterns
* Surfaces rate limits or latency issues

---

### 🚫 Why Not Write Directly to the DB?

Because:

* You’d bypass consistency & validation
* Keto wouldn’t register or index those tuples
* Benchmarks would be meaningless

---

## ✅ TL;DR

> Using the API is essential to simulate real-world usage, validate authorization correctness, and benchmark the actual control path — not just storage speed.

---

## 📖 References

* [Ory Keto Install Guide](https://www.ory.sh/docs/keto/install)
* [CockroachDB Start Guide](https://www.cockroachlabs.com/docs/stable/start-a-local-cluster.html)
