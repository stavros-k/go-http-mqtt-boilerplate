-- migrate:up
CREATE TABLE test_users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- migrate:down
DROP TABLE test_users;