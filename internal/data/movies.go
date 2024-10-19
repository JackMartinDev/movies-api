package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"movies.jackmartin.net/internal/validator"
)

type Movie struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Overview    string    `json:"overview"`
	Language    string    `json:"language"`
	ReleaseDate time.Time `json:"release_date"`
	Rating      float32   `json:"vote_average"`
	PosterURL   string    `json:"poster_url"`
	BackdropURL string    `json:"backdrop_url"`
	Genres      []string  `json:"genres"`
	Version     int32     `json:"version"`
	CreatedAt   time.Time `json:"-"`
}

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `
    INSERT INTO movies (title, overview, language, release_date, rating, poster_url, backdrop_url, genres) 
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	args := []any{
		movie.Title,
		movie.Overview,
		movie.Language,
		movie.ReleaseDate,
		movie.Rating,
		movie.PosterURL,
		movie.BackdropURL,
		pq.Array(movie.Genres),
	}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
    SELECT id, created_at, title, overview, language, release_date, rating, poster_url, backdrop_url, genres, version
    FROM movies
    WHERE id = $1`

	var movie Movie

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Overview,
		&movie.Language,
		&movie.ReleaseDate,
		&movie.Rating,
		&movie.PosterURL,
		&movie.BackdropURL,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	query := fmt.Sprintf(`
    SELECT count(*) OVER(), id, created_at, title, overview, language, release_date, rating, poster_url, backdrop_url, genres, version
    FROM movies
    WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
    AND (genres @> $2 OR $2 = '{}')
    ORDER BY %s %s, id ASC
    LIMIT $3 OFFSET $4
    `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	args := []any{title, pq.Array(genres), filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0
	movies := []*Movie{}

	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Overview,
			&movie.Language,
			&movie.ReleaseDate,
			&movie.Rating,
			&movie.PosterURL,
			&movie.BackdropURL,
			pq.Array(&movie.Genres),
			&movie.Version,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query := `
    UPDATE movies 
    SET title = $1, overview = $2, language = $3, release_date = $4, rating = $5, poster_url = $6, backdrop_url = $7, genres = $8, version = version + 1
    WHERE id = $9 AND version = $10
    RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	args := []any{
		&movie.Title,
		&movie.Overview,
		&movie.Language,
		&movie.ReleaseDate,
		&movie.Rating,
		&movie.PosterURL,
		&movie.BackdropURL,
		pq.Array(&movie.Genres),
		movie.ID,
		movie.Version,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
    DELETE FROM movies
    WHERE id = $1
  `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(!movie.ReleaseDate.IsZero(), "release_date", "must be provided")
	v.Check(movie.ReleaseDate.Year() >= 1888, "release_date", "year must be greater than 1888")
	v.Check(movie.ReleaseDate.Before(time.Now()), "release_date", "must not be in the future")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 10, "genres", "must not contain more than 10 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}
