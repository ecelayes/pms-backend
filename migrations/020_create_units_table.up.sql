CREATE TABLE IF NOT EXISTS units (
    id UUID PRIMARY KEY,
    property_id UUID NOT NULL, -- references properties(id) ... we can add FK constaint if needed but usually better to be explicit or use loose coupling
    unit_type_id UUID NOT NULL, -- references unit_types(id)
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_units_property_id ON units(property_id);
CREATE INDEX idx_units_unit_type_id ON units(unit_type_id);
