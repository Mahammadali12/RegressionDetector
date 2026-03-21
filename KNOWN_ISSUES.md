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

- [ ] mean_exec_time is a cumulative average — slow calls diluted by history.
      Fix: track delta between snapshots. Major rework, v2.

- [ ] No p95 detection. Mean hides tail latency regressions.

### Thresholds
- [ ] absChange > 50ms and percChange > 30% are hardcoded globally.
      Needs per-project configurability before pilot.

- [ ] Minimum sample count of 3 is too low for production reliability.

- [ ] Minimum stddev floor hardcoded at 5% of mean.

### Correlation
- [ ] Lookback window (1 hour before anomaly) and forward window (10 minutes
      after anomaly) are hardcoded constants in engine/correlator.go.
      Needs per-project configurability before pilot.

- [ ] No correlation for data-growth-driven regressions — if no deploy falls
      in the correlation window, the cause is invisible. Table/tuple size
      snapshots needed to surface this class of regression.

### UI
- [ ] Dashboard plots anomaly events only, not continuous time-series.
      Requires GET /api/v1/snapshots/{query_id} endpoint (not built).

- [ ] No deploy/migration event overlay.

- [ ] Auto-refresh is polling every 15s. Should be SSE or websocket in v2.

- [ ] GET /api/v1/anomalies returns every row with no pagination or limit.
      The dashboard will degrade as the anomaly log grows.

### Infrastructure
- [ ] Collector and monitored DB must be network-adjacent. No documentation for
      customer network setup yet. Needed before pilot week.

- [ ] No data retention enforcement. Storage grows unbounded.

- [ ] notify/notifier.go Notify() doesn't short-circuit when WebhookURL is empty.
      It attempts http.NewRequestWithContext with an empty URL, which errors and
      gets logged. It works, but it's not a clean no-op — it's an error path
      masquerading as an intentional skip.

- [ ] No transaction wraps the baseline update + anomaly insert in detector.go.
      A crash between those two writes leaves the DB in an inconsistent state.

- [ ] docker-compose.yml maps Postgres as 5433:5433 but Postgres listens on 5432
      internally. Correct mapping is 5433:5432. Direct local connections will fail.

- [ ] GitHub webhook signature verification secret is loaded from env but there
      is no validation that it meets minimum entropy requirements.

- [ ] store/snapshot.go contains both Save() (snapshots) and GetAll() (anomalies) — two unrelated domain concerns in one file. GetAll() should move to store/anomaly.go.

### Performance
- [ ] store.Save() does N individual Exec calls in a loop — one per record.
      Should be a batch insert. Under any real load this will be slow and
      burns unnecessary DB connections.

- [ ] source/postgres.go opens a brand new pgx.Connect on every polling cycle
      instead of holding a pool. Every 10 seconds = a new TCP connection.
      Fine for MVP, bad at scale.

- [ ] sink/http.go uses http.DefaultClient with no timeout.
      A slow or hung API server will block the collector goroutine forever.

- [ ] api/ingest.go returns 400 Bad Request when store.Save() fails.
      Database errors should return 500 Internal Server Error.