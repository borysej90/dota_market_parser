CREATE TABLE IF NOT EXISTS history
(
    id SERIAL NOT NULL CONSTRAINT history_pk PRIMARY KEY,
    item_id INT NOT NULL CONSTRAINT history_items_id_fk REFERENCES items,
    price FLOAT NOT NULL,
    quantity INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT now()
);
