CREATE TABLE IF NOT EXISTS users (
    id SERIAL NOT NULL,
    nim VARCHAR(14),
    full_name VARCHAR(100) NOT NULL,
    study_program VARCHAR(100),
    password VARCHAR(255),
    role user_role NOT NULL DEFAULT 'student',
    phone_number VARCHAR(14) NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE (nim, phone_number)
)