CREATE TABLE IF NOT EXISTS reviews (
    id bigserial PRIMARY KEY,  
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    content text NOT NULL,
    stars integer NOT NULL DEFAULT 0,
    spoiler boolean NOT NULL DEFAULT false,
    version integer NOT NULL DEFAULT 1,
    movie_id bigint NOT NULL,
    user_id bigint NOT NULL,
    CONSTRAINT fk_movie FOREIGN KEY (movie_id) REFERENCES movies (id) ON DELETE CASCADE,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
