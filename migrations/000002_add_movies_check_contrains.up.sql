ALTER TABLE movies ADD CONSTRAINT movies_release_date_check CHECK (date_part('year', release_date) BETWEEN 1888 AND date_part('year', now()));

ALTER TABLE movies ADD CONSTRAINT genres_length_check CHECK (array_length(genres, 1) BETWEEN 1 AND 5);
