-- Rate Plans with Meal Plan, Cancellation Policy, Payment Policy

CREATE TABLE rate_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    unit_type_id UUID REFERENCES unit_types(id),
    name TEXT NOT NULL,
    description TEXT,
    meal_plan JSONB NOT NULL DEFAULT '{}',
    cancellation_policy JSONB NOT NULL DEFAULT '{}',
    payment_policy JSONB NOT NULL DEFAULT '{}',
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);
CREATE TRIGGER update_rate_plans_modtime 
    BEFORE UPDATE ON rate_plans 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE INDEX idx_rate_plans_property ON rate_plans(property_id);
CREATE INDEX idx_rate_plans_unit_type ON rate_plans(unit_type_id);
