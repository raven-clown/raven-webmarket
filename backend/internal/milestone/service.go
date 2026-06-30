package milestone

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/raven-clown/raven-webmarket/backend/internal/lock"
	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	redisstore "github.com/raven-clown/raven-webmarket/backend/internal/redisstore"
)

type Service struct {
	db      *sql.DB
	redis   *redisstore.Store
	deliver func(ctx context.Context, payload models.DeliveryPayload) error
}

func New(db *sql.DB, redis *redisstore.Store, deliver func(ctx context.Context, payload models.DeliveryPayload) error) *Service {
	return &Service{db: db, redis: redis, deliver: deliver}
}

func (s *Service) GetTiers(ctx context.Context, discordID string) ([]models.MilestoneTier, float64, error) {
	monthYear := time.Now().UTC().Format("2006-01")
	var eventID uint
	var accumulation float64
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM milestone_events WHERE month_year = ? AND is_active = 1`, monthYear).Scan(&eventID)
	if err == sql.ErrNoRows {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	_ = s.db.QueryRowContext(ctx, `
		SELECT monthly_accumulation FROM user_shop_profiles WHERE discord_id = ?`, discordID).Scan(&accumulation)
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.event_id, t.tier_level, t.threshold_amount, t.reward_name, t.esx_item_name, t.esx_item_count,
			(SELECT COUNT(*) FROM milestone_claims c WHERE c.tier_id = t.id AND c.discord_id = ?) AS claimed
		FROM milestone_tiers t WHERE t.event_id = ? ORDER BY t.tier_level ASC`, discordID, eventID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var tiers []models.MilestoneTier
	for rows.Next() {
		var t models.MilestoneTier
		var claimed int
		if err := rows.Scan(&t.ID, &t.EventID, &t.TierLevel, &t.ThresholdAmount, &t.RewardName,
			&t.ESXItemName, &t.ESXItemCount, &claimed); err != nil {
			return nil, 0, err
		}
		t.Claimed = claimed > 0
		t.Eligible = accumulation >= t.ThresholdAmount && !t.Claimed
		tiers = append(tiers, t)
	}
	return tiers, accumulation, nil
}

func (s *Service) Claim(ctx context.Context, discordID, identifier string, tierID uint) error {
	lockKey := "milestone:" + discordID
	return lock.WithLock(ctx, s.redis.Session, lockKey, 30*time.Second, func() error {
		var t models.MilestoneTier
		var accumulation float64
		err := s.db.QueryRowContext(ctx, `
			SELECT t.id, t.event_id, t.tier_level, t.threshold_amount, t.reward_name, t.esx_item_name, t.esx_item_count
			FROM milestone_tiers t WHERE t.id = ?`, tierID).Scan(
			&t.ID, &t.EventID, &t.TierLevel, &t.ThresholdAmount, &t.RewardName, &t.ESXItemName, &t.ESXItemCount)
		if err != nil {
			return fmt.Errorf("tier not found")
		}
		_ = s.db.QueryRowContext(ctx, `
			SELECT monthly_accumulation FROM user_shop_profiles WHERE discord_id = ?`, discordID).Scan(&accumulation)
		if accumulation < t.ThresholdAmount {
			return fmt.Errorf("threshold not reached")
		}
		var claimed int
		_ = s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM milestone_claims WHERE event_id = ? AND tier_id = ? AND discord_id = ?`,
			t.EventID, tierID, discordID).Scan(&claimed)
		if claimed > 0 {
			return fmt.Errorf("already claimed")
		}
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO milestone_claims (event_id, tier_id, discord_id, identifier) VALUES (?, ?, ?, ?)`,
			t.EventID, tierID, discordID, identifier)
		if err != nil {
			return err
		}
		ref := fmt.Sprintf("MS-%d-%s", tierID, discordID)
		payload := models.DeliveryPayload{
			Identifier: identifier,
			DiscordID:  discordID,
			Items:      []models.DeliveryItem{{Name: t.ESXItemName, Count: t.ESXItemCount}},
			SourceType: "milestone",
			SourceRef:  ref,
		}
		return s.deliver(ctx, payload)
	})
}
