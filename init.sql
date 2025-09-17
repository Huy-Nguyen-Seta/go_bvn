CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    tags TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE activity_logs (
    id SERIAL PRIMARY KEY,
    action VARCHAR(50) NOT NULL,
    post_id INTEGER REFERENCES posts(id),
    logged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_posts_tags ON posts USING GIN(tags);

CREATE INDEX idx_posts_created_at ON posts(created_at DESC);

INSERT INTO posts (title, content, tags) VALUES
('Check 1', 'Content 1', ARRAY['go', 'Attribue', 'Tophats']),
('Attribue', 'Attribue team', ARRAY['docker', 'Attribue', 'containers']),
('Tophats', 'Learning material team', ARRAY['redis', 'Attribue', 'Tophats']),
('Aris', 'check search', ARRAY['elasticsearch', 'search', 'database']),
('Database check1', 'Check 2.', ARRAY['tophats', 'Attribue', 'performance']);