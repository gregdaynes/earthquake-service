CREATE TABLE IF NOT EXISTS entries
(
    id integer
        constraint sample_table_pk primary key,
    guid text,
    title text,
    content text,
    updated timestamp,
    published timestamp,
    categories string,
    elevation integer,
    latitude real,
    longitude real,
    magnitude real,
    time timestamp
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_entries_guid
ON entries ("guid");

CREATE INDEX IF NOT EXISTS idx_entries_latlng
ON entries (latitude, longitude);

CREATE INDEX IF NOT EXISTS idx_time
ON entries (time);
