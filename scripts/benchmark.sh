#!/bin/bash
set -e

if [[ "$1" == "--help" ]]; then
  echo ""
  echo "📊 benchmark.sh: Run benchmark matrix for Keto workload simulation"
  echo ""
  echo "Usage:"
  echo "  ./scripts/benchmark.sh [--mode local|cloud]"
  echo ""
  echo "Options:"
  echo "  --mode      Select mode: 'local' (default) or 'cloud'"
  echo ""
  echo "💡 Make sure Ory Keto and CockroachDB are reachable and match the mode."
  echo "🔗 https://www.ory.sh/docs/keto/install"
  exit 0
fi

MODE="local"

while [[ $# -gt 0 ]]; do
  case $1 in
    --mode)
      MODE="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

if [[ "$MODE" == "cloud" ]]; then
  export KETO_CONFIG_PATH=./keto/config.cloud.yaml
else
  export KETO_CONFIG_PATH=./keto/config.local.yaml
fi

APP_BINARY="./crdb-ory-keto-demo"
OUTPUT_CSV="./benchmark_results.csv"
KETO_API="http://localhost:4467"

echo "🧼 Cleaning up old containers..."
docker-compose down -v --remove-orphans

echo "🚀 Starting Keto (mode: $MODE)..."
docker-compose up -d

echo "⏳ Waiting for migrations to complete..."
until [ "$(docker inspect -f '{{.State.Status}}' keto-migrate 2>/dev/null)" = "exited" ]; do
  echo "⌛ Still waiting for keto-migrate to finish..."
  sleep 2
done

EXIT_CODE=$(docker inspect -f '{{.State.ExitCode}}' keto-migrate 2>/dev/null)
if [ "$EXIT_CODE" != "0" ]; then
  echo "❌ keto-migrate failed with exit code $EXIT_CODE"
  docker logs keto-migrate
  exit 1
fi

echo "✅ Migrations completed successfully."

echo "⏳ Waiting for Keto API to respond..."
until curl -sf "$KETO_API/health/alive" > /dev/null; do
  echo "⌛ Still waiting for Keto API..."
  sleep 2
done

echo "✅ Keto API is up."

echo "🔍 Verifying Keto API is reachable..."
if ! curl -sf "$KETO_API/health/alive" > /dev/null; then
  echo "❌ Could not reach Ory Keto API at $KETO_API"
  echo "💡 Please ensure Keto is running. See: https://www.ory.sh/docs/keto/install"
  exit 1
fi

echo "📈 Starting benchmarks..."
echo "timestamp,tuple_count,concurrency,checks_per_sec,duration_sec,allowed,denied,failed" > "$OUTPUT_CSV"

matrix=(
  "1000 5 5"
  "5000 10 10"
  "10000 20 20"
)

for row in "${matrix[@]}"; do
  read -r TUPLES CONC CHECKS <<< "$row"
  echo "🔄 Benchmark: $TUPLES tuples, $CONC workers, $CHECKS checks/sec"

  LOG="bench_${TUPLES}_${CONC}.log"
  START=$(date +%s)

  $APP_BINARY \
    --tuple-count="$TUPLES" \
    --concurrency="$CONC" \
    --checks-per-second="$CHECKS" \
    --log-file="$LOG" \
    --keto-api="$KETO_API"

  END=$(date +%s)
  DURATION=$((END - START))

  ALLOWED=$(grep "📈 Allowed" "$LOG" | awk '{print $NF}' || echo "0")
  DENIED=$(grep "📉 Denied" "$LOG" | awk '{print $NF}' || echo "0")
  FAILED=$(grep "🚨 Failed writes" "$LOG" | awk '{print $NF}' || echo "0")

  echo "$(date +%Y-%m-%dT%H:%M:%S),$TUPLES,$CONC,$CHECKS,$DURATION,$ALLOWED,$DENIED,$FAILED" >> "$OUTPUT_CSV"
  echo "✅ Done: $TUPLES in ${DURATION}s → allowed=$ALLOWED denied=$DENIED failed=$FAILED"
done
