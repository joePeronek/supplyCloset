CREATE TABLE IF NOT EXISTS inventory (
    id INT PRIMARY KEY,
    name TEXT NOT NULL,
    quantity INT NOT NULL
);

CREATE TABLE IF NOT EXISTS issued (
    id SERIAL PRIMARY KEY,
    item_id INT NOT NULL,
    item_name TEXT NOT NULL,
    person TEXT NOT NULL,
    issued_by TEXT NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL
);
