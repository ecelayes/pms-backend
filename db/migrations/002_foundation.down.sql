DROP TRIGGER IF EXISTS update_users_modtime ON users;
DROP TRIGGER IF EXISTS update_organizations_modtime ON organizations;
DROP TRIGGER IF EXISTS update_organization_members_modtime ON organization_members;
DROP TRIGGER IF EXISTS update_properties_modtime ON properties;
DROP TRIGGER IF EXISTS update_unit_types_modtime ON unit_types;

DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;
