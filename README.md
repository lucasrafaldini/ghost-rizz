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