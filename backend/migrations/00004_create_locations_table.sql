-- +goose Up
-- +goose StatementBegin

-- Места конфликтов
CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    historical_name VARCHAR(255),
    type SMALLINT NOT NULL DEFAULT 0,
    parent_id UUID REFERENCES locations(id) ON DELETE SET NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_locations_type ON locations (type);
CREATE INDEX idx_locations_parent ON locations (parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_locations_name_trgm ON locations USING gin (name gin_trgm_ops);
CREATE INDEX idx_locations_coords ON locations (latitude, longitude)
    WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON locations
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS locations;
-- +goose StatementEnd
