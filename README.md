# idstream

A dynamic ID discovery engine for APIs that assign sequential integer IDs. Given a stream of resources that go live over time, idstream finds them as fast as possible with minimal lag.

## The Problem

Some APIs assign resources incrementing numeric IDs. A resource may not be immediately available after creation — it could be processing, delayed, or permanently failed — and the API returns the same error in all cases. A consumer has no way to know whether to retry or move on.

idstream implements several strategies for discovering these resources efficiently, each making different tradeoffs between coverage, lag, and API call volume.

## Algorithms

| Algorithm | Strategy |
|---|---|
| **Naive** | Single-threaded linear scan. Waits 1s on error before retrying. |
| **Chaser** | N concurrent workers racing to claim the next index. Retries up to M times with a 10s gap, then abandons and moves on. |
| **BinaryFrontier** | Anchors the live frontier via `GetLatestLiveAd`, computes its index from the ID, then concurrently scans the known window. Maintains a retry map for indexes that weren't live yet, expiring them after the maximum possible delay. |
| **Lookahead** | Single-threaded. On every tick, finds the frontier and sweeps N steps ahead of it — simple and effective at catching resources the moment they go live. |

## Benchmark Results (24h simulation, slow → burst → cooldown)

```
Algorithm      Attempts   Discovered   AvgLag       MaxLag
naive          86406      5            1m48s        3m0s
chaser         1039943    13837        2h26m        10h3m
binaryfrontier 4251784    78910        5s           4m29s
lookahead      43079256   85987        13s          1m21s
```

BinaryFrontier wins on lag. Lookahead wins on coverage.

## Running

The `seed/` package provides a simulated API for benchmarking — it generates a realistic resource stream with Poisson-distributed arrival rates and enforces live/failed timing rules. It is not part of the discovery engine itself.

```bash
go run .
```

Adjust the phases and simulation duration in `main.go`.

## Structure

```
algorithms/         Discovery strategies (Naive, Chaser, BinaryFrontier, Lookahead)
seed/               Simulated API for benchmarking
testing/            Recorder, result types, and comparison tooling
```
