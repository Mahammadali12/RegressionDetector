-- Drift Detector — Database Schema
-- Run once on a fresh Postgres instance.
-- In Docker, place this file in db/init.sql and mount it into
-- /docker-entrypoint-initdb.d/ — Postgres runs it automatically on first start.

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
    query_id         BIGINT        NOT NULL,
    query            TEXT,
    snapshot_time    TIMESTAMPTZ   NOT NULL,
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