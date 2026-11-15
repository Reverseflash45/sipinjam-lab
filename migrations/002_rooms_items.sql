-- 002: ruang laboratorium dan alat yang dapat dipinjam.
CREATE TABLE IF NOT EXISTS rooms (
    id         SERIAL       PRIMARY KEY,
    code       VARCHAR(20)  NOT NULL UNIQUE,
    name       VARCHAR(100) NOT NULL,
    building   VARCHAR(50)  NOT NULL,
    capacity   INT          NOT NULL CHECK (capacity > 0),
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS items (
    id         SERIAL       PRIMARY KEY,
    code       VARCHAR(20)  NOT NULL UNIQUE,
    name       VARCHAR(100) NOT NULL,
    category   VARCHAR(50)  NOT NULL,
    stock      INT          NOT NULL CHECK (stock >= 0),
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
