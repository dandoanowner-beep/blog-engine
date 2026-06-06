#!/usr/bin/env bash
# Handler test assertion linter.
# Flags test functions that assert an HTTP status code but never decode/assert the response body.
# Exit 0 = PASS, Exit 1 = FAIL.
set -euo pipefail
cd "$(dirname "$0")/.."

TMPFILE=$(mktemp)
trap "rm -f $TMPFILE" EXIT

find . -path "*/internal/*handler_test.go" ! -path "*/vendor/*" | sort | xargs awk '
  /^func Test/ {
    fn = $2
    gsub(/\(.*/, "", fn)
    depth = 0; has_code = 0; has_body = 0; active = 1
  }
  active {
    n = split($0, chars, "")
    for (i = 1; i <= n; i++) {
      if (chars[i] == "{") depth++
      else if (chars[i] == "}") {
        depth--
        if (depth == 0) {
          if (has_code && !has_body && !no_body_ok) {
            print FILENAME ": " fn " — status-code only, no body assertion"
          }
          active = 0; fn = ""; has_code = 0; has_body = 0; no_body_ok = 0
          break
        }
      }
    }
    if (/rec\.Code/) has_code = 1
    if (/StatusNoContent|204/) no_body_ok = 1
    if (/resp\[|\.Decode\(|json\.Unmarshal/) has_body = 1
  }
' > "$TMPFILE"

if [ -s "$TMPFILE" ]; then
  cat "$TMPFILE"
  echo ""
  COUNT=$(wc -l < "$TMPFILE")
  echo "FAIL: ${COUNT} handler test(s) assert only HTTP status code."
  echo "Every handler test MUST decode the response body and assert at least one field."
  exit 1
fi
echo "PASS: All handler tests assert response body fields."
