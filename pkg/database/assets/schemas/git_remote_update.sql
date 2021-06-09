DROP TABLE IF EXISTS git_remote_update;
CREATE TABLE git_remote_update
(
    id TEXT
        PRIMARY KEY,
    ut TIMESTAMP,
    rs INTEGER
);
