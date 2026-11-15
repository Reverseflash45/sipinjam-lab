CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS bookings (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL REFERENCES users(id),
    room_id     INT REFERENCES rooms(id),
    purpose     VARCHAR(200) NOT NULL,
    start_at    TIMESTAMPTZ NOT NULL,
    end_at      TIMESTAMPTZ NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled', 'returned')),
    reviewed_by INT REFERENCES users(id),
    review_note VARCHAR(200),
    returned_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_at > start_at)
);

CREATE TABLE IF NOT EXISTS booking_items (
    booking_id INT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    item_id    INT NOT NULL REFERENCES items(id),
    quantity   INT NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (booking_id, item_id)
);

ALTER TABLE bookings
    ADD CONSTRAINT bookings_room_no_overlap
    EXCLUDE USING gist (
        room_id WITH =,
        tstzrange(start_at, end_at, '[)') WITH &&
    )
    WHERE (status = 'approved' AND room_id IS NOT NULL);