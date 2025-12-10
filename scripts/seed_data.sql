TRUNCATE reservations, price_rules, room_types, hotels, guests, users CASCADE;

INSERT INTO users (id, email, password, salt, role) 
VALUES (
    '018e9a9d-0c8e-7f12-8d3a-123456789001',
    'admin@pms.com', 
    'hashed_password_placeholder', 
    'random_salt', 
    'admin'
);

INSERT INTO guests (id, email, first_name, last_name, phone) 
VALUES (
    '018e9a9d-0c8e-7445-9ac2-8f9202930432',
    'vip@guest.com', 
    'Lionel', 
    'Messi', 
    '10101010'
);

INSERT INTO hotels (id, organization_id, name, code)
VALUES (
    '018e9a9d-0c8e-7111-2222-333344445555',
    '018e9a9d-0c8e-7f12-8d3a-123456789001',
    'Hotel Seed',
    'SEED'
);

-- 4. Room Type
INSERT INTO room_types (id, hotel_id, name, code, total_quantity, max_occupancy, max_adults, max_children, amenities)
VALUES (
    '018e9a9d-0c8e-7666-7777-888899990000',
    '018e9a9d-0c8e-7111-2222-333344445555',
    'Suite Presidencial',
    'PRE',
    5,
    4,
    2,
    2,
    ARRAY['wifi', 'jacuzzi', 'tv']
);

INSERT INTO price_rules (id, room_type_id, validity_range, price, priority)
VALUES (
    '018e9a9d-0c8e-7aaa-bbbb-ccccddddeeee',
    '018e9a9d-0c8e-7666-7777-888899990000',
    '[2025-01-01, 2025-12-31)',
    150.00,
    0
);

INSERT INTO reservations (
    id, reservation_code, room_type_id, guest_id, 
    stay_range, total_price, status, adults, children
)
VALUES (
    '018e9a9d-0c8e-7fff-0000-111122223333',
    'SEED-PRE-X999',
    '018e9a9d-0c8e-7666-7777-888899990000',
    '018e9a9d-0c8e-7445-9ac2-8f9202930432',
    '[2025-02-10, 2025-02-15)',
    750.00,
    'confirmed',
    2,
    0
);
