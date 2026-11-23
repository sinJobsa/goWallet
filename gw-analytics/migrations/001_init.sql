CREATE DATABASE IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.events (
                                                transaction_id String,
                                                user_id        String,
                                                from_currency  String,
                                                to_currency    String,
                                                amount         Float64,
                                                status         String,
                                                created_at     DateTime,
                                                received_at    DateTime,
                                                latency_ms     Int64,
                                                ingested_at    DateTime DEFAULT now()
    ) ENGINE = ReplacingMergeTree(ingested_at)
    ORDER BY (transaction_id, created_at);

-- 3. Материализованное представление: агрегаты по 1 минуте
CREATE MATERIALIZED VIEW IF NOT EXISTS analytics.aggregates_1m
ENGINE = SummingMergeTree()
ORDER BY (window_start, status)
AS SELECT
              toStartOfMinute(created_at) AS window_start,
              status,
              count()          AS count,
    avg(latency_ms)  AS avg_latency_ms,
    max(latency_ms)  AS max_latency_ms,
    min(latency_ms)  AS min_latency_ms
   FROM analytics.events
   GROUP BY window_start, status;