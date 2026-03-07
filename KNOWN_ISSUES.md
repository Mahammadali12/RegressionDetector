## Known Issues & Technical Debt

### Detection
- Z-score can be unrealistically high (10,000+) when baseline stddev is near-zero.
  Fix: add minimum stddev floor or require minimum sample variance before flagging.
  
- Baseline absorbs regressed values over time, gradually masking sustained regressions.
  Fix: freeze baseline updates when anomaly is active (needs anomaly state tracking).

- mean_exec_time is a cumulative average — slow calls are diluted by historical fast calls.
  Fix: track delta between snapshots instead of raw mean_exec_time (major rework, v2).

### Thresholds
- absChange > 50ms threshold is hardcoded. Needs per-project configurability.
- Sample count minimum of 3 is too low for production reliability.

### UI
- Dashboard plots anomaly events only, not continuous time-series.
  Full timeline view requires GET /api/v1/snapshots/{query_id} endpoint (not yet built).
- No deploy/migration event overlay (event_records endpoint not yet built).
- Auto-refresh is polling (15s). Should be websocket or SSE in v2.