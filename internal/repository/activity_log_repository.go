package repository

import (
	"blog-api/internal/models"
	"database/sql"
)

type ActivityLogRepository struct {
	db *sql.DB
}

func NewActivityLogRepository(db *sql.DB) *ActivityLogRepository {
	return &ActivityLogRepository{db: db}
}

func (r *ActivityLogRepository) Create(tx *sql.Tx, log *models.ActivityLog) error {
	query := `
		INSERT INTO activity_logs (action, post_id, logged_at)
		VALUES ($1, $2, NOW())
		RETURNING id, logged_at`

	err := tx.QueryRow(query, log.Action, log.PostID).
		Scan(&log.ID, &log.LoggedAt)
	
	return err
}