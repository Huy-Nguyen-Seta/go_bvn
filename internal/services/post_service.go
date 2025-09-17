package services

import (
	"blog-api/internal/models"
	"blog-api/internal/repository"
	"database/sql"
	"fmt"
)

type PostService struct {
	postRepo        *repository.PostRepository
	activityLogRepo *repository.ActivityLogRepository
	cacheService    *CacheService
	searchService   *SearchService
}

func NewPostService(
	postRepo *repository.PostRepository,
	activityLogRepo *repository.ActivityLogRepository,
	cacheService *CacheService,
	searchService *SearchService,
) *PostService {
	return &PostService{
		postRepo:        postRepo,
		activityLogRepo: activityLogRepo,
		cacheService:    cacheService,
		searchService:   searchService,
	}
}

func (s *PostService) CreatePost(req *models.PostRequest) (*models.Post, error) {
	post := &models.Post{
		Title:   req.Title,
		Content: req.Content,
		Tags:    models.StringArray(req.Tags),
	}

	// Begin transaction
	tx, err := s.postRepo.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Create post
	if err = s.postRepo.Create(tx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Create activity log
	activityLog := &models.ActivityLog{
		Action: "new_post",
		PostID: post.ID,
	}

	if err = s.activityLogRepo.Create(tx, activityLog); err != nil {
		return nil, fmt.Errorf("failed to create activity log: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Index in Elasticsearch (async, don't fail if this fails)
	go func() {
		if indexErr := s.searchService.IndexPost(post); indexErr != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to index post in Elasticsearch: %v\n", indexErr)
		}
	}()

	return post, nil
}

func (s *PostService) GetPost(id int64) (*models.PostResponse, error) {
	// Try to get from cache first (Cache-Aside pattern)
	post, err := s.cacheService.GetPost(id)
	if err != nil {
		// Log error but continue to database
		fmt.Printf("Cache error: %v\n", err)
	}

	if post == nil {
		// Cache miss - get from database
		post, err = s.postRepo.GetByID(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("post not found")
			}
			return nil, fmt.Errorf("failed to get post: %w", err)
		}

		// Set in cache
		if cacheErr := s.cacheService.SetPost(post); cacheErr != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to cache post: %v\n", cacheErr)
		}
	}

	// Get related posts (bonus feature)
	relatedPosts, err := s.searchService.SearchRelatedPosts([]string(post.Tags), post.ID, 5)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to get related posts: %v\n", err)
		relatedPosts = []models.Post{}
	}

	response := &models.PostResponse{
		Post:         post,
		RelatedPosts: relatedPosts,
	}

	return response, nil
}

func (s *PostService) UpdatePost(id int64, req *models.PostRequest) (*models.Post, error) {
	post := &models.Post{
		ID:      id,
		Title:   req.Title,
		Content: req.Content,
		Tags:    models.StringArray(req.Tags),
	}

	// Begin transaction
	tx, err := s.postRepo.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Update post
	if err = s.postRepo.Update(tx, post); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// Create activity log
	activityLog := &models.ActivityLog{
		Action: "update_post",
		PostID: post.ID,
	}

	if err = s.activityLogRepo.Create(tx, activityLog); err != nil {
		return nil, fmt.Errorf("failed to create activity log: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache
	if cacheErr := s.cacheService.DeletePost(id); cacheErr != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to invalidate cache: %v\n", cacheErr)
	}

	// Update index in Elasticsearch (async)
	go func() {
		if indexErr := s.searchService.IndexPost(post); indexErr != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to update post in Elasticsearch: %v\n", indexErr)
		}
	}()

	return post, nil
}

func (s *PostService) SearchByTag(tag string, limit, offset int) ([]models.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	posts, err := s.postRepo.SearchByTag(tag, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts by tag: %w", err)
	}

	return posts, nil
}

func (s *PostService) SearchPosts(query string, limit int) ([]models.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	posts, err := s.searchService.SearchPosts(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}

	return posts, nil
}