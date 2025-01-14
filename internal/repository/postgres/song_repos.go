package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourusername/yourproject/internal/domain"
)

type songRepo struct {
	db *sqlx.DB
}

func NewSongRepository(db *sqlx.DB) domain.SongRepository {
	return &songRepo{db: db}
}

func (r *songRepo) Create(song *domain.Song) error {
	query := `
        INSERT INTO songs (group_name, title, release_date, text, link, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
        RETURNING id, created_at, updated_at
    `
	err := r.db.QueryRow(
		query,
		song.Group,
		song.Title,
		song.ReleaseDate,
		song.Text,
		song.Link,
	).Scan(&song.ID, &song.CreatedAt, &song.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "error creating song")
	}
	return nil
}

func (r *songRepo) GetByID(id int64) (*domain.Song, error) {
	var song domain.Song

	query := `
        SELECT
            id,
            group_name AS "group",
            title AS "title",
            release_date,
            text,
            link,
            created_at,
            updated_at
        FROM songs
        WHERE id = $1
    `
	err := r.db.Get(&song, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "error getting song by id")
	}
	return &song, nil
}

func (r *songRepo) GetAll(filter *domain.SongFilter) ([]domain.Song, int64, error) {
	// Базовый запрос
	baseQuery := `
        SELECT
            id,
            group_name AS "group",
            title,
            release_date,
            text,
            link,
            created_at,
            updated_at
        FROM songs
    `
	where := []string{}
	args := []interface{}{}

	// Фильтрация
	if filter.Group != nil && *filter.Group != "" {
		where = append(where, fmt.Sprintf("group_name ILIKE $%d", len(args)+1))
		args = append(args, "%"+*filter.Group+"%")
	}
	if filter.Title != nil && *filter.Title != "" {
		where = append(where, fmt.Sprintf("title ILIKE $%d", len(args)+1))
		args = append(args, "%"+*filter.Title+"%")
	}

	finalQuery := baseQuery
	if len(where) > 0 {
		finalQuery += " WHERE " + strings.Join(where, " AND ")
	}

	// Пагинация
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// Для подсчёта общего количества
	countQuery := `
        SELECT COUNT(*)
        FROM (
            ` + finalQuery + `
        ) as sub
    `

	var total int64
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "error counting songs")
	}

	finalQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d OFFSET %d", pageSize, offset)

	rows, err := r.db.Queryx(finalQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "error selecting songs")
	}
	defer rows.Close()

	var songs []domain.Song
	for rows.Next() {
		var s domain.Song
		if err := rows.StructScan(&s); err != nil {
			return nil, 0, errors.Wrap(err, "error scanning song")
		}
		songs = append(songs, s)
	}
	return songs, total, nil
}

func (r *songRepo) Update(song *domain.Song) error {
	query := `
        UPDATE songs
        SET
            group_name = $1,
            title = $2,
            release_date = $3,
            text = $4,
            link = $5,
            updated_at = NOW()
        WHERE id = $6
    `
	_, err := r.db.Exec(query, song.Group, song.Title, song.ReleaseDate, song.Text, song.Link, song.ID)
	if err != nil {
		return errors.Wrap(err, "error updating song")
	}
	return nil
}

func (r *songRepo) Delete(id int64) error {
	query := `DELETE FROM songs WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "error deleting song")
	}
	return nil
}
