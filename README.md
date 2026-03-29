# idstream

A simulation and benchmarking framework for comparing strategies that discover ads from an API with sequential integer IDs.

## The Problem

An ad platform assigns ads incrementing numeric IDs. Each ad takes 2–7 minutes to go live after creation, and roughly 50% fail permanently. A consumer polling `GET /ad/:id` must discover new ads as fast as possible — but every call returns the same error whether the ad doesn't exist yet, failed, or is just slow to go live. The question: **what polling strategy discovers the most ads with the lowest lag?**

## How It Works

The seed generates a realistic ad stream using Poisson-distributed arrival rates across configurable phases (e.g. slow → burst → cooldown). Algorithms poll a simulated `Accessor` that enforces the timing rules, and a `Recorder` wraps every call to collect stats.

### Algorithms

| Algorithm | Strategy |
|---|---|
| **Naive** | Single-threaded linear scan. Waits 1s on error before retrying the same index. |
| **Chaser** | N concurrent workers each racing to claim the next index. Retries up to M times with a 10s gap, then abandons and moves on. |
| **BinaryFrontier** | Queries `GetLatestLiveAd` to find the current frontier, computes its slice index from the ID, then concurrently scans the known window. Maintains a retry map for indexes that weren't live yet, expiring them after the max possible delay. |
| **Lookahead** | Single-threaded. On every tick, finds the frontier and scans N steps ahead of it. Simple and effective — catches ads as they go live by repeatedly sweeping the active window. |

### Sample Results (24h simulation, slow → burst → cooldown)

```
Algorithm      Attempts   Discovered   AvgLag       MaxLag
naive          86406      5            1m48s        3m0s
chaser         1039943    13837        2h26m        10h3m
binaryfrontier 4251784    78910        5s           4m29s
lookahead      43079256   85987        13s          1m21s
```

BinaryFrontier wins on lag. Lookahead wins on coverage.

## Running

```bash
go run .
```

Adjust the phases and simulation duration in `main.go`.

## Structure

```
seed/               Ad generation and accessor (the simulated API)
algorithms/         Naive, Chaser, BinaryFrontier, Lookahead
testing/common/     Getter interface, Recorder, Result types
testing/eval/       Frontier gap analysis
testing/compare/    Wires algorithms to the recorder and prints results
```
