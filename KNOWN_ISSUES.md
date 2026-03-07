## Known Issues & Technical Debt

### Detection
- [ ] Z-score can be unrealistically high when baseline stddev is near-zero.
      Mitigated by 5% mean floor in effectiveStddev, but floor value is hardcoded.
      Needs per-project tuning.

- [ ] Baseline accumulates all samples forever — no 14-day rolling window.
      Long-running stable queries will eventually bury regressions.
      Fix: evict samples older than 14 days.

- [ ] Baseline absorbs regressed values over time, masking sustained regressions.
      Fix: freeze baseline updates when anomaly is active.

- [ ] Duplicate anomaly records inserted on every polling cycle during sustained
      regression. No "active anomaly" state tracking.
      Tracked by: TestNoDuplicateAnomaliesForSustainedRegression (currently failing).

- [ ] mean_exec_time is a cumulative average — slow calls diluted by history.
      Fix: track delta between snapshots. Major rework, v2.

- [ ] No p95 detection. Mean hides tail latency regressions.

### Thresholds
- [ ] absChange > 50ms and percChange > 30% are hardcoded globally.
      Needs per-project configurability before pilot.

- [ ] Minimum sample count of 3 is too low for production reliability.

- [ ] Minimum stddev floor hardcoded at 5% of mean.

### UI
- [ ] Dashboard plots anomaly events only, not continuous time-series.
      Requires GET /api/v1/snapshots/{query_id} endpoint (not built).

- [ ] No deploy/migration event overlay.

- [ ] Auto-refresh is polling every 15s. Should be SSE or websocket in v2.

### Infrastructure
- [ ] No Docker setup. Tests require local Docker daemon.
- [ ] No data retention enforcement. Storage grows unbounded.