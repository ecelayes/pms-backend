CREATE INDEX IF NOT EXISTS idx_reservations_guest_id ON reservations(guest_id);
CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_room_types_hotel_id ON room_types(hotel_id);

CREATE INDEX IF NOT EXISTS idx_room_types_amenities ON room_types USING GIN (amenities);
