#!/usr/bin/env bash
# Tests that `make check` stops the Compose stack when `make up` fails.
# A fake compose command fails on `up` and records each call.
set -euo pipefail
cd "$(dirname "$0")/.."

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

cat > "$tmp/compose" <<'FAKE'
#!/usr/bin/env bash
echo "$1" >> "$(dirname "$0")/calls"
[ "$1" != "up" ]
FAKE
chmod +x "$tmp/compose"

# -o skips lint and test, so that only the stack steps run.
if make check -o lint -o test COMPOSE="$tmp/compose" > "$tmp/output" 2>&1; then
  echo "FAIL: make check succeeded, but make up failed"
  exit 1
fi

calls=$(tr '\n' ' ' < "$tmp/calls")
if [ "$calls" != "up logs down " ]; then
  echo "FAIL: compose calls were '$calls', want 'up logs down '"
  cat "$tmp/output"
  exit 1
fi
echo "ok: make check runs logs and down when make up fails"
