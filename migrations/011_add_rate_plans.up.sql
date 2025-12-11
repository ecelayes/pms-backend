CREATE TABLE rate_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hotel_id UUID NOT NULL REFERENCES hotels(id),
    room_type_id UUID REFERENCES room_types(id),
    
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

CREATE INDEX idx_rate_plans_hotel ON rate_plans(hotel_id);
CREATE INDEX idx_rate_plans_room ON rate_plans(room_type_id);
