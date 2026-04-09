CREATE TABLE IF NOT EXISTS rates (
    id UUID PRIMARY KEY,
    ask NUMERIC(20,8) NOT NULL,
    bid NUMERIC(20,8) NOT NULL,
    strategy VARCHAR(32) NOT NULL,
    n INT NOT NULL,
    m INT NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rates_fetched_at ON rates (fetched_at DESC);
