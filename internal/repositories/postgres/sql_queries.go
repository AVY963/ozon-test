package postgres

const (
	UserInsertQuery = `
		INSERT INTO users (id, username, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	UserSelectByIDQuery = `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	UserSelectByUsernameQuery = `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	UserSelectByEmailQuery = `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	UserUpdateQuery = `
		UPDATE users 
		SET username = $2, email = $3, updated_at = $4 
		WHERE id = $1
	`

	UserDeleteQuery = `DELETE FROM users WHERE id = $1`

	UserExistsQuery = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`

	UserSelectByIDsQuery = `
		SELECT id, username, email, created_at, updated_at 
		FROM users 
		WHERE id = ANY($1)
		ORDER BY created_at DESC
	`
)

const (
	PostInsertQuery = `
		INSERT INTO posts (id, author_id, title, content, comments_disabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	PostSelectByIDQuery = `
		SELECT id, author_id, title, content, comments_disabled, created_at, updated_at
		FROM posts
		WHERE id = $1
	`

	PostUpdateQuery = `
		UPDATE posts
		SET title = $2, content = $3, comments_disabled = $4, updated_at = $5
		WHERE id = $1
	`

	PostDeleteQuery = `DELETE FROM posts WHERE id = $1`

	PostCountAllQuery = `SELECT COUNT(*) FROM posts`

	PostSelectAllQuery = `
		SELECT id, author_id, title, content, comments_disabled, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	PostCountByAuthorQuery = `SELECT COUNT(*) FROM posts WHERE author_id = $1`

	PostSelectByAuthorQuery = `
		SELECT id, author_id, title, content, comments_disabled, created_at, updated_at
		FROM posts
		WHERE author_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	PostExistsQuery = `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`

	PostCommentsEnabledQuery = `SELECT NOT comments_disabled FROM posts WHERE id = $1`

	PostSelectByIDsQuery = `
		SELECT id, author_id, title, content, comments_disabled, created_at, updated_at
		FROM posts
		WHERE id = ANY($1)
		ORDER BY created_at DESC
	`
)

const (
	CommentInsertQuery = `
		INSERT INTO comments (id, post_id, author_id, parent_id, content, path, level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	CommentSelectByIDQuery = `
		SELECT id, post_id, author_id, parent_id, content, path, level, created_at, updated_at
		FROM comments
		WHERE id = $1
	`

	CommentUpdateQuery = `
		UPDATE comments
		SET content = $2, updated_at = $3
		WHERE id = $1
	`

	CommentDeleteQuery = `DELETE FROM comments WHERE id = $1`

	CommentCountByPostQuery = `SELECT COUNT(*) FROM comments WHERE post_id = $1 AND parent_id IS NULL`

	CommentSelectByPostQuery = `
		SELECT id, post_id, author_id, parent_id, content, path, level, created_at, updated_at
		FROM comments
		WHERE post_id = $1 AND parent_id IS NULL
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	CommentCountByParentQuery = `SELECT COUNT(*) FROM comments WHERE parent_id = $1`

	CommentSelectByParentQuery = `
		SELECT id, post_id, author_id, parent_id, content, path, level, created_at, updated_at
		FROM comments
		WHERE parent_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	CommentSelectThreadQuery = `
		SELECT id, post_id, author_id, parent_id, content, path, level, created_at, updated_at
		FROM comments
		WHERE (path = $1 OR path LIKE $2) AND level <= $3
		ORDER BY path, created_at ASC
	`

	CommentCountByPathQuery = `SELECT COUNT(*) FROM comments WHERE path LIKE $1`

	CommentSelectByPathQuery = `
		SELECT id, post_id, author_id, parent_id, content, path, level, created_at, updated_at
		FROM comments
		WHERE path LIKE $1
		ORDER BY path, created_at ASC
		LIMIT $2 OFFSET $3
	`

	CommentExistsQuery = `SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1)`

	CommentSelectByIDsQuery = `
		SELECT id, post_id, author_id, parent_id, content, path, level, created_at, updated_at
		FROM comments
		WHERE id = ANY($1)
		ORDER BY created_at ASC
	`
)
