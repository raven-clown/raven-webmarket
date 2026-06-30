package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/raven-clown/raven-webmarket/backend/internal/admin"
	"github.com/raven-clown/raven-webmarket/backend/internal/auth"
	"github.com/raven-clown/raven-webmarket/backend/internal/cart"
	"github.com/raven-clown/raven-webmarket/backend/internal/catalog"
	"github.com/raven-clown/raven-webmarket/backend/internal/config"
	"github.com/raven-clown/raven-webmarket/backend/internal/delivery"
	"github.com/raven-clown/raven-webmarket/backend/internal/middleware"
	"github.com/raven-clown/raven-webmarket/backend/internal/milestone"
	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	"github.com/raven-clown/raven-webmarket/backend/internal/order"
	"github.com/raven-clown/raven-webmarket/backend/internal/payment"
	"github.com/raven-clown/raven-webmarket/backend/internal/redeem"
	"github.com/raven-clown/raven-webmarket/backend/internal/response"
)

type API struct {
	cfg       *config.Config
	auth      *auth.Service
	catalog   *catalog.Service
	cart      *cart.Service
	order     *order.Service
	milestone *milestone.Service
	redeem    *redeem.Service
	payment   *payment.Service
	delivery  *delivery.Service
	admin     *admin.Service
}

func NewAPI(
	cfg *config.Config,
	authSvc *auth.Service,
	catalogSvc *catalog.Service,
	cartSvc *cart.Service,
	orderSvc *order.Service,
	milestoneSvc *milestone.Service,
	redeemSvc *redeem.Service,
	paymentSvc *payment.Service,
	deliverySvc *delivery.Service,
	adminSvc *admin.Service,
) *API {
	return &API{
		cfg: cfg, auth: authSvc, catalog: catalogSvc, cart: cartSvc,
		order: orderSvc, milestone: milestoneSvc, redeem: redeemSvc,
		payment: paymentSvc, delivery: deliverySvc, admin: adminSvc,
	}
}

func (a *API) Routes(r chi.Router, authMw, adminMw func(http.Handler) http.Handler) {
	r.Get("/healthz", a.Health)
	r.Get("/api/v1/catalog/banners", a.GetBanners)
	r.Get("/api/v1/catalog/categories", a.GetCategories)
	r.Get("/api/v1/catalog/products", a.GetProducts)
	r.Get("/api/v1/catalog/packages", a.GetPackages)
	r.Get("/api/v1/auth/discord", a.DiscordLogin)
	r.Get("/api/v1/auth/callback", a.DiscordCallback)
	r.Post("/api/v1/payments/webhook", a.PaymentWebhook)
	r.Get("/api/v1/game/mailbox/{identifier}", a.GameMailbox)
	r.Post("/api/v1/game/mailbox/{id}/claim", a.ClaimMailbox)

	r.Group(func(pr chi.Router) {
		pr.Use(authMw)
		pr.Get("/api/v1/auth/me", a.Me)
		pr.Post("/api/v1/auth/logout", a.Logout)
		pr.Get("/api/v1/cart", a.GetCart)
		pr.Post("/api/v1/cart/items", a.AddCartItem)
		pr.Put("/api/v1/cart/items", a.UpdateCartItem)
		pr.Delete("/api/v1/cart/items", a.RemoveCartItem)
		pr.Post("/api/v1/orders/checkout", a.Checkout)
		pr.Get("/api/v1/milestones", a.GetMilestones)
		pr.Post("/api/v1/milestones/claim", a.ClaimMilestone)
		pr.Get("/api/v1/redeem/catalog", a.RedeemCatalog)
		pr.Post("/api/v1/redeem", a.RedeemItem)
		pr.Post("/api/v1/payments/create", a.CreatePayment)
		pr.Post("/api/v1/payments/slip", a.UploadSlip)
		pr.Get("/api/v1/payments/history", a.PaymentHistory)
	})

	r.Group(func(ar chi.Router) {
		ar.Use(authMw)
		ar.Use(adminMw)
		ar.Get("/api/v1/admin/users/search", a.AdminSearchUser)
		ar.Get("/api/v1/admin/users/{discordID}/topups", a.AdminUserTopups)
		ar.Post("/api/v1/admin/reset-accumulations", a.AdminReset)
		ar.Get("/api/v1/admin/audit-logs", a.AdminAuditLogs)
		ar.Get("/api/v1/admin/kpi/revenue", a.AdminKPIRevenue)
		ar.Get("/api/v1/admin/kpi/peak", a.AdminKPIPeak)
		ar.Get("/api/v1/admin/kpi/frequency", a.AdminKPIFrequency)
		ar.Get("/api/v1/admin/kpi/top-spenders", a.AdminTopSpenders)
		ar.Post("/api/v1/admin/products", a.AdminUpsertProduct)
		ar.Post("/api/v1/admin/banners", a.AdminUpsertBanner)
		ar.Post("/api/v1/admin/cache/invalidate", a.AdminInvalidateCache)
	})
}

func (a *API) Health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) GetBanners(w http.ResponseWriter, r *http.Request) {
	items, err := a.catalog.GetBanners(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) GetCategories(w http.ResponseWriter, r *http.Request) {
	items, err := a.catalog.GetCategories(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) GetProducts(w http.ResponseWriter, r *http.Request) {
	catID, _ := strconv.ParseUint(r.URL.Query().Get("category_id"), 10, 64)
	search := r.URL.Query().Get("search")
	featured := r.URL.Query().Get("featured") == "1"
	items, err := a.catalog.GetProducts(r.Context(), uint(catID), search, featured)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) GetPackages(w http.ResponseWriter, r *http.Request) {
	featured := r.URL.Query().Get("featured") == "1"
	items, err := a.catalog.GetPackages(r.Context(), featured)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) DiscordLogin(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		state = "raven"
	}
	http.Redirect(w, r, a.auth.AuthURL(state), http.StatusTemporaryRedirect)
}

func (a *API) DiscordCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		response.Error(w, http.StatusBadRequest, "missing code")
		return
	}
	token, err := a.auth.HandleCallback(r.Context(), code)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	auth.WriteAuthCookie(w, a.cfg, token)
	http.Redirect(w, r, a.cfg.FrontendURL+"/shop?login=success", http.StatusTemporaryRedirect)
}

func (a *API) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	data, err := a.auth.Me(r.Context(), user.DiscordID)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	data["is_admin"] = user.IsAdmin
	response.JSON(w, http.StatusOK, data)
}

func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	a.auth.Logout(r.Context(), user.DiscordID)
	http.SetCookie(w, &http.Cookie{Name: "raven_token", Value: "", Path: "/", MaxAge: -1})
	response.JSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (a *API) GetCart(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	c, err := a.cart.Get(r.Context(), user.DiscordID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, c)
}

func (a *API) AddCartItem(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	var item models.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	c, err := a.cart.Add(r.Context(), user.DiscordID, item)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, c)
}

func (a *API) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	var req struct {
		Type     string `json:"type"`
		ID       uint   `json:"id"`
		Quantity int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	c, err := a.cart.Update(r.Context(), user.DiscordID, req.Type, req.ID, req.Quantity)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, c)
}

func (a *API) RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	var req struct {
		Type string `json:"type"`
		ID   uint   `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	c, err := a.cart.Remove(r.Context(), user.DiscordID, req.Type, req.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, c)
}

func (a *API) Checkout(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cartData, err := a.cart.Get(r.Context(), user.DiscordID)
	if err != nil || len(cartData.Items) == 0 {
		response.Error(w, http.StatusBadRequest, "cart empty")
		return
	}
	ref, err := a.order.Checkout(r.Context(), user.DiscordID, user.Identifier, cartData.Items)
	if err != nil {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}
	_ = a.cart.Clear(r.Context(), user.DiscordID)
	response.JSON(w, http.StatusOK, map[string]string{"order_ref": ref})
}

func (a *API) GetMilestones(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	tiers, accumulation, err := a.milestone.GetTiers(r.Context(), user.DiscordID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"accumulation": accumulation,
		"tiers":        tiers,
	})
}

func (a *API) ClaimMilestone(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	var req struct {
		TierID uint `json:"tier_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.milestone.Claim(r.Context(), user.DiscordID, user.Identifier, req.TierID); err != nil {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "claimed"})
}

func (a *API) RedeemCatalog(w http.ResponseWriter, r *http.Request) {
	items, err := a.redeem.Catalog(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) RedeemItem(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	var req struct {
		CatalogID uint `json:"catalog_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.redeem.Redeem(r.Context(), user.DiscordID, user.Identifier, req.CatalogID); err != nil {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "redeemed"})
}

func (a *API) CreatePayment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	var req struct {
		Amount        float64 `json:"amount"`
		PaymentMethod string  `json:"payment_method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	ref, err := a.payment.CreatePending(r.Context(), user.DiscordID, user.Identifier, req.PaymentMethod, req.Amount)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"tx_ref": ref})
}

func (a *API) UploadSlip(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	var req struct {
		TxRef  string `json:"tx_ref"`
		Slip   string `json:"slip_base64"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	url, err := a.payment.UploadSlip(r.Context(), req.TxRef, user.DiscordID, req.Slip)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"slip_url": url})
}

func (a *API) PaymentHistory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	items, err := a.payment.GetHistory(r.Context(), user.DiscordID, 50)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) PaymentWebhook(w http.ResponseWriter, r *http.Request) {
	var payload payment.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if payload.Ref == "" {
		response.Error(w, http.StatusBadRequest, "missing ref")
		return
	}
	if err := a.payment.ProcessWebhook(r.Context(), payload); err != nil {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "processed"})
}

func (a *API) GameMailbox(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")
	items, err := a.delivery.PendingMailbox(r.Context(), identifier)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) ClaimMailbox(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 64)
	var req struct {
		Identifier string `json:"identifier"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := a.delivery.ClaimMailbox(r.Context(), id, req.Identifier); err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "claimed"})
}

func (a *API) AdminSearchUser(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	users, err := a.admin.SearchUser(r.Context(), q)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, users)
}

func (a *API) AdminUserTopups(w http.ResponseWriter, r *http.Request) {
	discordID := chi.URLParam(r, "discordID")
	items, err := a.admin.UserTopups(r.Context(), discordID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) AdminReset(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if err := a.admin.ResetAccumulations(r.Context(), user.DiscordID, r.RemoteAddr); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "reset"})
}

func (a *API) AdminAuditLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := a.admin.AuditLogs(r.Context(), 100)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, logs)
}

func (a *API) AdminKPIRevenue(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "daily"
	}
	items, err := a.admin.RevenueOverview(r.Context(), period)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) AdminKPIPeak(w http.ResponseWriter, r *http.Request) {
	amount, peakTime, err := a.admin.PeakTopup(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"amount": amount, "peak_at": peakTime,
	})
}

func (a *API) AdminKPIFrequency(w http.ResponseWriter, r *http.Request) {
	items, err := a.admin.TransactionFrequency(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) AdminTopSpenders(w http.ResponseWriter, r *http.Request) {
	items, err := a.admin.TopSpenders(r.Context(), 20)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) AdminUpsertProduct(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.UpsertProduct(r.Context(), user.DiscordID, r.RemoteAddr, data); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.catalog.InvalidateCache(r.Context())
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) AdminUpsertBanner(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.UpsertBanner(r.Context(), user.DiscordID, r.RemoteAddr, data); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.catalog.InvalidateCache(r.Context())
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) AdminInvalidateCache(w http.ResponseWriter, r *http.Request) {
	a.catalog.InvalidateCache(r.Context())
	response.JSON(w, http.StatusOK, map[string]string{"status": "invalidated"})
}
