# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	campus-stationery/cmd/app	[no test files]
--- FAIL: TestReferencedMarkerDeletionIsRejectedAndPurchaseOrderRemainsValid (0.00s)
    service_test.go:79: delete error = <nil>, want product referenced
    service_test.go:82: marker lookup: product not found: P0003
    service_test.go:85: purchase order lookup: invalid purchase order: missing product P0003
FAIL
FAIL	campus-stationery/internal/stationery	0.003s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/app): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/app): exit `0`
