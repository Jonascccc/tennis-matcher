CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE app_user (
    user_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE tennis_profile (
    user_id BIGINT PRIMARY KEY REFERENCES app_user(user_id) ON DELETE CASCADE,
    handedness TEXT CHECK (handedness IN ('R', 'L')) DEFAULT 'R',
    level_est FLOAT DEFAULT 2.0,
    elo INT DEFAULT 1200,
    preferred_formats TEXT[] DEFAULT ARRAY['SINGLES'],
    radius_km INT DEFAULT 10,
    availability JSONB DEFAULT '{}'::jsonb
);

CREATE TABLE user_locations (
    user_id BIGINT REFERENCES app_user(user_id) ON DELETE CASCADE,
    label TEXT,
    geom GEOGRAPHY(POINT, 4326),
    updated_at TIMESTAMP DEFAULT now(),
    PRIMARY KEY (user_id, label)
);

CREATE TABLE courts (
    court_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    surface TEXT DEFAULT 'hard',
    geom GEOGRAPHY(POINT, 4326)
);

INSERT INTO courts(name,surface,geom)
VALUES('Toronto Ramsden Park Tennis Courts', 'hard', ST_SetSRID(ST_MakePoint(-79.3949022534484, 43.67639762901083), 4326)::geography);