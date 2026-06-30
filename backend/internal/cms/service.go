package cms

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	redisstore "github.com/raven-clown/raven-webmarket/backend/internal/redisstore"
)

const cacheTTL = 3 * time.Minute

type Service struct {
	db    *sql.DB
	redis *redisstore.Store
}

func New(db *sql.DB, redis *redisstore.Store) *Service {
	return &Service{db: db, redis: redis}
}

func (s *Service) ListPosts(ctx context.Context, postType, placement string, limit int) ([]models.SitePost, error) {
	if limit <= 0 {
		limit = 50
	}
	cacheKey := fmt.Sprintf("cache:posts:%s:%s:%d", postType, placement, limit)
	if raw, err := s.redis.Session.Get(ctx, cacheKey).Bytes(); err == nil {
		var items []models.SitePost
		if json.Unmarshal(raw, &items) == nil {
			return items, nil
		}
	}
	query := `
		SELECT id, post_type, title_en, COALESCE(title_th,''), COALESCE(body_en,''), COALESCE(body_th,''),
			COALESCE(image_url,''), COALESCE(link_url,''), placement, sort_order, is_pinned, is_active,
			publish_date, created_at, updated_at
		FROM site_posts WHERE is_active = 1`
	args := []interface{}{}
	if postType != "" {
		query += " AND post_type = ?"
		args = append(args, postType)
	}
	if placement != "" && placement != "all" {
		query += " AND (placement = ? OR placement = 'all')"
		args = append(args, placement)
	}
	query += " ORDER BY is_pinned DESC, sort_order ASC, publish_date DESC, id DESC LIMIT ?"
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanPosts(rows)
	if err != nil {
		return nil, err
	}
	if data, err := json.Marshal(items); err == nil {
		s.redis.Session.Set(ctx, cacheKey, data, cacheTTL)
	}
	return items, nil
}

func (s *Service) UpsertPost(ctx context.Context, data map[string]interface{}) error {
	id, _ := data["id"].(float64)
	postType, _ := data["post_type"].(string)
	titleEn, _ := data["title_en"].(string)
	if postType == "" || titleEn == "" {
		return fmt.Errorf("post_type and title_en required")
	}
	titleTh := strField(data, "title_th")
	bodyEn := strField(data, "body_en")
	bodyTh := strField(data, "body_th")
	imageURL := strField(data, "image_url")
	linkURL := strField(data, "link_url")
	placement := strField(data, "placement")
	if placement == "" {
		placement = "home"
	}
	sortOrder := intField(data, "sort_order")
	isPinned := intField(data, "is_pinned")
	isActive := intField(data, "is_active")
	if isActive == 0 && data["is_active"] == nil {
		isActive = 1
	}
	publishDate := strField(data, "publish_date")
	if id > 0 {
		_, err := s.db.ExecContext(ctx, `
			UPDATE site_posts SET post_type=?, title_en=?, title_th=?, body_en=?, body_th=?,
				image_url=?, link_url=?, placement=?, sort_order=?, is_pinned=?, is_active=?,
				publish_date=NULLIF(?, '')
			WHERE id=?`, postType, titleEn, titleTh, bodyEn, bodyTh, imageURL, linkURL,
			placement, sortOrder, isPinned, isActive, publishDate, uint(id))
		s.invalidatePostCache(ctx)
		return err
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO site_posts (post_type, title_en, title_th, body_en, body_th, image_url, link_url,
			placement, sort_order, is_pinned, is_active, publish_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''))`,
		postType, titleEn, titleTh, bodyEn, bodyTh, imageURL, linkURL,
		placement, sortOrder, isPinned, isActive, publishDate)
	s.invalidatePostCache(ctx)
	return err
}

func (s *Service) ListThreads(ctx context.Context, limit int) ([]models.ForumThread, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, discord_id, author_name, title, body, is_pinned, is_locked, reply_count, created_at, updated_at
		FROM forum_threads ORDER BY is_pinned DESC, updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.ForumThread
	for rows.Next() {
		var t models.ForumThread
		var pinned, locked int
		if err := rows.Scan(&t.ID, &t.DiscordID, &t.AuthorName, &t.Title, &t.Body,
			&pinned, &locked, &t.ReplyCount, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.IsPinned = pinned == 1
		t.IsLocked = locked == 1
		items = append(items, t)
	}
	return items, nil
}

func (s *Service) GetThread(ctx context.Context, id uint) (*models.ForumThread, []models.ForumReply, error) {
	var t models.ForumThread
	var pinned, locked int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, discord_id, author_name, title, body, is_pinned, is_locked, reply_count, created_at, updated_at
		FROM forum_threads WHERE id = ?`, id).Scan(
		&t.ID, &t.DiscordID, &t.AuthorName, &t.Title, &t.Body,
		&pinned, &locked, &t.ReplyCount, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, nil, err
	}
	t.IsPinned = pinned == 1
	t.IsLocked = locked == 1
	replies, err := s.ListReplies(ctx, id)
	return &t, replies, err
}

func (s *Service) ListReplies(ctx context.Context, threadID uint) ([]models.ForumReply, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, thread_id, discord_id, author_name, body, created_at
		FROM forum_replies WHERE thread_id = ? ORDER BY created_at ASC`, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.ForumReply
	for rows.Next() {
		var r models.ForumReply
		if err := rows.Scan(&r.ID, &r.ThreadID, &r.DiscordID, &r.AuthorName, &r.Body, &r.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	return items, nil
}

func (s *Service) CreateThread(ctx context.Context, discordID, authorName, title, body string) (uint, error) {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if len(title) < 3 || len(body) < 5 {
		return 0, fmt.Errorf("title or body too short")
	}
	if len(title) > 512 || len(body) > 10000 {
		return 0, fmt.Errorf("content too long")
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO forum_threads (discord_id, author_name, title, body) VALUES (?, ?, ?, ?)`,
		discordID, authorName, title, body)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return uint(id), nil
}

func (s *Service) CreateReply(ctx context.Context, threadID uint, discordID, authorName, body string) error {
	body = strings.TrimSpace(body)
	if len(body) < 2 || len(body) > 5000 {
		return fmt.Errorf("reply too short or too long")
	}
	var locked int
	err := s.db.QueryRowContext(ctx, `SELECT is_locked FROM forum_threads WHERE id = ?`, threadID).Scan(&locked)
	if err != nil {
		return err
	}
	if locked == 1 {
		return fmt.Errorf("thread locked")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO forum_replies (thread_id, discord_id, author_name, body) VALUES (?, ?, ?, ?)`,
		threadID, discordID, authorName, body)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE forum_threads SET reply_count = reply_count + 1, updated_at = UTC_TIMESTAMP() WHERE id = ?`, threadID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) invalidatePostCache(ctx context.Context) {
	iter := s.redis.Session.Scan(ctx, 0, "cache:posts:*", 100).Iterator()
	for iter.Next(ctx) {
		s.redis.Session.Del(ctx, iter.Val())
	}
}

func scanPosts(rows *sql.Rows) ([]models.SitePost, error) {
	var items []models.SitePost
	for rows.Next() {
		var p models.SitePost
		var active, pinned int
		var pub sql.NullTime
		if err := rows.Scan(&p.ID, &p.PostType, &p.TitleEN, &p.TitleTH, &p.BodyEN, &p.BodyTH,
			&p.ImageURL, &p.LinkURL, &p.Placement, &p.SortOrder, &pinned, &active,
			&pub, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.IsPinned = pinned == 1
		p.IsActive = active == 1
		if pub.Valid {
			t := pub.Time
			p.PublishDate = &t
		}
		items = append(items, p)
	}
	return items, nil
}

func strField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func intField(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}
