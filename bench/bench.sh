#!/usr/bin/env bash
# bench/bench.sh — CLI benchmark: gjq vs jq
#
# Usage:
#   ./bench/bench.sh            # run all benchmarks
#   ./bench/bench.sh --quick    # fewer iterations (10)
#   ./bench/bench.sh --chart    # generate a text bar chart
#
# Requires: go, jq, python3

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DATA_DIR="$PROJECT_DIR/tests/data"
GJQ_BIN="$PROJECT_DIR/gjq"

ITERATIONS=20
DO_CHART=false

for arg in "$@"; do
    case "$arg" in
        --quick)   ITERATIONS=10 ;;
        --chart)   DO_CHART=true ;;
        --help|-h)
            echo "Usage: $0 [--quick] [--chart]"
            echo "  --quick   Run fewer iterations (10 instead of 20)"
            echo "  --chart   Generate a text bar chart of results"
            exit 0
            ;;
    esac
done

echo "Building gjq..."
go build -o "$GJQ_BIN" "$PROJECT_DIR" 2>&1
echo ""

if ! command -v jq &>/dev/null; then
    echo "Error: jq not found. Install with: brew install jq"
    exit 1
fi

JQ_VERSION=$(jq --version 2>&1)
GJQ_VERSION=$("$GJQ_BIN" --version 2>&1 || echo "unknown")

printf "gjq  %s\n" "$GJQ_VERSION"
printf "jq   %s\n" "$JQ_VERSION"
printf "Iterations: %d\n\n" "$ITERATIONS"

time_cmd() {
    local iters=$1; shift
    local cmd_json
    cmd_json=$(printf '%s\0' "$@" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.buffer.read().rstrip(b"\x00").decode().split("\x00")))')
    python3 -c "
import subprocess, time, statistics
iters = $iters
cmd = $cmd_json
timings = []
for _ in range(iters):
    start = time.perf_counter()
    subprocess.run(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    timings.append((time.perf_counter() - start) * 1000)
timings.sort()
median = statistics.median(timings)
print(f'{median:.2f}')
"
}

# bar WIDTH PERCENT — prints a bar of WIDTH chars
bar() {
    local width=$1 percent=$2
    python3 -c "
w, p = $width, $percent
filled = int(p * w / 100)
print('█' * filled + '░' * (w - filled), end='')
"
}

# --- benchmark definitions ---

# Each entry: LABEL | FILE | GJQ_QUERY | JQ_QUERY | GJQ_EXTRA_FLAGS
declare -a BENCHES=(
    # --- nobel_prizes.json (227 KB) ---
    "prizes.category          |nobel_prizes.json|prizes[*].category                       |.prizes[].category                    |"
    "prizes.laureates.surname |nobel_prizes.json|prizes[*].laureates[*].surname           |.prizes[].laureates[].surname         |"
    "recursive firstname      |nobel_prizes.json|**.firstname                             |[.. | .firstname? // empty]           |"
    "recursive motivation (-F)|nobel_prizes.json|motivation                               |[.. | .motivation? // empty]          |-F"
    "prizes.laureates.share   |nobel_prizes.json|prizes[*].laureates[*].share             |.prizes[].laureates[].share           |"
    "case-insensitive firstname|nobel_prizes.json|**.Firstname                            |[.. | .firstname? // empty]           |-i"

    # --- simple.json (106 B) ---
    "simple: name             |simple.json      |name                                     |.name                                 |"
    "simple: name.first       |simple.json      |name.first                               |.name.first                           |"
    "simple: hobbies[0]       |simple.json      |hobbies[0]                               |.hobbies[0]                           |"
    "simple: wildcard *       |simple.json      |*                                        |[.[]]                                 |"

    # --- nested.json (200 B) ---
    "nested: users[*].name    |nested.json      |users[*].name                            |.users[].name                         |"
    "nested: deep recursive   |nested.json      |**.deep                                  |[.. | .deep? // empty]               |"

    # --- openapi_paths.json (204 B) ---
    "openapi: *.*.summary     |openapi_paths.json|paths.*.*.summary                       |.paths[][]?.summary                   |"
)

# --- run benchmarks ---

printf "%-30s | %10s | %10s | %8s\n" "Benchmark" "gjq (ms)" "jq (ms)" "Speedup"
printf '%s\n' "$(printf '%.0s-' {1..72})"

declare -a LABELS=()
declare -a GJQ_TIMES=()
declare -a JQ_TIMES=()

for entry in "${BENCHES[@]}"; do
    IFS='|' read -r label file gjq_query jq_query gjq_flags <<< "$entry"
    label=$(echo "$label" | sed 's/^ *//;s/ *$//')
    file=$(echo "$file" | sed 's/^ *//;s/ *$//')
    gjq_query=$(echo "$gjq_query" | sed 's/^ *//;s/ *$//')
    jq_query=$(echo "$jq_query" | sed 's/^ *//;s/ *$//')
    gjq_flags=$(echo "$gjq_flags" | sed 's/^ *//;s/ *$//')

    filepath="$DATA_DIR/$file"

    if [ -n "$gjq_flags" ]; then
        gjq_ms=$(time_cmd "$ITERATIONS" "$GJQ_BIN" "$gjq_flags" "$gjq_query" "$filepath")
    else
        gjq_ms=$(time_cmd "$ITERATIONS" "$GJQ_BIN" "$gjq_query" "$filepath")
    fi
    jq_ms=$(time_cmd "$ITERATIONS" jq "$jq_query" "$filepath")

    speedup=$(python3 -c "print(f'{$jq_ms / $gjq_ms:.2f}x')")

    printf "%-30s | %8.2fms | %8.2fms | %8s\n" "$label" "$gjq_ms" "$jq_ms" "$speedup"

    LABELS+=("$label")
    GJQ_TIMES+=("$gjq_ms")
    JQ_TIMES+=("$jq_ms")
done

# --- chart ---
if [ "$DO_CHART" = true ]; then
    echo ""
    echo "=== Performance Comparison (gjq vs jq) ==="
    echo ""

    max_ms=$(python3 -c "
times = [$(printf '%s,' "${GJQ_TIMES[@]}" "${JQ_TIMES[@]}" | sed 's/,$//')]
print(f'{max(times):.2f}')
")

    for i in "${!LABELS[@]}"; do
        gjq_pct=$(python3 -c "print(int(${GJQ_TIMES[$i]} * 100 / $max_ms))")
        jq_pct=$(python3 -c "print(int(${JQ_TIMES[$i]} * 100 / $max_ms))")

        printf "%-28s gjq " "${LABELS[$i]}"
        bar 40 "$gjq_pct"
        printf " %6.1fms\n" "${GJQ_TIMES[$i]}"

        printf "%-28s  jq " ""
        bar 40 "$jq_pct"
        printf " %6.1fms\n" "${JQ_TIMES[$i]}"

        echo ""
    done
fi

echo ""
echo "Done. ($ITERATIONS iterations per command, median reported)"
