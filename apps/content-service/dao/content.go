package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"myim/apps/content-service/model"
	contentdb "myim/db/content"
)

var ErrNotFound = errors.New("content not found")

func (d *Dao) CreateContent(ctx context.Context, content *model.Content) error {
	media, err := json.Marshal(content.MediaURLs)
	if err != nil {
		return err
	}
	return d.queries.CreateContent(ctx, contentdb.CreateContentParams{ContentID: content.ContentID, AuthorUserID: content.AuthorUserID, Text: content.Text, MediaUrls: media, Status: int16(content.Status), CreatedAt: content.CreatedAt, PublishedAt: nullTime(content.PublishedAt), UpdatedAt: content.UpdatedAt})
}

func (d *Dao) GetContent(ctx context.Context, contentID string) (*model.Content, error) {
	row, err := d.queries.GetContent(ctx, contentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return contentFromRow(row), nil
}

func (d *Dao) ListContents(ctx context.Context, userID string, page, size int) ([]*model.Content, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	rows, err := d.queries.ListContents(ctx, contentdb.ListContentsParams{AuthorUserID: userID, Limit: int32(size), Offset: int32((page - 1) * size)})
	if err != nil {
		return nil, err
	}
	result := make([]*model.Content, 0, len(rows))
	for _, row := range rows {
		result = append(result, contentFromRow(row))
	}
	return result, nil
}

func (d *Dao) PublishContent(ctx context.Context, contentID string, now time.Time) (*model.Content, error) {
	rows, err := d.queries.PublishContent(ctx, contentdb.PublishContentParams{PublishedAt: sql.NullTime{Time: now, Valid: true}, ContentID: contentID})
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, ErrNotFound
	}
	return d.GetContent(ctx, contentID)
}

func (d *Dao) DeleteContent(ctx context.Context, contentID string) error {
	rows, err := d.queries.DeleteContent(ctx, contentdb.DeleteContentParams{UpdatedAt: time.Now(), ContentID: contentID})
	if err != nil {
		return err
	}
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
	queries := d.queries.WithTx(tx)
	if liked {
		err = queries.AddLike(ctx, contentdb.AddLikeParams{ContentID: contentID, UserID: userID})
	} else {
		err = queries.RemoveLike(ctx, contentdb.RemoveLikeParams{ContentID: contentID, UserID: userID})
	}
	if err != nil {
		return 0, err
	}
	count, err := queries.CountLikes(ctx, contentID)
	if err != nil {
		return 0, err
	}
	if err = queries.UpdateLikeCount(ctx, contentdb.UpdateLikeCountParams{LikeCount: count, UpdatedAt: time.Now(), ContentID: contentID}); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

func (d *Dao) CreateComment(ctx context.Context, comment *model.Comment) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := d.queries.WithTx(tx)
	if err = queries.CreateComment(ctx, contentdb.CreateCommentParams{CommentID: comment.CommentID, ContentID: comment.ContentID, AuthorUserID: comment.AuthorUserID, ParentID: comment.ParentID, Text: comment.Text}); err != nil {
		return err
	}
	if err = queries.IncrementCommentCount(ctx, contentdb.IncrementCommentCountParams{UpdatedAt: time.Now(), ContentID: comment.ContentID}); err != nil {
		return err
	}
	return tx.Commit()
}

func (d *Dao) ListComments(ctx context.Context, contentID string, page, size int) ([]*model.Comment, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	rows, err := d.queries.ListComments(ctx, contentdb.ListCommentsParams{ContentID: contentID, Limit: int32(size), Offset: int32((page - 1) * size)})
	if err != nil {
		return nil, err
	}
	result := make([]*model.Comment, 0, len(rows))
	for _, row := range rows {
		result = append(result, &model.Comment{CommentID: row.CommentID, ContentID: row.ContentID, AuthorUserID: row.AuthorUserID, ParentID: row.ParentID, Text: row.Text, Status: int32(row.Status), CreatedAt: row.CreatedAt})
	}
	return result, nil
}

func (d *Dao) DeleteComment(ctx context.Context, commentID, authorID string) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := d.queries.WithTx(tx)
	rows, err := queries.DeleteComment(ctx, contentdb.DeleteCommentParams{CommentID: commentID, AuthorUserID: authorID})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	if err = queries.DecrementCommentCount(ctx, contentdb.DecrementCommentCountParams{UpdatedAt: time.Now(), CommentID: commentID}); err != nil {
		return err
	}
	return tx.Commit()
}

func nullTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *value, Valid: true}
}

func contentFromRow(row contentdb.Content) *model.Content {
	content := &model.Content{ContentID: row.ContentID, AuthorUserID: row.AuthorUserID, Text: row.Text, Status: int32(row.Status), LikeCount: row.LikeCount, CommentCount: row.CommentCount, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
	_ = json.Unmarshal(row.MediaUrls, &content.MediaURLs)
	if row.PublishedAt.Valid {
		published := row.PublishedAt.Time
		content.PublishedAt = &published
	}
	return content
}
