# EVM Chain Performance Testing with tpser

**tpser** is a Go CLI for EVM chain performance testing: historical block-range TPS analysis and sustained load generation for any Ethereum-compatible node.

When you're operating or developing on an EVM-compatible blockchain — whether that's Ethereum mainnet, a Layer 2 like Linea, or a private network — you eventually need to answer two questions: how has this chain been performing, and how much load can it actually handle?

These are different problems. Historical analysis needs a block-range scanner that can crunch past data efficiently. Load testing needs a sustained transaction generator that can hammer a node at a configurable rate and measure what breaks first.

**tpser** is a single Go binary that covers both modes.

## Two Modes, One Tool

### Mode 1: Block-Range Analyser

The analyser mode scans a range of historical blocks, computes per-block TPS, gas utilisation, and transaction counts, and emits a structured report. This is what you reach for when you want to understand real-world chain behaviour — peak TPS windows, sustained throughput, how busy the chain has been around a specific incident.

```bash
tpser analyse \
  --rpc https://rpc.yourchain.example \
  --from 1000000 \
  --to 1001000 \
  --output report.json
```

The output includes per-block metrics and aggregate statistics:

```json
{
  "block_range": { "from": 1000000, "to": 1001000 },
  "total_transactions": 48291,
  "avg_tps": 15.3,
  "peak_tps": 87.2,
  "avg_gas_utilisation": 0.71
}
```

This is useful for capacity planning, post-incident reviews, and producing the kind of reproducible benchmarks that are worth including in engineering reports.

### Mode 2: Sustained Load Generator

The load generator creates and broadcasts raw EVM transactions at a configurable rate for a set duration. It uses pre-funded accounts to sign transactions, bypasses mempool congestion by constructing nonces manually, and measures actual throughput rather than submitted throughput.

```bash
tpser load \
  --rpc https://rpc.yourchain.example \
  --private-key 0xYOURPRIVATEKEY \
  --tps 50 \
  --duration 300s
```

This drives 50 transactions per second into the target node for five minutes, then reports how many were included in blocks versus how many were dropped or delayed. That delta is what tells you where the bottleneck actually is — the node's ingestion rate, the block production rate, or the network propagation time.

## Targeting Any EVM-Compatible Chain

tpser works against any JSON-RPC endpoint that conforms to the Ethereum spec. The `--rpc` flag accepts HTTP and WebSocket endpoints:

```bash
# Mainnet via Infura
tpser analyse --rpc https://mainnet.infura.io/v3/YOUR_KEY --from 19000000 --to 19001000

# Local development node
tpser load --rpc http://localhost:8545 --private-key 0xac0974bec39... --tps 100 --duration 60s

# Layer 2 testnet
tpser load --rpc https://rpc.sepolia.linea.build --private-key 0x... --tps 200 --duration 120s
```

## Automated Slack Reporting

For CI pipelines or scheduled benchmarks, tpser supports Slack webhook integration. Each completed run can post a summary automatically:

```bash
tpser analyse \
  --rpc https://rpc.yourchain.example \
  --from 19000000 --to 19001000 \
  --slack-webhook https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

This is how the Polygon Edge team used a predecessor of this tool: every deployment triggered a benchmark run, and the Slack report gave engineers immediate visibility into whether the new build had regressed throughput.

## What to Look For in Load Test Results

When interpreting load generator output, a few numbers matter most:

**Inclusion rate** — the ratio of submitted transactions to on-chain inclusions. A rate below 95% under sustained load usually points to mempool backpressure or a misconfigured gas price floor.

**Block fullness** — if blocks are consistently at 100% gas utilisation during the test, the chain's theoretical maximum TPS is being hit. Back off the load and watch where the ceiling is.

**Latency distribution** — how long from submission to inclusion? p50, p95, and p99 matter separately. A low p50 with a high p99 often signals intermittent block production issues rather than a throughput ceiling.

## Building and Running

tpser is a standard Go module with no CGo dependencies:

```bash
git clone https://github.com/ZeljkoBenovic/tpser
cd tpser
go build -o tpser .
./tpser --help
```

Or pull the latest release binary directly from the GitHub releases page and drop it into your `$PATH`.

## Conclusion

Whether you're benchmarking a new node release, validating infrastructure capacity before a traffic spike, or producing reproducible performance data for a post-mortem, tpser gives you a single tool with no runtime dependencies and a clean CLI interface.

Browse the [tpser source on GitHub](https://github.com/ZeljkoBenovic/tpser) and run your first EVM chain benchmark.
