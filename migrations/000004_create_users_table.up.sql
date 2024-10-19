CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,  
    username varchar(25) UNIQUE NOT NULL,
    email varchar(100) UNIQUE NOT NULL,
    password_hash text NOT NULL,
    profile_url text NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
