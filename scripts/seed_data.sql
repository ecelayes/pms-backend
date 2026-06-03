TRUNCATE 
    reservations, 
    price_rules, 
    rate_plans, 
    unit_types, 
    units,
    properties, 
    organization_members, 
    users, 
    organizations, 
    guests,
    guest_services,
    amenities
CASCADE;

-- Organization
INSERT INTO organizations (id, name, code, created_at, updated_at)
VALUES (
    '018e9a9d-0c8e-7000-0000-000000000001',
    'Global Resorts Inc.',
    'GLB-INC',
    NOW(), NOW()
);

-- Super Admin
INSERT INTO users (id, email, password, salt, role, first_name, last_name, phone, created_at, updated_at) 
VALUES (
    '00000000-0000-0000-0000-000000000000', 
    'admin@platform.com', 
    '$2a$10$kypbnGGCpJ7UQlysnqzJG.6H.dUewn7UPVWA3Ip.E.8U4jlVnFNnu', -- 'password123'
    'salt_admin', 
    'super_admin',
    'System', 'Admin', '000-0000',
    NOW(), NOW()
);

-- CEO Owner
INSERT INTO users (id, email, password, salt, role, first_name, last_name, phone, created_at, updated_at) 
VALUES (
    '018e9a9d-0c8e-7000-0000-000000000002', 
    'ceo@globalresorts.com', 
    '$2a$10$kypbnGGCpJ7UQlysnqzJG.6H.dUewn7UPVWA3Ip.E.8U4jlVnFNnu', -- 'password123'
    'random_salt', 
    'user',
    'John', 'CEO', '555-9999',
    NOW(), NOW()
);

INSERT INTO organization_members (id, organization_id, user_id, role, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '018e9a9d-0c8e-7000-0000-000000000001',
    '00000000-0000-0000-0000-000000000000',
    'super_admin',
    NOW(), NOW()
);

INSERT INTO organization_members (id, organization_id, user_id, role, created_at, updated_at)
VALUES (
    '018e9a9d-0c8e-7000-0000-000000000003',
    '018e9a9d-0c8e-7000-0000-000000000001',
    '018e9a9d-0c8e-7000-0000-000000000002',
    'owner',
    NOW(), NOW()
);

-- Properties
INSERT INTO properties (id, organization_id, name, code, type, created_at, updated_at)
VALUES ('018e9a9d-0c8e-7000-0000-000000000004', '018e9a9d-0c8e-7000-0000-000000000001', 'Grand Miami Beach', 'MIA', 'HOTEL', NOW(), NOW());

-- Unit Types
INSERT INTO unit_types (id, property_id, name, code, total_quantity, base_price_cents, base_price_currency, max_occupancy, max_adults, max_children, amenities, created_at, updated_at)
VALUES ('018e9a9d-0c8e-7000-0000-000000000005', '018e9a9d-0c8e-7000-0000-000000000004', 'Ocean View Suite', 'OCN', 10, 15000, 'USD', 4, 2, 2, ARRAY['wifi', 'jacuzzi', 'tv', 'minibar'], NOW(), NOW());

-- Rate Plans
INSERT INTO rate_plans (id, property_id, unit_type_id, name, description, meal_plan, cancellation_policy, payment_policy, active, created_at, updated_at)
VALUES ('018e9a9d-0c8e-7000-0000-000000000009', '018e9a9d-0c8e-7000-0000-000000000004', '018e9a9d-0c8e-7000-0000-000000000005', 'Standard Rate', 'Flexible cancellation', '{"type": 0, "included": false, "price_per_pax": 0}', '{"is_refundable": true, "rules": []}', '{"timing": 0, "method": 0, "prepay_percent": 0}', true, NOW(), NOW());

-- Price Rules
INSERT INTO price_rules (id, unit_type_id, validity_range, price, price_cents, price_currency, created_at, updated_at)
VALUES ('018e9a9d-0c8e-7000-0000-000000000006', '018e9a9d-0c8e-7000-0000-000000000005', '[2025-01-01, 2025-12-31)', 250.00, 25000, 'USD', NOW(), NOW());

-- Guests
INSERT INTO guests (id, email, first_name, last_name, phone, created_at, updated_at) 
VALUES ('018e9a9d-0c8e-7000-0000-000000000007', 'leomessi@mail.com', 'Lionel', 'Messi', '10101010', NOW(), NOW());

-- Reservations
INSERT INTO reservations (id, reservation_code, property_id, unit_type_id, rate_plan_id, guest_id, guest_email, stay_range, total_price, price_cents, price_currency, status, adults, children, created_at, updated_at)
VALUES ('018e9a9d-0c8e-7000-0000-000000000008', 'MIA-OCN-SEED', '018e9a9d-0c8e-7000-0000-000000000004', '018e9a9d-0c8e-7000-0000-000000000005', '018e9a9d-0c8e-7000-0000-000000000009', '018e9a9d-0c8e-7000-0000-000000000007', 'leomessi@mail.com', '[2025-06-10, 2025-06-15)', 1250.00, 125000, 'USD', 'confirmed', 2, 2, NOW(), NOW());
