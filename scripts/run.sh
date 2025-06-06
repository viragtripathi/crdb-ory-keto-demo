#!/bin/bash
set -e

if [[ "$1" == "--help" ]]; then
  echo ""
  echo "📦 run.sh: Start Keto and run a sample workload"
  echo ""
  echo "Usage:"
  echo "  ./scripts/run.sh [--mode local|cloud]"
  echo ""
  echo "Options:"
  echo "  --mode      Select mode: 'local' (default) or 'cloud'"
  echo ""
  echo "🛠 Starts Docker containers for Keto and Prometheus"
  echo "   Then runs the workload simulator with default settings"
  echo ""
  echo "💡 Make sure CockroachDB is reachable and the config matches the mode"
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
until curl -sf http://localhost:4466/health/alive > /dev/null; do
  echo "⌛ Still waiting for Keto API..."
  sleep 2
done

echo "✅ Keto API is up."

echo "🔍 Verifying Keto API is reachable..."
if ! curl -sf http://localhost:4467/health/alive > /dev/null; then
  echo "❌ Could not reach Ory Keto API at http://localhost:4467"
  echo "💡 Please ensure Keto is running. See: https://www.ory.sh/docs/keto/install"
  exit 1
fi

echo "🔥 Running workload simulator..."
./crdb-ory-keto-demo \
  --tuple-count=1000 \
  --concurrency=10 \
  --checks-per-second=10 \
  --keto-api=http://localhost:4467 \
  --log-file=run.log
