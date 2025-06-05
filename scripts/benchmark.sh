#!/bin/bash
set -e

# Configurable matrix
DATABASE_URL="${DATABASE_URL:-postgresql://root@localhost:26257/defaultdb?sslmode=disable}"
APP_BINARY="./crdb-ory-keto-demo"
OUTPUT_CSV="./benchmark_results.csv"

# Test matrix (tuples, concurrency, checks/sec)
matrix=(
  "1000 5 5"
  "5000 10 10"
  "10000 20 20"
  "25000 50 25"
)

# Write header
echo "timestamp,db_type,tuple_count,concurrency,checks_per_sec,duration_sec,allowed,denied,failed" > "$OUTPUT_CSV"

for row in "${matrix[@]}"; do
  read -r TUPLES CONC CHECKS <<< "$row"
  echo "🔄 Running benchmark: $TUPLES tuples, $CONC workers, $CHECKS checks/sec"

  LOG="bench_${TUPLES}_${CONC}.log"
  START=$(date +%s)

  DATABASE_URL="$DATABASE_URL" $APP_BINARY \
    --tuple-count="$TUPLES" \
    --concurrency="$CONC" \
    --checks-per-second="$CHECKS" \
    --log-file="$LOG"

  END=$(date +%s)
  DURATION=$((END - START))

# Ensure logs are fully written
sleep 1  # short wait for write flush

# Extract summary
  ALLOWED=$(grep "📈 Allowed" "$LOG" | awk '{print $NF}' || echo "0")
  DENIED=$(grep "📉 Denied" "$LOG" | awk '{print $NF}' || echo "0")
  FAILED=$(grep "🚨 Failed inserts" "$LOG" | awk '{print $NF}' || echo "0")

  ALLOWED=${ALLOWED:-0}
  DENIED=${DENIED:-0}
  FAILED=${FAILED:-0}


  echo "$(date +%Y-%m-%dT%H:%M:%S),CockroachDB,$TUPLES,$CONC,$CHECKS,$DURATION,$ALLOWED,$DENIED,$FAILED" >> "$OUTPUT_CSV"
  echo "✅ Done: $TUPLES in ${DURATION}s → allowed=$ALLOWED denied=$DENIED"
done

