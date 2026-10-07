<div align="center">

# SnapCache

**An experimental generic cache implementation and eviction-policy playground in Go**

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat-square&logo=go&logoColor=white)
![Generics](https://img.shields.io/badge/API-Generic%20K%2FV-5D87BF?style=flat-square)
![Policies](https://img.shields.io/badge/Policies-SnapCache%20%7C%20LRU%20%7C%20S3--FIFO%20%7C%20SIEVE-444444?style=flat-square)

</div>

## Overview

SnapCache is a small research-oriented repository for experimenting with in-memory cache eviction policies in Go. The root package provides a generic, mutex-protected key-value cache with explicit capacity checks and batched eviction. The repository also contains LRU, S3-FIFO, and SIEVE implementations used for comparison and workload experiments.

## SnapCache design

The root `snapcache` package stores entries in two structures:

- a map for constant-time key lookup; and
- a linked list that preserves insertion order for eviction.

`Set` inserts or updates an entry, `Get` performs a lookup, and `Evict` removes entries from the front of the main list. The current snapshot size is initialized to one, so each eviction call removes at most one entry. Capacity enforcement is explicit: callers can inspect `Full` and invoke `Evict` when needed.

```go
package main

import (
    "fmt"

    "github.com/doogunwo/snapcache"
)

func main() {
    cache := snapcache.New[string, int](3)

    cache.Set("alpha", 1)
    cache.Set("beta", 2)
    cache.Set("gamma", 3)

    if cache.Full() {
        cache.Evict()
    }

    value, ok := cache.Get("beta")
    fmt.Println(value, ok)
}
```

## Included policies

| Package | Implementation |
| --- | --- |
| `snapcache` | Generic map-plus-list cache with explicit snapshot eviction |
| `lru` | LRU cache implementation with namespace-aware keys and handles |
| `s3fifo` | Small/main FIFO queues, a ghost queue, frequency counters, TTL, and eviction callbacks |
| `sieve` | SIEVE eviction with visited bits, a moving hand, TTL, and eviction callbacks |

### S3-FIFO

The S3-FIFO implementation divides entries between a small queue and a main queue. Frequently reused entries can move from the small queue into the main queue, while evicted small-queue keys are tracked in a bounded ghost queue. The package also implements optional TTL cleanup through rotating expiration buckets.

### SIEVE

The SIEVE implementation maintains insertion order and a moving eviction hand. A cache hit marks an entry as visited. During eviction, the hand clears visited entries and continues until it finds an unvisited candidate. Optional TTL cleanup and eviction callbacks use the same cache interface as S3-FIFO.

## API summary

### Root SnapCache

| Method | Behavior |
| --- | --- |
| `New[K, V](maxSize)` | Creates a generic cache with the requested capacity |
| `Set(key, value)` | Adds a new entry or updates an existing value |
| `Get(key)` | Returns a value and a hit flag |
| `Full()` | Reports whether the main list has reached capacity |
| `Evict()` | Removes up to the configured snapshot count and returns the number removed |
| `Purge()` | Clears all stored entries |

### S3-FIFO and SIEVE

Both packages expose `Set`, `Get`, `Peek`, `Contains`, `Remove`, `Len`, `Purge`, `Close`, and `SetOnEvicted`. Their constructors also accept an optional TTL duration.

## Repository structure

```text
snapcache/
├── snapcache.go        # Root generic SnapCache implementation
├── lru/                # LRU implementation and benchmark code
├── s3fifo/             # S3-FIFO implementation, ghost queue, and tests
├── sieve/              # SIEVE implementation
├── test/               # Root SnapCache tests
├── go.mod
└── go.sum
```

## Project status

This repository is an experimental implementation for studying cache behavior and comparing eviction policies. It is not presented as a drop-in replacement for a production cache library. Benchmark helpers and tests are included alongside the implementations, but results depend on external workload CSV files and the selected cache capacity.
