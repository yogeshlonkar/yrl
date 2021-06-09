DROP TABLE IF EXISTS git_status;
CREATE TABLE git_status
(
    pk INTEGER
        PRIMARY KEY,
    id TEXT,
    ss BLOB,
    ut TIMESTAMP
);
