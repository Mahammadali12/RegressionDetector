package engine_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"regressiondetector/engine"
	"regressiondetector/internal/collector/types"
)

// ── schema ────────────────────────────────────────────────────────────────────

const schema = `
CREATE TABLE IF NOT EXISTS baseline_stats (
	query_id     BIGINT PRIMARY KEY,
	mean         DOUBLE PRECISION,
	stddev       DOUBLE PRECISION,
	sample_count BIGINT,
	last_seen    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS anomaly_records (
	id              BIGSERIAL PRIMARY KEY,
	query_id        BIGINT,
	window_start    TIMESTAMPTZ,
	window_end      TIMESTAMPTZ,
	metric          TEXT,
	z_score         DOUBLE PRECISION,
	absolute_change DOUBLE PRECISION,
	baseline_mean   DOUBLE PRECISION
);

CREATE TABLE IF NOT EXISTS snapshot_records (
	query_id         BIGINT,
	query            TEXT,
	snapshot_time    TIMESTAMPTZ NOT NULL,
	calls            BIGINT,
	total_exec_time  DOUBLE PRECISION,
	mean_exec_time   DOUBLE PRECISION,
	stddev_exec_time DOUBLE PRECISION,
	rows             BIGINT,
	shared_blks_read BIGINT,
	shared_blks_hit  BIGINT,
	temp_blks_read   BIGINT,
	PRIMARY KEY (query_id, snapshot_time)
);
`

// ── test harness ──────────────────────────────────────────────────────────────

func setupDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	connStr := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	if _, err := pool.Exec(ctx, schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	cleanup := func() {
		pool.Close()
		if err := container.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate container: %v", err)
		}
	}

	return pool, cleanup
}

// makeRow builds a synthetic PgStatRow for a given query ID and mean exec time.
func makeRow(queryID int64, meanExecTime float64) types.PgStatRow {
	return types.PgStatRow{
		QueryID:      queryID,
		SnapshotTime: time.Now().UTC(),
		MeanExecTime: meanExecTime,
	}
}

// countAnomalies returns the number of anomaly records for a given query ID.
func countAnomalies(t *testing.T, pool *pgxpool.Pool, queryID int64) int {
	t.Helper()
	var count int
	err := pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM anomaly_records WHERE query_id = $1", queryID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count anomalies: %v", err)
	}
	return count
}

// seedBaseline feeds n stable samples into the detector to build a real baseline.
func seedBaseline(t *testing.T, d *engine.Detector, queryID int64, meanMs float64, n int) {
	t.Helper()
	ctx := context.Background()
	for i := 0; i < n; i++ {
		if err := d.Analyze(ctx, makeRow(queryID, meanMs)); err != nil {
			t.Fatalf("seed sample %d failed: %v", i, err)
		}
	}
}

type noopNotifier struct{}

func (noopNotifier) Notify(ctx context.Context, queryID int64, zScore float64, absChange float64, baselineMean float64) error {
	return nil
}

type spyNotifier struct {
	mu    sync.Mutex
	calls int
}

func (s *spyNotifier) Notify(ctx context.Context, queryID int64, zScore float64, absChange float64, baselineMean float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return nil
}

func (s *spyNotifier) CallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

// ── tests ─────────────────────────────────────────────────────────────────────

// TestNoAnomalyBeforeMinimumSampleCount verifies the detector does not fire
// on the first few samples before a stable baseline has been established.
func TestNoAnomalyBeforeMinimumSampleCount(t *testing.T) {
	pool, cleanup := setupDB(t)
	defer cleanup()

	d := engine.NewDetector(pool, noopNotifier{})
	ctx := context.Background()

	const queryID = int64(1001)

	// Feed only 2 samples — below the minimum required for z-score computation.
	if err := d.Analyze(ctx, makeRow(queryID, 10.0)); err != nil {
		t.Fatalf("sample 1 failed: %v", err)
	}
	if err := d.Analyze(ctx, makeRow(queryID, 200.0)); err != nil {
		t.Fatalf("sample 2 failed: %v", err)
	}

	if n := countAnomalies(t, pool, queryID); n != 0 {
		t.Errorf("expected 0 anomalies before baseline is established, got %d", n)
	}
}

// TestNoAnomalyOnStableQuery verifies a query with consistent performance
// never triggers a false positive, regardless of how many samples accumulate.
func TestNoAnomalyOnStableQuery(t *testing.T) {
	pool, cleanup := setupDB(t)
	defer cleanup()

	d := engine.NewDetector(pool, noopNotifier{})
	const queryID = int64(1002)

	// 30 stable samples at ~10ms with tiny natural jitter.
	baselines := []float64{
		10.1, 10.0, 9.9, 10.2, 10.1, 10.0, 9.8, 10.3,
		10.1, 10.0, 9.9, 10.2, 10.1, 10.0, 9.8, 10.3,
		10.1, 10.0, 9.9, 10.2, 10.1, 10.0, 9.8, 10.3,
		10.1, 10.0, 9.9, 10.2, 10.1, 10.0,
	}

	ctx := context.Background()
	for i, ms := range baselines {
		if err := d.Analyze(ctx, makeRow(queryID, ms)); err != nil {
			t.Fatalf("sample %d failed: %v", i, err)
		}
	}

	if n := countAnomalies(t, pool, queryID); n != 0 {
		t.Errorf("stable query produced %d false positive anomalies, expected 0", n)
	}
}

// TestAnomalyDetectedOnClearRegression verifies the detector catches a query
// that goes from ~10ms to ~200ms after a stable baseline is established.
func TestAnomalyDetectedOnClearRegression(t *testing.T) {
	pool, cleanup := setupDB(t)
	defer cleanup()

	d := engine.NewDetector(pool, noopNotifier{})
	const queryID = int64(1003)

	// Build a solid baseline: 20 samples at ~10ms.
	seedBaseline(t, d, queryID, 10.0, 20)

	// Inject a clear regression: 200ms is a 190ms absolute change, well above
	// the 50ms threshold, and many standard deviations above the baseline mean.
	ctx := context.Background()
	if err := d.Analyze(ctx, makeRow(queryID, 200.0)); err != nil {
		t.Fatalf("regression sample failed: %v", err)
	}

	if n := countAnomalies(t, pool, queryID); n == 0 {
		t.Error("expected anomaly to be detected after clear regression, got 0")
	}
}

// TestNearZeroStddevDoesNotProduceSpuriousAnomaly verifies that a query with
// extremely low variance (stddev ≈ 0) does not trigger a false positive due to
// NaN or Inf z-score from division by near-zero stddev.
//
// This guards against the 10,000+ z-score bug observed in early testing.
func TestNearZeroStddevDoesNotProduceSpuriousAnomaly(t *testing.T) {
	pool, cleanup := setupDB(t)
	defer cleanup()

	d := engine.NewDetector(pool, noopNotifier{})
	const queryID = int64(1004)

	// All samples at exactly the same value — stddev will be 0 or near-zero.
	seedBaseline(t, d, queryID, 10.0, 20)

	// A 30ms reading: absolute change is 20ms — below the 50ms threshold.
	// With near-zero stddev, a naive z-score would be enormous.
	// The detector must NOT fire here.
	ctx := context.Background()
	if err := d.Analyze(ctx, makeRow(queryID, 30.0)); err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	if n := countAnomalies(t, pool, queryID); n != 0 {
		t.Errorf("near-zero stddev produced %d spurious anomaly records, expected 0", n)
	}
}

// TestNoDuplicateAnomaliesForSustainedRegression verifies that a sustained
// regression — where the query stays slow across multiple polling cycles —
// does not insert a new anomaly record on every single cycle.
//
// Known limitation: the current implementation WILL insert on every cycle
// because it has no "anomaly active" state tracking. This test documents
// that behaviour and will fail until deduplication is implemented.
func TestNoDuplicateAnomaliesForSustainedRegression(t *testing.T) {
	pool, cleanup := setupDB(t)
	defer cleanup()

	d := engine.NewDetector(pool, noopNotifier{})
	const queryID = int64(1005)

	// Build stable baseline.
	seedBaseline(t, d, queryID, 10.0, 20)

	// Simulate 5 consecutive polling cycles where the query stays slow.
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if err := d.Analyze(ctx, makeRow(queryID, 200.0)); err != nil {
			t.Fatalf("slow sample %d failed: %v", i, err)
		}
	}

	n := countAnomalies(t, pool, queryID)
	if n > 1 {
		// This documents the known duplicate-insertion bug.
		// TODO: implement anomaly deduplication / "active anomaly" state tracking.
		t.Errorf("sustained regression inserted %d duplicate anomaly records; expected 1 — "+
			"deduplication not yet implemented (see KNOWN_ISSUES.md)", n)
	}
}

// TestSmallAbsoluteChangeDoesNotTrigger verifies that a statistically
// significant but practically irrelevant change — e.g. 1ms → 3ms — does not
// fire an anomaly, because it falls below the 50ms absolute threshold.
func TestSmallAbsoluteChangeDoesNotTrigger(t *testing.T) {
	pool, cleanup := setupDB(t)
	defer cleanup()

	d := engine.NewDetector(pool, noopNotifier{})
	const queryID = int64(1006)

	// Baseline at 1ms.
	seedBaseline(t, d, queryID, 1.0, 20)

	// 3ms is a 200% relative change but only 2ms absolute — below threshold.
	ctx := context.Background()
	if err := d.Analyze(ctx, makeRow(queryID, 3.0)); err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	if n := countAnomalies(t, pool, queryID); n != 0 {
		t.Errorf("small absolute change (2ms) produced %d anomalies, expected 0", n)
	}
}

// TestBaselineMeanUpdatesAfterSamples verifies that the rolling mean in
// baseline_stats actually shifts as new samples arrive.
func TestBaselineMeanUpdatesAfterSamples(t *testing.T) {
	pool, cleanup := setupDB(t)
	defer cleanup()

	d := engine.NewDetector(pool, noopNotifier{})
	const queryID = int64(1007)
	ctx := context.Background()

	seedBaseline(t, d, queryID, 10.0, 5)

	var meanBefore float64
	pool.QueryRow(ctx,
		"SELECT mean FROM baseline_stats WHERE query_id = $1", queryID).
		Scan(&meanBefore)

	// Feed 5 more samples at 50ms — mean should increase.
	seedBaseline(t, d, queryID, 50.0, 5)

	var meanAfter float64
	pool.QueryRow(ctx,
		"SELECT mean FROM baseline_stats WHERE query_id = $1", queryID).
		Scan(&meanAfter)

	if meanAfter <= meanBefore {
		t.Errorf("baseline mean did not increase after slow samples: before=%.3f after=%.3f",
			meanBefore, meanAfter)
	}
}

// TestNotifierCalledWhenAnomalyInserted verifies detector invokes notifier
// when a qualifying anomaly is inserted.
func TestNotifierCalledWhenAnomalyInserted(t *testing.T) {
	pool, cleanup := setupDB(t)
	defer cleanup()

	spy := &spyNotifier{}
	d := engine.NewDetector(pool, spy)
	const queryID = int64(1008)

	seedBaseline(t, d, queryID, 10.0, 20)

	ctx := context.Background()
	if err := d.Analyze(ctx, makeRow(queryID, 200.0)); err != nil {
		t.Fatalf("regression sample failed: %v", err)
	}

	if n := countAnomalies(t, pool, queryID); n == 0 {
		t.Fatal("expected anomaly to be inserted before notifier assertion")
	}

	if spy.CallCount() != 1 {
		t.Fatalf("expected notifier to be called once, got %d", spy.CallCount())
	}
}
