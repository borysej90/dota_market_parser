CREATE TABLE IF NOT EXISTS items
(
    id serial NOT NULL constraint items_pk PRIMARY KEY,
    name VARCHAR NOT NULL,
    create_at TIMESTAMP DEFAULT now()
);
