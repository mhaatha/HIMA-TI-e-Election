CREATE TABLE IF NOT EXISTS voting_access (
    user_id INT NOT NULL,
    hashed VARCHAR(255) NOT NULL,

    PRIMARY KEY (user_id),
    UNIQUE (hashed),
    CONSTRAINT fk_users
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
)