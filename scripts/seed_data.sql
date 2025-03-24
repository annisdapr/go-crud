-- Insert data ke tabel users
INSERT INTO users (id, name, email, created_at, updated_at) VALUES
(1, 'Annisa Dwi', 'annisa@example.com', NOW(), NOW()),
(2, 'John Doe', 'john@example.com', NOW(), NOW()),
(3, 'Jane Smith', 'jane@example.com', NOW(), NOW());

-- Insert data ke tabel repositories
INSERT INTO repositories (id, user_id, name, url, ai_enabled, created_at, updated_at) VALUES
(1, 1, 'Go CRUD API', 'https://github.com/annisa/go-crud', TRUE, NOW(), NOW()),
(2, 1, 'Golang CLI Tool', 'https://github.com/annisa/go-cli', FALSE, NOW(), NOW()),
(3, 2, 'React Frontend', 'https://github.com/johndoe/react-app', TRUE, NOW(), NOW()),
(4, 3, 'Python ML Project', 'https://github.com/janesmith/ml-project', FALSE, NOW(), NOW());
