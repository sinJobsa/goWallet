CREATE TABLE IF NOT EXISTS rates (
    currency    VARCHAR(3) PRIMARY KEY,
    rate_to_rub NUMERIC(10, 4) NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
INSERT INTO rates (currency, rate_to_rub) VALUES
    ('USD', 90.5),
    ('EUR', 98.2),
    ('RUB', 1.0)
ON CONFLICT (currency) DO UPDATE SET
    rate_to_rub = EXCLUDED.rate_to_rub,
    updated_at = NOW();
