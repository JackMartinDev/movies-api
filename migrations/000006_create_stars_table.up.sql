CREATE TABLE IF NOT EXISTS stars (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    review_id bigint NOT NULL,
    user_id bigint NOT NULL,
    CONSTRAINT fk_review FOREIGN KEY (review_id) REFERENCES reviews (id) ON DELETE CASCADE,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    UNIQUE (review_id, user_id)
);
