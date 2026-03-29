# idstream

Real-time discovery engine for sequential ID streams. Crawls APIs that assign monotonically increasing integer IDs and surfaces new resources with minimal discovery latency.

## Problem

APIs that issue sequential IDs create a discovery problem: a consumer polling `GET /resource/:id` cannot distinguish between an ID that does not exist yet, is still processing, or has permanently failed — all return the same error. Naive linear polling is too slow; blind parallel workers waste throughput on dead IDs.

idstream implements and benchmarks several strategies for solving this, each making different tradeoffs between throughput, coverage, and discovery latency.

The main algo of the repo is the BinaryFrontier, but feel free to add your solutions to the problem. The less attempts means less hammering on the proxies.

## Algorithms

| Algorithm | Strategy |
|---|---|
| **Naive** | Single-threaded linear scan with 1s backoff on error. |
| **Chaser** | N concurrent workers with shared index coordination. Each worker retries its assigned index up to M times before abandoning and claiming the next. |
| **BinaryFrontier** | Uses `GetLatestLiveAd` to anchor the live frontier, derives its index arithmetically from the sequential ID, then fans out a concurrent scan over the known window. Tracks a retry set for indexes that were not yet live at scan time, evicting them after the maximum processing delay. |
| **Lookahead** | Single-threaded. On each tick, resolves the current frontier and sweeps a fixed window ahead of it. Low complexity, high coverage. |

## Benchmark

24-hour simulation across three arrival rate phases (slow → burst → cooldown), Poisson-distributed.

```
Total discoverable resources: 79,322

Algorithm        Attempts      Discovered   Coverage   AvgLatency   MaxLatency
naive            86,406        5            0.0%       1m48s        3m0s
chaser           2,090,137     16,377       20.6%      3h14m        13h51m
binaryfrontier   4,884,714     79,153       99.8%      5s           4m29s
lookahead        43,079,256    5,537        7.0%       509ms        7.9s
```

BinaryFrontier achieves 99.8% coverage with 5s average discovery latency.

## Usage

```bash
go run .
```

Configure arrival phases and simulation window in `main.go`.

## Structure

```
algorithms/    Discovery strategies
seed/          Simulated API for benchmarking (not part of the engine)
testing/       Recorder, result types, comparison tooling
```
