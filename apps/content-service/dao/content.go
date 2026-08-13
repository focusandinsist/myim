package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"myim/apps/content-service/model"
	"time"
)

var ErrNotFound = errors.New("content not found")

func (d *Dao) CreateContent(ctx context.Context, c *model.Content) error {
	media, _ := json.Marshal(c.MediaURLs)
	_, err := d.db.ExecContext(ctx, `INSERT INTO contents (content_id,author_user_id,text,media_urls,status,created_at,published_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, c.ContentID, c.AuthorUserID, c.Text, media, c.Status, c.CreatedAt, c.PublishedAt, c.UpdatedAt)
	return err
}

func (d *Dao) GetContent(ctx context.Context, id string) (*model.Content, error) {
	row := d.db.QueryRowContext(ctx, `SELECT content_id,author_user_id,text,media_urls,status,like_count,comment_count,created_at,published_at,updated_at FROM contents WHERE content_id=$1`, id)
	return scanContent(row)
}

func scanContent(row interface{ Scan(...any) error }) (*model.Content, error) {
	c := new(model.Content)
	var media []byte
	if err := row.Scan(&c.ContentID, &c.AuthorUserID, &c.Text, &media, &c.Status, &c.LikeCount, &c.CommentCount, &c.CreatedAt, &c.PublishedAt, &c.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(media, &c.MediaURLs)
	return c, nil
}

func (d *Dao) ListContents(ctx context.Context, userID string, page, size int) ([]*model.Content, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	rows, err := d.db.QueryContext(ctx, `SELECT content_id,author_user_id,text,media_urls,status,like_count,comment_count,created_at,published_at,updated_at FROM contents WHERE author_user_id=$1 AND status<>3 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Content
	for rows.Next() {
		c, e := scanContent(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (d *Dao) PublishContent(ctx context.Context, id string, now time.Time) (*model.Content, error) {
	r, err := d.db.ExecContext(ctx, `UPDATE contents SET status=2,published_at=$1,updated_at=$1 WHERE content_id=$2 AND status=1`, now, id)
	if err != nil {
		return nil, err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return nil, ErrNotFound
	}
	return d.GetContent(ctx, id)
}

func (d *Dao) DeleteContent(ctx context.Context, id string) error {
	r, e := d.db.ExecContext(ctx, `UPDATE contents SET status=3,updated_at=$1 WHERE content_id=$2 AND status<>3`, time.Now(), id)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *Dao) ToggleLike(ctx context.Context, id, user string, liked bool) (int64, error) {
	tx, e := d.db.BeginTx(ctx, nil)
	if e != nil {
		return 0, e
	}
	defer tx.Rollback()
	if liked {
		_, e = tx.ExecContext(ctx, `INSERT INTO content_likes(content_id,user_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, id, user)
	} else {
		_, e = tx.ExecContext(ctx, `DELETE FROM content_likes WHERE content_id=$1 AND user_id=$2`, id, user)
	}
	if e != nil {
		return 0, e
	}
	var count int64
	e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM content_likes WHERE content_id=$1`, id).Scan(&count)
	if e == nil {
		_, e = tx.ExecContext(ctx, `UPDATE contents SET like_count=$1,updated_at=$2 WHERE content_id=$3`, count, time.Now(), id)
	}
	if e == nil {
		e = tx.Commit()
	}
	return count, e
}

func (d *Dao) CreateComment(ctx context.Context, c *model.Comment) error {
	_, e := d.db.ExecContext(ctx, `INSERT INTO content_comments(comment_id,content_id,author_user_id,parent_id,text,status) VALUES($1,$2,$3,$4,$5,1)`, c.CommentID, c.ContentID, c.AuthorUserID, c.ParentID, c.Text)
	if e == nil {
		_, e = d.db.ExecContext(ctx, `UPDATE contents SET comment_count=comment_count+1,updated_at=$1 WHERE content_id=$2`, time.Now(), c.ContentID)
	}
	return e
}

func (d *Dao) ListComments(ctx context.Context, id string, page, size int) ([]*model.Comment, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	rows, e := d.db.QueryContext(ctx, `SELECT comment_id,content_id,author_user_id,parent_id,text,status,created_at FROM content_comments WHERE content_id=$1 AND status=1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`, id, size, (page-1)*size)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []*model.Comment
	for rows.Next() {
		c := new(model.Comment)
		if e = rows.Scan(&c.CommentID, &c.ContentID, &c.AuthorUserID, &c.ParentID, &c.Text, &c.Status, &c.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (d *Dao) DeleteComment(ctx context.Context, id, author string) error {
	result, err := d.db.ExecContext(ctx, `UPDATE content_comments SET status=2 WHERE comment_id=$1 AND author_user_id=$2 AND status=1`, id, author)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	_, err = d.db.ExecContext(ctx, `UPDATE contents SET comment_count=GREATEST(comment_count-1, 0), updated_at=$1 WHERE content_id=(SELECT content_id FROM content_comments WHERE comment_id=$2)`, time.Now(), id)
	return err
}

func (d *Dao) Follow(ctx context.Context, from, to string, active bool) error {
	_, e := d.db.ExecContext(ctx, `INSERT INTO follows(follower_user_id,followee_user_id,status) VALUES($1,$2,$3) ON CONFLICT(follower_user_id,followee_user_id) DO UPDATE SET status=EXCLUDED.status,updated_at=CURRENT_TIMESTAMP`, from, to, map[bool]int{false: 2, true: 1}[active])
	return e
}
