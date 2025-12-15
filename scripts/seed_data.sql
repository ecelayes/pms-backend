TRUNCATE 
    reservations, 
    price_rules, 
    rate_plans, 
    room_types, 
    hotels, 
    organization_members, 
    users, 
    organizations, 
    guests 
CASCADE;

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
    '018e9a9d-0c8e-7000-0000-000000000003',
    '018e9a9d-0c8e-7000-0000-000000000001',
    '018e9a9d-0c8e-7000-0000-000000000002',
    'owner',
    NOW(), NOW()
);

INSERT INTO hotels (id, organization_id, name, code, created_at, updated_at)
VALUES ('018e9a9d-0c8e-7000-0000-000000000004', '018e9a9d-0c8e-7000-0000-000000000001', 'Grand Miami Beach', 'MIA', NOW(), NOW());

INSERT INTO room_types (id, hotel_id, name, code, total_quantity, base_price, max_occupancy, max_adults, max_children, amenities, created_at, updated_at)
VALUES ('018e9a9d-0c8e-7000-0000-000000000005', '018e9a9d-0c8e-7000-0000-000000000004', 'Ocean View Suite', 'OCN', 10, 150.00, 4, 2, 2, ARRAY['wifi', 'jacuzzi', 'tv', 'minibar'], NOW(), NOW());

INSERT INTO rate_plans (id, hotel_id, room_type_id, name, description, meal_plan, cancellation_policy, payment_policy, active, created_at, updated_at)
VALUES ('018e9a9d-0c8e-7000-0000-000000000009', '018e9a9d-0c8e-7000-0000-000000000004', '018e9a9d-0c8e-7000-0000-000000000005', 'Standard Rate', 'Flexible cancellation', '{"type": 0, "included": false, "price_per_pax": 0}', '{"is_refundable": true, "rules": []}', '{"timing": 0, "method": 0, "prepay_percent": 0}', true, NOW(), NOW());

INSERT INTO price_rules (id, room_type_id, validity_range, price, created_at, updated_at)
VALUES ('018e9a9d-0c8e-7000-0000-000000000006', '018e9a9d-0c8e-7000-0000-000000000005', '[2025-01-01, 2025-12-31)', 250.00, NOW(), NOW());

INSERT INTO guests (id, email, first_name, last_name, phone, created_at, updated_at) 
VALUES ('018e9a9d-0c8e-7000-0000-000000000007', 'leomessi@mail.com', 'Lionel', 'Messi', '10101010', NOW(), NOW());

INSERT INTO reservations (id, reservation_code, room_type_id, rate_plan_id, guest_id, stay_range, total_price, status, adults, children, created_at, updated_at)
VALUES ('018e9a9d-0c8e-7000-0000-000000000008', 'MIA-OCN-SEED', '018e9a9d-0c8e-7000-0000-000000000005', '018e9a9d-0c8e-7000-0000-000000000009', '018e9a9d-0c8e-7000-0000-000000000007', '[2025-06-10, 2025-06-15)', 1250.00, 'confirmed', 2, 2, NOW(), NOW());
