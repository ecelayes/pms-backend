ALTER TABLE reservations 
ADD COLUMN rate_plan_id UUID REFERENCES rate_plans(id);

CREATE INDEX idx_reservations_rate_plan ON reservations(rate_plan_id);
