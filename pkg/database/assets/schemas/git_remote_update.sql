DROP TABLE IF EXISTS git_remote_update;
CREATE TABLE git_remote_update
(
    id TEXT
        PRIMARY KEY,
    r  TEXT,
    ut TIMESTAMP,
    rs INTEGER
);
