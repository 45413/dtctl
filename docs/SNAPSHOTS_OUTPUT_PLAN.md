# Snapshots Output Summary

## Goal

Enable `dtctl query ... -o snapshot` to enrich each record with a decoded `parsed_snapshot` built from:

- `snapshot.data` (base64 protobuf payload)
- `snapshot.string_map` (string cache index mapping)

The implementation target was parity with the existing reference behavior (Variant2 / namespace conversion), not heuristic reconstruction.

## What Was Implemented

### 1) Snapshot output mode

- `-o snapshot` now enriches records with a single user-facing field: `parsed_snapshot`.
- Intermediate fields from earlier prototypes were removed from final output.

### 2) Typed protobuf decode path (reference-aligned)

The parser was migrated from schema-less wire decoding to typed decoding using:

- `dynatrace.com/protocols/v11/messages/rookout`
- `AugReportMessage`
- `Arguments2` (`Variant2` root)

Current flow:

1. Decode `snapshot.data` from base64.
2. Unmarshal into `AugReportMessage`.
3. Load string cache from `snapshot.string_map` (when supplied).
4. Build cache helpers (`strings`, `buffers`).
5. Convert `Variant2` recursively to dict-like output (reference style).

### 3) Variant conversion behavior

`variant2ToDict` now covers main and edge branches and emits reference-style keys such as:

- `@OT`
- `@CT`
- `@value`
- `@OS`
- `@attributes`
- `@max_depth`

Handled variant types include:

- `NONE`, `UNDEFINED`
- `INT`, `LONG`, `LARGE_INT`
- `DOUBLE`
- `STRING`, `MASKED`
- `BINARY`
- `TIME`
- `ENUM`
- `LIST`, `SET`, `MAP`
- `OBJECT`, `UKNOWN_OBJECT`, `NAMESPACE`
- `TRACEBACK`
- `ERROR`, `COMPLEX`, `LIVETAIL`
- `FORMATTED_MESSAGE`, `DYNAMIC`, `MAX_DEPTH`

### 4) Correctness decisions made during implementation

- Removed variable-name heuristics and hardcoded local/object shaping from old prototype code.
- Rejected recursive “base64-inside-string” parsing because it is not part of the reference behavior.
- Kept output contract stable around `parsed_snapshot`.

## Files Updated

- `pkg/output/snapshot.go`
  - Typed snapshot decode and Variant2 conversion.
  - Cache handling and recursive dict conversion.
  - Timestamp formatting and large integer safety behavior.

- `pkg/output/snapshot_test.go`
  - Switched to typed fixture payloads (`AugReportMessage` marshaling).
  - Added edge-case regression coverage (formatted message, large int, set/reverse order, error, timestamp).

- `go.mod`, `go.sum`
  - Added direct dependency: `dynatrace.com/protocols/v11 v11.331.0`.

## Validation Performed

- Unit tests:
  - `go test ./pkg/output` passes.

- Formatting/build checks:
  - `gofmt` applied to snapshot output files.

- Runtime check example:
  - `./dtctl query "fetch application.snapshots | sort timestamp desc | limit 1" -o snapshot > snapshot.out4.json`
  - Output includes populated `parsed_snapshot` with nested decoded content.

## Current Status

✅ `-o snapshot` is fully wired and producing typed decoded snapshot output.

✅ Parser is aligned to reference-style Variant2 namespace conversion approach.

✅ Edge-case handling and tests are in place for the most relevant variant categories.

## Known Remaining Risk (Low)

Some very rare production-only Variant2 patterns may still require minor parity polish if observed in new live data samples. The core path and major edge branches are implemented and validated.