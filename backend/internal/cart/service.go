package cart

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	redisstore "github.com/raven-clown/raven-webmarket/backend/internal/redisstore"
)

type Service struct {
	redis *redisstore.Store
}

func New(redis *redisstore.Store) *Service {
	return &Service{redis: redis}
}

func cartKey(discordID string) string {
	return "cart:" + discordID
}

func (s *Service) Get(ctx context.Context, discordID string) (*models.Cart, error) {
	raw, err := s.redis.Cart.Get(ctx, cartKey(discordID)).Bytes()
	if err != nil {
		return &models.Cart{Items: []models.CartItem{}}, nil
	}
	var cart models.Cart
	if err := json.Unmarshal(raw, &cart); err != nil {
		return &models.Cart{Items: []models.CartItem{}}, nil
	}
	return &cart, nil
}

func (s *Service) Save(ctx context.Context, discordID string, cart *models.Cart) error {
	cart.Total = 0
	for _, item := range cart.Items {
		cart.Total += item.Price * float64(item.Quantity)
	}
	data, err := json.Marshal(cart)
	if err != nil {
		return err
	}
	return s.redis.Cart.Set(ctx, cartKey(discordID), data, 0).Err()
}

func (s *Service) Add(ctx context.Context, discordID string, item models.CartItem) (*models.Cart, error) {
	cart, err := s.Get(ctx, discordID)
	if err != nil {
		return nil, err
	}
	found := false
	for i, existing := range cart.Items {
		if existing.Type == item.Type && existing.ID == item.ID {
			cart.Items[i].Quantity += item.Quantity
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, item)
	}
	if err := s.Save(ctx, discordID, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *Service) Update(ctx context.Context, discordID, itemType string, id uint, quantity int) (*models.Cart, error) {
	cart, err := s.Get(ctx, discordID)
	if err != nil {
		return nil, err
	}
	newItems := make([]models.CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		if item.Type == itemType && item.ID == id {
			if quantity <= 0 {
				continue
			}
			item.Quantity = quantity
		}
		newItems = append(newItems, item)
	}
	cart.Items = newItems
	if err := s.Save(ctx, discordID, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *Service) Remove(ctx context.Context, discordID, itemType string, id uint) (*models.Cart, error) {
	return s.Update(ctx, discordID, itemType, id, 0)
}

func (s *Service) Clear(ctx context.Context, discordID string) error {
	return s.redis.Cart.Del(ctx, cartKey(discordID)).Err()
}

func (s *Service) ItemKey(itemType string, id uint) string {
	return fmt.Sprintf("%s:%d", itemType, id)
}
