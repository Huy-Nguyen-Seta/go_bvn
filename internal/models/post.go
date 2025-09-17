package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	Tags      StringArray `json:"tags" db:"tags"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type PostRequest struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Tags    []string `json:"tags"`
}

type PostResponse struct {
	*Post
	RelatedPosts []Post `json:"related_posts,omitempty"`
}

type ActivityLog struct {
	ID       int64     `json:"id" db:"id"`
	Action   string    `json:"action" db:"action"`
	PostID   int64     `json:"post_id" db:"post_id"`
	LoggedAt time.Time `json:"logged_at" db:"logged_at"`
}

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	return pq.StringArray(s).Value()
}

func (s *StringArray) Scan(value interface{}) error {
	pqArray := pq.StringArray{}
	if err := pqArray.Scan(value); err != nil {
		return err
	}
	*s = StringArray(pqArray)
	return nil
}

func (s StringArray) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(s))
}

func (s *StringArray) UnmarshalJSON(data []byte) error {
	var slice []string
	if err := json.Unmarshal(data, &slice); err != nil {
		return err
	}
	*s = StringArray(slice)
	return nil
}