CREATE TABLE IF NOT EXISTS votes_summary (
    candidate_id INT NOT NULL,
    total INT NOT NULL DEFAULT 0,
    
    CONSTRAINT fk_candidates
        FOREIGN KEY (candidate_id)
        REFERENCES candidates (id)
        ON DELETE CASCADE
);