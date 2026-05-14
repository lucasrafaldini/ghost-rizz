# ghost-rizz

A massive, highly-concurrent EXIF metadata cleaner and fuzzer written in pure Go. 

`ghost-rizz` is designed to process huge datasets of JPEG images incredibly fast by taking advantage of Go routines. It provides features to generate dummy datasets for testing, strip all metadata cleanly, or fuzz metadata with randomized tags to test metadata ingestion pipelines.

## Benchmark

The following benchmark demonstrates the tool's concurrent performance when processing **1000 JPEG images** on a standard machine.

| Operation | Description | Time (1000 images) |
| :--- | :--- | :--- |
| **`generate`** | Creates 1000 procedural images from scratch, injecting them with rich EXIF metadata (Make, Model, Software, GPS, etc.) | **~3.023s** |
| **`clean`** | Parses 1000 JPEGs concurrently and completely strips their EXIF segments | **~0.188s** |
| **`fuzz`** | Parses 1000 JPEGs concurrently and completely randomizes their existing EXIF metadata | **~0.664s** |
| **`report` (inputs)** | Concurrently parses 1000 JPEGs to extract detailed EXIF tags and saves to a CSV | **~0.883s** |
| **`report` (outputs)** | Concurrently parses 2000 JPEGs (1000 clean, 1000 fuzzed) to extract EXIF data to a CSV | **~0.866s** |

> **Note:** Due to the concurrent architecture using Go's `sync.WaitGroup`, the processing scales exceedingly well. Stripping the metadata of 1000 images takes roughly 200 milliseconds, meaning the tool can confidently handle hundreds of thousands of images in mere seconds (primarily bottlenecked by disk I/O, not CPU).

## Commands

The CLI comes with three primary subcommands:

### 1. Generate
Creates a specified number of procedural images (with unique colors) injected with realistic, hardcoded EXIF metadata.
```bash
go run ./cmd/ghost-rizz generate -count 1000 -out ./input_photos
```

### 2. Fuzz
Processes images from an input folder concurrently using `goroutines`. Supports two modes:
- **`clean`**: Completely strips the APP1 (EXIF) segment from the image.
- **`fuzz`**: Preserves the EXIF structure but randomizes the values of standard tags (Make, Model, Software, DateTime, ExposureTime, GPS coordinates).

Output files are automatically suffixed with `_clean` or `_fuzz` and placed in the target directory.
```bash
go run ./cmd/ghost-rizz fuzz -in ./input_photos -out ./output_photos -mode fuzz
go run ./cmd/ghost-rizz fuzz -in ./input_photos -out ./output_photos -mode clean
```

### 3. Report
Scans an entire directory concurrently and generates a highly detailed CSV file (`report.csv`) directly inside that directory, extracting all relevant EXIF tags for auditing "before-and-after" states.
```bash
go run ./cmd/ghost-rizz report -in ./output_photos
```

## Testing & Coverage

This project features a comprehensive unit testing suite achieving **93.1%+ coverage**, heavily testing goroutine synchronization, parsing logic, and the command-line interface.

To run the tests locally:
```bash
go test ./... -v
```

To view the coverage report:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

A **GitHub Actions** workflow (`.github/workflows/ci.yml`) is included and runs on every push and pull request. It enforces:
1. Native formatting (`gofmt`).
2. Linting (`golangci-lint`).
3. Strict code coverage (automatically failing the CI if coverage falls below **90.0%**).

## Automation & Benchmarking

The project includes an automation script (`run_benchmark.sh`) which runs the complete lifecycle of the tool: compiling the binary, generating 1000 dummy images, fuzzing them, cleaning them, and generating CSV reports. It captures the execution `time` for each step and saves the results to `benchmark_results.txt`.

To execute the integration test and benchmark:
```bash
./run_benchmark.sh
```