package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"myim/apps/content-service/model"
)

var ErrNotFound = errors.New("content not found") // Content资源不存在

func (d *Dao) CreateContent(ctx context.Context, content *model.Content) error {
	media, _ := json.Marshal(content.MediaURLs)
	const query = `
		INSERT INTO contents
			(content_id, author_user_id, text, media_urls, status,
			 like_count, comment_count, created_at, published_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, 0, 0, $6, $7, $8)
	`
	_, err := d.db.ExecContext(ctx, query, content.ContentID, content.AuthorUserID,
		content.Text, media, content.Status, content.CreatedAt, content.PublishedAt, content.UpdatedAt)
	return err
}

func (d *Dao) GetContent(ctx context.Context, contentID string) (*model.Content, error) {
	const query = `
		SELECT
			content_id, author_user_id, text, media_urls, status,
			like_count, comment_count, created_at, published_at, updated_at
		FROM contents
		WHERE content_id = $1
	`
	return scanContent(d.db.QueryRowContext(ctx, query, contentID))
}

func scanContent(row interface{ Scan(...any) error }) (*model.Content, error) {
	content := new(model.Content)
	var media []byte
	if err := row.Scan(&content.ContentID, &content.AuthorUserID, &content.Text, &media,
		&content.Status, &content.LikeCount, &content.CommentCount, &content.CreatedAt,
		&content.PublishedAt, &content.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(media, &content.MediaURLs)
	return content, nil
}

func (d *Dao) ListContents(ctx context.Context, userID string, page, size int) ([]*model.Content, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	const query = `
		SELECT
			content_id, author_user_id, text, media_urls, status,
			like_count, comment_count, created_at, published_at, updated_at
		FROM contents
		WHERE author_user_id = $1 AND status <> 3
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := d.db.QueryContext(ctx, query, userID, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contents []*model.Content
	for rows.Next() {
		content, scanErr := scanContent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		contents = append(contents, content)
	}
	return contents, rows.Err()
}

func (d *Dao) PublishContent(ctx context.Context, contentID string, now time.Time) (*model.Content, error) {
	const query = `
		UPDATE contents
		SET status = 2, published_at = $1, updated_at = $1
		WHERE content_id = $2 AND status = 1
	`
	result, err := d.db.ExecContext(ctx, query, now, contentID)
	if err != nil {
		return nil, err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, ErrNotFound
	}
	return d.GetContent(ctx, contentID)
}

func (d *Dao) DeleteContent(ctx context.Context, contentID string) error {
	const query = `
		UPDATE contents
		SET status = 3, updated_at = $1
		WHERE content_id = $2 AND status <> 3
	`
	result, err := d.db.ExecContext(ctx, query, time.Now(), contentID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *Dao) ToggleLike(ctx context.Context, contentID, userID string, liked bool) (int64, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if liked {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO content_likes (content_id, user_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, contentID, userID)
	} else {
		_, err = tx.ExecContext(ctx, `
			DELETE FROM content_likes
			WHERE content_id = $1 AND user_id = $2
		`, contentID, userID)
	}
	if err != nil {
		return 0, err
	}
	var count int64
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM content_likes
		WHERE content_id = $1
	`, contentID).Scan(&count)
	if err == nil {
		_, err = tx.ExecContext(ctx, `
			UPDATE contents
			SET like_count = $1, updated_at = $2
			WHERE content_id = $3
		`, count, time.Now(), contentID)
	}
	if err == nil {
		err = tx.Commit()
	}
	return count, err
}

func (d *Dao) CreateComment(ctx context.Context, comment *model.Comment) error {
	const insertQuery = `
		INSERT INTO content_comments
			(comment_id, content_id, author_user_id, parent_id, text, status)
		VALUES
			($1, $2, $3, $4, $5, 1)
	`
	_, err := d.db.ExecContext(ctx, insertQuery, comment.CommentID, comment.ContentID,
		comment.AuthorUserID, comment.ParentID, comment.Text)
	if err != nil {
		return err
	}
	_, err = d.db.ExecContext(ctx, `
		UPDATE contents
		SET comment_count = comment_count + 1, updated_at = $1
		WHERE content_id = $2
	`, time.Now(), comment.ContentID)
	return err
}

func (d *Dao) ListComments(ctx context.Context, contentID string, page, size int) ([]*model.Comment, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	const query = `
		SELECT
			comment_id, content_id, author_user_id, parent_id,
			text, status, created_at
		FROM content_comments
		WHERE content_id = $1 AND status = 1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := d.db.QueryContext(ctx, query, contentID, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*model.Comment
	for rows.Next() {
		comment := new(model.Comment)
		if err = rows.Scan(&comment.CommentID, &comment.ContentID, &comment.AuthorUserID,
			&comment.ParentID, &comment.Text, &comment.Status, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

func (d *Dao) DeleteComment(ctx context.Context, commentID, authorID string) error {
	result, err := d.db.ExecContext(ctx, `
		UPDATE content_comments
		SET status = 2
		WHERE comment_id = $1 AND author_user_id = $2 AND status = 1
	`, commentID, authorID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	_, err = d.db.ExecContext(ctx, `
		UPDATE contents
		SET comment_count = GREATEST(comment_count - 1, 0), updated_at = $1
		WHERE content_id = (SELECT content_id FROM content_comments WHERE comment_id = $2)
	`, time.Now(), commentID)
	return err
}

func (d *Dao) Follow(ctx context.Context, followerID, followeeID string, active bool) error {
	status := 2
	if active {
		status = 1
	}
	_, err := d.db.ExecContext(ctx, `
		INSERT INTO follows
			(follower_user_id, followee_user_id, status)
		VALUES
			($1, $2, $3)
		ON CONFLICT (follower_user_id, followee_user_id)
		DO UPDATE SET status = EXCLUDED.status, updated_at = CURRENT_TIMESTAMP
	`, followerID, followeeID, status)
	return err
}
