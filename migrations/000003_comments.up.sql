-- Создание таблицы комментариев
CREATE TABLE comments (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL CHECK (LENGTH(content) <= 2000),
    path TEXT NOT NULL, -- Materialized path для быстрого поиска
    level INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для комментариев
CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_author_id ON comments(author_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);
CREATE INDEX idx_comments_path ON comments(path);
CREATE INDEX idx_comments_level ON comments(level);
CREATE INDEX idx_comments_created_at ON comments(created_at);
-- Составные индексы для оптимизации запросов
CREATE INDEX idx_comments_post_parent_created ON comments(post_id, parent_id, created_at);
CREATE INDEX idx_comments_path_created ON comments(path, created_at);