#!/usr/bin/env bash
# Business-logic coverage gate.
# Measures handler + service + JWT coverage, excluding SQL/cloud repository adapters.
# Exit 0 = PASS (>=80%), Exit 1 = FAIL.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "=== Running tests with coverage ==="
go test ./internal/... -coverprofile=coverage_biz.out -covermode=atomic -count=1

echo ""
echo "=== Business-Logic Coverage (handlers + services, excl. repository adapters) ==="
NONREPO=$(go tool cover -func=coverage_biz.out | grep -Ev "repository|r2\.go|total:")
echo "$NONREPO"

AVG=$(echo "$NONREPO" | awk '{gsub(/%/,"",$3); sum+=$3; n++} END {if(n>0) printf "%.1f", sum/n; else print "0"}')
TOTAL=$(go tool cover -func=coverage_biz.out | grep total | awk '{print $3}')

echo ""
echo "=== Summary ==="
echo "Business-logic avg : ${AVG}%  (gate: >=80%)"
echo "Total internal     : ${TOTAL} (informational)"
echo ""

if awk "BEGIN {exit !($AVG >= 80)}"; then
  echo "PASS: ${AVG}% meets the 80% business-logic coverage gate."
else
  echo "FAIL: ${AVG}% is below the 80% minimum."
  echo ""
  echo "Under-covered functions (< 80%):"
  echo "$NONREPO" | awk '{gsub(/%/,"",$3); if($3+0 < 80) print "  " $0}' | head -30
  exit 1
fi
