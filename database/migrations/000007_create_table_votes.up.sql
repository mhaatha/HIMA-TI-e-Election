CREATE TABLE IF NOT EXISTS votes (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    candidate_id INT NOT NULL,
    hashed_nim VARCHAR(255) NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    CONSTRAINT fk_candidates
        FOREIGN KEY (candidate_id)
        REFERENCES candidates (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_voting_access
        FOREIGN KEY (hashed_nim)
        REFERENCES voting_access (hashed)
        ON DELETE CASCADE    
);