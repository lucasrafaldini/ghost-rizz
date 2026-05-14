#!/bin/bash
set -euo pipefail

# Configuration
COUNT=1000
IN_DIR="input_photos"
CLEAN_DIR="output_clean"
FUZZ_DIR="output_fuzz"
RESULTS_FILE="benchmark_results.txt"

echo "========================================" > "$RESULTS_FILE"
echo "Ghost-Rizz Benchmark ($COUNT Images)" >> "$RESULTS_FILE"
echo "========================================" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# Clean up previous runs
echo "[*] Cleaning up old directories..."
rm -rf "$IN_DIR" "$CLEAN_DIR" "$FUZZ_DIR"

# Build the latest binary
echo "[*] Building ghost-rizz binary..."
go build -o ghost-rizz ./cmd/ghost-rizz

# 1. Generate
echo "[*] Running Generate..."
echo "--- GENERATE ---" >> "$RESULTS_FILE"
{ time ./ghost-rizz generate -count "$COUNT" -out "$IN_DIR" ; } 2>> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# 2. Clean
echo "[*] Running Clean..."
echo "--- CLEAN ---" >> "$RESULTS_FILE"
{ time ./ghost-rizz fuzz -in "$IN_DIR" -out "$CLEAN_DIR" -mode clean ; } 2>> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# 3. Fuzz
echo "[*] Running Fuzz..."
echo "--- FUZZ ---" >> "$RESULTS_FILE"
{ time ./ghost-rizz fuzz -in "$IN_DIR" -out "$FUZZ_DIR" -mode fuzz ; } 2>> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# 4. Report Inputs
echo "[*] Running Report on Inputs..."
echo "--- REPORT (INPUTS) ---" >> "$RESULTS_FILE"
{ time ./ghost-rizz report -in "$IN_DIR" ; } 2>> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# 5. Report Outputs
echo "[*] Running Report on Clean Outputs..."
echo "--- REPORT (CLEAN) ---" >> "$RESULTS_FILE"
{ time ./ghost-rizz report -in "$CLEAN_DIR" ; } 2>> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

echo "[*] Running Report on Fuzz Outputs..."
echo "--- REPORT (FUZZ) ---" >> "$RESULTS_FILE"
{ time ./ghost-rizz report -in "$FUZZ_DIR" ; } 2>> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

echo "[+] Benchmark complete! Results saved to $RESULTS_FILE"
