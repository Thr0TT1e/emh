-- +goose Up
-- +goose StatementBegin

-- Связь герой <-> награда
CREATE TABLE hero_awards (
    hero_id UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    award_id UUID NOT NULL REFERENCES awards(id) ON DELETE RESTRICT,
    award_date DATE,
    decree_number VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (hero_id, award_id)
);

CREATE INDEX idx_hero_awards_award ON hero_awards (award_id);

-- Связь герой <-> конфликт
CREATE TABLE hero_conflicts (
    hero_id UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    conflict_id UUID NOT NULL REFERENCES conflicts(id) ON DELETE RESTRICT,
    specific_location VARCHAR(500),
    rank_at_conflict VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (hero_id, conflict_id)
);

CREATE INDEX idx_hero_conflicts_conflict ON hero_conflicts (conflict_id);

-- Связь герой <-> локация
CREATE TABLE hero_locations (
    hero_id UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE RESTRICT,
    type SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (hero_id, location_id, type)
);

CREATE INDEX idx_hero_locations_location ON hero_locations (location_id);
CREATE INDEX idx_hero_locations_type ON hero_locations (type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS hero_locations;
DROP TABLE IF EXISTS hero_conflicts;
DROP TABLE IF EXISTS hero_awards;
-- +goose StatementEnd
