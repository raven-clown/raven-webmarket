package delivery

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/raven-clown/raven-webmarket/backend/internal/config"
	"github.com/raven-clown/raven-webmarket/backend/internal/models"
)

type Service struct {
	cfg *config.Config
	db  *sql.DB
}

func New(cfg *config.Config, db *sql.DB) *Service {
	return &Service{cfg: cfg, db: db}
}

func (s *Service) Deliver(ctx context.Context, payload models.DeliveryPayload) error {
	if s.cfg.FiveMWebhookURL == "" {
		return s.queueMailbox(ctx, payload)
	}
	data, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.FiveMWebhookURL, bytes.NewReader(data))
	if err != nil {
		return s.queueMailbox(ctx, payload)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", s.cfg.FiveMWebhookSecret)
	if s.cfg.FiveMAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.FiveMAPIKey)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		return s.queueMailbox(ctx, payload)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return nil
}

func (s *Service) queueMailbox(ctx context.Context, payload models.DeliveryPayload) error {
	data, err := json.Marshal(payload.Items)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO web_mailbox (identifier, discord_id, payload, source_type, source_ref, status)
		VALUES (?, ?, ?, ?, ?, 'pending')`,
		payload.Identifier, payload.DiscordID, string(data), payload.SourceType, payload.SourceRef)
	return err
}

func (s *Service) PendingMailbox(ctx context.Context, identifier string) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, payload, source_type, source_ref, created_at
		FROM web_mailbox WHERE identifier = ? AND status = 'pending' ORDER BY id ASC`, identifier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []map[string]interface{}
	for rows.Next() {
		var id uint64
		var payload, sourceType, sourceRef string
		var createdAt time.Time
		if err := rows.Scan(&id, &payload, &sourceType, &sourceRef, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"id": id, "payload": payload, "source_type": sourceType,
			"source_ref": sourceRef, "created_at": createdAt,
		})
	}
	return items, nil
}

func (s *Service) ClaimMailbox(ctx context.Context, id uint64, identifier string) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE web_mailbox SET status = 'claimed', claimed_at = UTC_TIMESTAMP()
		WHERE id = ? AND identifier = ? AND status = 'pending'`, id, identifier)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
