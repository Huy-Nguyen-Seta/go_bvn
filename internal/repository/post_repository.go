package repository

import (
	"blog-api/internal/models"
	"database/sql"

	"github.com/lib/pq"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(tx *sql.Tx, post *models.Post) error {
	query := `
		INSERT INTO posts (title, content, tags, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := tx.QueryRow(query, post.Title, post.Content, pq.Array(post.Tags)).
		Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	
	return err
}

func (r *PostRepository) GetByID(id int64) (*models.Post, error) {
	query := `
		SELECT id, title, content, tags, created_at, updated_at
		FROM posts
		WHERE id = $1`

	post := &models.Post{}
	err := r.db.QueryRow(query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Tags,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return post, nil
}

func (r *PostRepository) Update(tx *sql.Tx, post *models.Post) error {
	query := `
		UPDATE posts
		SET title = $2, content = $3, tags = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := tx.QueryRow(query, post.ID, post.Title, post.Content, pq.Array(post.Tags)).
		Scan(&post.UpdatedAt)
	
	return err
}

func (r *PostRepository) SearchByTag(tag string, limit, offset int) ([]models.Post, error) {
	query := `
		SELECT id, title, content, tags, created_at, updated_at
		FROM posts
		WHERE $1 = ANY(tags)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.Tags,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (r *PostRepository) GetRelatedPosts(tags []string, excludeID int64, limit int) ([]models.Post, error) {
	if len(tags) == 0 {
		return []models.Post{}, nil
	}

	// Fix the query building - simplified approach
	query := `
		SELECT DISTINCT id, title, content, tags, created_at, updated_at
		FROM posts
		WHERE tags && $1 AND id != $2
		ORDER BY created_at DESC
		LIMIT $3`

	rows, err := r.db.Query(query, pq.Array(tags), excludeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.Tags,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (r *PostRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}