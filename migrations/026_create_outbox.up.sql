CREATE TABLE outbox (
    id TEXT PRIMARY KEY, -- Repo passes res.ID() which is string? No, outbox ID is usually UUID but repo passes $1 which is res.ID()? 
    -- Wait. Repo Line 113: `res.ID()` passed as `$1` (id).
    -- And `res.ID()` passed as `$2` (aggregate_id).
    -- So `id` == `aggregate_id`. 
    -- If we have MULTIPLE events for same aggregate, PK conflict!
    -- Repo Code: `INSERT INTO outbox (id, aggregate_id, ...) VALUES ($1, $2, ...)`
    -- Args: `res.ID(), res.ID(), ...`
    -- This assumes 1:1 mapping between Reservation Creation and Event.
    -- If we update reservation later?
    -- `PostgresReservationRepo.Update` does NOT seem to insert event (Line 193).
    -- So strictly for "ReservationCreated", `id` = `aggregate_id` is fine IF only one creation event per ID.
    -- But usually Outbox ID should be unique random UUID.
    -- Repo logic seems to reuse Reservation ID as Outbox ID?
    -- Line 113 `res.ID()` as ID.
    -- This works for Creation.
    -- I'll define `id` as `TEXT` or `UUID`? `res.ID()` is string.
    -- Postgres `reservations.id` is UUID.
    -- I'll use UUID for `id`.
    
    aggregate_id UUID NOT NULL, 
    type TEXT NOT NULL,
    payload JSONB NOT NULL,
    occurred_on TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ
);

CREATE INDEX idx_outbox_processed_at ON outbox(processed_at) WHERE processed_at IS NULL;
