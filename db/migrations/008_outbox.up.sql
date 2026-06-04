-- Outbox pattern for reliable event publishing

CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_id UUID NOT NULL,
    type TEXT NOT NULL,
    payload JSONB NOT NULL,
    occurred_on TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ
);
CREATE INDEX idx_outbox_processed_at ON outbox(processed_at) WHERE processed_at IS NULL;
