CREATE TABLE IF NOT EXISTS candidates (
    id SERIAL NOT NULL,
    number SMALLINT NOT NULL,
    president VARCHAR(255) NOT NULL,
    vice VARCHAR(255) NOT NULL,
    vision TEXT,
    mission TEXT,
    photo_key VARCHAR(255) NOT NULL,
    president_study_program VARCHAR(100) NOT NULL,     
    vice_study_program VARCHAR(100) NOT NULL,
    president_NIM VARCHAR(14) NOT NULL,
    vice_NIM VARCHAR(14) NOT NULL,     
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE (photo_key, president_NIM, vice_NIM)
)