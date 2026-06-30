package models

import "time"

type Category struct {
	ID        uint   `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

type Product struct {
	ID             uint       `json:"id"`
	CategoryID     uint       `json:"category_id"`
	SKU            string     `json:"sku"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	ImageURL       string     `json:"image_url"`
	RegularPrice   float64    `json:"regular_price"`
	SalePrice      float64    `json:"sale_price"`
	ESXItemName    string     `json:"esx_item_name"`
	ESXItemCount   int        `json:"esx_item_count"`
	StockLimit     int        `json:"stock_limit"`
	StockSold      int        `json:"stock_sold"`
	MaxLimitPerID  int        `json:"max_limit_per_id"`
	ExpiryDate     *time.Time `json:"expiry_date"`
	IsFeatured     bool       `json:"is_featured"`
	IsActive       bool       `json:"is_active"`
	DiscountPct    float64    `json:"discount_pct,omitempty"`
	StockRemaining int        `json:"stock_remaining,omitempty"`
}

type Banner struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	ImageURL  string `json:"image_url"`
	LinkURL   string `json:"link_url"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

type Package struct {
	ID             uint       `json:"id"`
	SKU            string     `json:"sku"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	ImageURL       string     `json:"image_url"`
	RegularPrice   float64    `json:"regular_price"`
	SalePrice      float64    `json:"sale_price"`
	StockLimit     int        `json:"stock_limit"`
	StockSold      int        `json:"stock_sold"`
	MaxLimitPerID  int        `json:"max_limit_per_id"`
	ExpiryDate     *time.Time `json:"expiry_date"`
	IsFeatured     bool       `json:"is_featured"`
	IsActive       bool       `json:"is_active"`
	Items          []PackageItem `json:"items,omitempty"`
	DiscountPct    float64    `json:"discount_pct,omitempty"`
}

type PackageItem struct {
	ESXItemName  string `json:"esx_item_name"`
	ESXItemCount int    `json:"esx_item_count"`
}

type CartItem struct {
	Type     string  `json:"type"`
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type Cart struct {
	Items []CartItem `json:"items"`
	Total float64    `json:"total"`
}

type UserProfile struct {
	DiscordID            string  `json:"discord_id"`
	Identifier           string  `json:"identifier"`
	DisplayName          string  `json:"display_name"`
	MonthlyAccumulation  float64 `json:"monthly_accumulation"`
	RedeemPoints         float64 `json:"redeem_points"`
	TotalTopupAmount     float64 `json:"total_topup_amount"`
	TopupCount           int     `json:"topup_count"`
}

type MilestoneTier struct {
	ID              uint    `json:"id"`
	EventID         uint    `json:"event_id"`
	TierLevel       int     `json:"tier_level"`
	ThresholdAmount float64 `json:"threshold_amount"`
	RewardName      string  `json:"reward_name"`
	ESXItemName     string  `json:"esx_item_name"`
	ESXItemCount    int     `json:"esx_item_count"`
	Claimed         bool    `json:"claimed"`
	Eligible        bool    `json:"eligible"`
}

type RedeemItem struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	ImageURL      string  `json:"image_url"`
	PointCost     float64 `json:"point_cost"`
	ESXItemName   string  `json:"esx_item_name"`
	ESXItemCount  int     `json:"esx_item_count"`
	StockLimit    int     `json:"stock_limit"`
	StockRedeemed int     `json:"stock_redeemed"`
	IsActive      bool    `json:"is_active"`
}

type TopupTransaction struct {
	ID            uint      `json:"id"`
	TxRef         string    `json:"tx_ref"`
	DiscordID     string    `json:"discord_id"`
	Identifier    string    `json:"identifier"`
	Amount        float64   `json:"amount"`
	PointsEarned  float64   `json:"points_earned"`
	PaymentMethod string    `json:"payment_method"`
	GatewayRef    string    `json:"gateway_ref"`
	SlipURL       string    `json:"slip_url"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type AuditLog struct {
	ID             uint      `json:"id"`
	AdminDiscordID string    `json:"admin_discord_id"`
	Action         string    `json:"action"`
	TargetType     string    `json:"target_type"`
	TargetID       string    `json:"target_id"`
	Detail         string    `json:"detail"`
	IPAddress      string    `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at"`
}

type KPIRevenue struct {
	Period string  `json:"period"`
	Amount float64 `json:"amount"`
}

type KPIPaymentMethod struct {
	Method string `json:"method"`
	Count  int    `json:"count"`
}

type KPITopSpender struct {
	DiscordID   string  `json:"discord_id"`
	DisplayName string  `json:"display_name"`
	TotalAmount float64 `json:"total_amount"`
	TopupCount  int     `json:"topup_count"`
}

type DeliveryPayload struct {
	Identifier string `json:"identifier"`
	DiscordID  string `json:"discord_id"`
	Items      []DeliveryItem `json:"items"`
	SourceType string `json:"source_type"`
	SourceRef  string `json:"source_ref"`
}

type DeliveryItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
