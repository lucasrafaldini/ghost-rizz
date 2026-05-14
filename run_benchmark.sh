#!/bin/bash
set -euo pipefail

# Configuration
COUNT=1000
IN_DIR="input_photos"
CLEAN_DIR="output_clean"
FUZZ_DIR="output_fuzz"
RESULTS_FILE="benchmark_results.txt"

# Trap errors to provide a clean exit message
trap 'echo "[-] Benchmark failed! Check $RESULTS_FILE for details."; exit 1' ERR

echo "========================================" > "$RESULTS_FILE"
echo "Ghost-Rizz Benchmark ($COUNT Images)" >> "$RESULTS_FILE"
echo "========================================" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# Function to run and time steps
run_step() {
    local label="$1"
    shift
    echo "[*] Running $label..."
    echo "--- $label ---" >> "$RESULTS_FILE"
    
    # Use /usr/bin/time for cleaner output if available (especially on macOS/Linux)
    if [[ "$OSTYPE" == "darwin"* ]] || [[ "$OSTYPE" == "linux-gnu"* ]]; then
        /usr/bin/time -p "$@" 2>&1 >/dev/null | grep real | awk '{print $2 " seconds"}' >> "$RESULTS_FILE"
    else
        # Fallback to bash built-in time
        { time "$@"; } 2>> "$RESULTS_FILE"
    fi
    echo "" >> "$RESULTS_FILE"
}

# Clean up previous runs
echo "[*] Cleaning up old directories..."
rm -rf "$IN_DIR" "$CLEAN_DIR" "$FUZZ_DIR"

# Build the latest binary
echo "[*] Building ghost-rizz binary..."
go build -o ghost-rizz ./cmd/ghost-rizz

# Run benchmark steps
run_step "GENERATE" ./ghost-rizz generate -count "$COUNT" -out "$IN_DIR"
run_step "CLEAN" ./ghost-rizz fuzz -in "$IN_DIR" -out "$CLEAN_DIR" -mode clean
run_step "FUZZ" ./ghost-rizz fuzz -in "$IN_DIR" -out "$FUZZ_DIR" -mode fuzz
run_step "REPORT (INPUTS)" ./ghost-rizz report -in "$IN_DIR"
run_step "REPORT (CLEAN)" ./ghost-rizz report -in "$CLEAN_DIR"
run_step "REPORT (FUZZ)" ./ghost-rizz report -in "$FUZZ_DIR"

echo "[+] Benchmark complete! Results saved to $RESULTS_FILE"
