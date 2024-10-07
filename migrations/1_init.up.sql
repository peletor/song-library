-- Groups
CREATE TABLE IF NOT EXISTS groups (
                                    id SERIAL PRIMARY KEY,
                                    name TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_group_name ON groups (name);

-- Songs
CREATE TABLE IF NOT EXISTS songs (
                                    id SERIAL PRIMARY KEY,
                                    group_id INT REFERENCES groups(id),
                                    name TEXT NOT NULL,
                                    release_date DATE NOT NULL DEFAULT '0001-01-01'::DATE,
                                    text TEXT[] NOT NULL DEFAULT '{}'::text[],
                                    link TEXT DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_song_name ON songs (name);

CREATE INDEX IF NOT EXISTS idx_release_date ON songs (release_date);