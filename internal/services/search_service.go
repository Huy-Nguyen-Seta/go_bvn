package services

import (
	"blog-api/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type SearchService struct {
	client *elasticsearch.Client
	index  string
}

func NewSearchService(client *elasticsearch.Client) *SearchService {
	service := &SearchService{
		client: client,
		index:  "posts",
	}

	// Create index if it doesn't exist
	service.createIndex()

	return service
}

func (s *SearchService) createIndex() {
	mapping := `{
		"mappings": {
			"properties": {
				"id": {"type": "long"},
				"title": {"type": "text", "analyzer": "standard"},
				"content": {"type": "text", "analyzer": "standard"},
				"tags": {"type": "keyword"},
				"created_at": {"type": "date"},
				"updated_at": {"type": "date"}
			}
		}
	}`

	req := esapi.IndicesCreateRequest{
		Index: s.index,
		Body:  strings.NewReader(mapping),
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		log.Printf("Error creating index: %v", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() && !strings.Contains(res.String(), "resource_already_exists_exception") {
		log.Printf("Error creating index: %s", res.String())
	}
}

func (s *SearchService) IndexPost(post *models.Post) error {
	doc := map[string]interface{}{
		"id":         post.ID,
		"title":      post.Title,
		"content":    post.Content,
		"tags":       post.Tags,
		"created_at": post.CreatedAt,
		"updated_at": post.UpdatedAt,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      s.index,
		DocumentID: fmt.Sprintf("%d", post.ID),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}

	return nil
}

func (s *SearchService) SearchPosts(query string, size int) ([]models.Post, error) {
	var buf bytes.Buffer
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"title", "content"},
				"type":   "best_fields",
			},
		},
		"size": size,
		"sort": []map[string]interface{}{
			{"created_at": map[string]string{"order": "desc"}},
		},
	}

	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex(s.index),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits, ok := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		return []models.Post{}, nil
	}

	var posts []models.Post
	for _, hit := range hits {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})

		var post models.Post
		sourceBytes, _ := json.Marshal(source)
		json.Unmarshal(sourceBytes, &post)
		
		posts = append(posts, post)
	}

	return posts, nil
}

func (s *SearchService) SearchRelatedPosts(tags []string, excludeID int64, size int) ([]models.Post, error) {
	if len(tags) == 0 {
		return []models.Post{}, nil
	}

	var buf bytes.Buffer
	shouldClauses := make([]map[string]interface{}, len(tags))
	for i, tag := range tags {
		shouldClauses[i] = map[string]interface{}{
			"term": map[string]interface{}{
				"tags": tag,
			},
		}
	}

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": shouldClauses,
				"must_not": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"id": excludeID,
						},
					},
				},
				"minimum_should_match": 1,
			},
		},
		"size": size,
		"sort": []map[string]interface{}{
			{"created_at": map[string]string{"order": "desc"}},
		},
	}

	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex(s.index),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching related posts: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits, ok := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		return []models.Post{}, nil
	}

	var posts []models.Post
	for _, hit := range hits {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})

		var post models.Post
		sourceBytes, _ := json.Marshal(source)
		json.Unmarshal(sourceBytes, &post)
		
		posts = append(posts, post)
	}

	return posts, nil
}