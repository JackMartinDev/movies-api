CREATE TABLE IF NOT EXISTS movies (
    id bigserial PRIMARY KEY,  
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    overview text NOT NULL,
    language text NOT NULL,
    release_date date NOT NULL,
    rating float NOT NULL,
    poster_url text NOT NULL,
    backdrop_url text NOT NULL,
    genres text[] NOT NULL,
    version integer NOT NULL DEFAULT 1
);
