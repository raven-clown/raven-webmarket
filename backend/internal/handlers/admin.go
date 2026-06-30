package handlers

import (
	"encoding/json"
	"strconv"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/raven-clown/raven-webmarket/backend/internal/admin"
	"github.com/raven-clown/raven-webmarket/backend/internal/middleware"
	"github.com/raven-clown/raven-webmarket/backend/internal/response"
)


func (a *API) clientIP(r *http.Request) string {
	return middleware.ClientIP(a.cfg, r)
}

func (a *API) RegisterAdminRoutes(r chi.Router, adminAuth func(http.Handler) http.Handler, devOnly func(http.Handler) http.Handler) {
	r.Post("/api/v1/admin/auth/login", a.AdminLogin)
	r.Post("/api/v1/admin/auth/logout", a.AdminLogout)

	r.Group(func(ar chi.Router) {
		ar.Use(adminAuth)
		ar.Get("/api/v1/admin/auth/me", a.AdminMe)
		ar.Get("/api/v1/admin/health", a.AdminHealth)
		ar.Get("/api/v1/admin/monitoring/config", a.AdminGetMonitoring)
		ar.Get("/api/v1/admin/autoscale/config", a.AdminGetAutoscale)
		ar.Get("/api/v1/admin/autoscale/manifest", a.AdminAutoscaleManifest)
		ar.Get("/api/v1/admin/users/search", a.AdminSearchUser)
		ar.Get("/api/v1/admin/users/{discordID}/topups", a.AdminUserTopups)
		ar.Get("/api/v1/admin/audit-logs", a.AdminAuditLogs)
		ar.Get("/api/v1/admin/activity-logs", a.AdminActivityLogs)
		ar.Get("/api/v1/admin/purchase-logs", a.AdminPurchaseLogs)
		ar.Get("/api/v1/admin/purchase-logs/{orderRef}", a.AdminPurchaseLogDetail)
		ar.Get("/api/v1/admin/kpi/revenue", a.AdminKPIRevenue)
		ar.Get("/api/v1/admin/kpi/peak", a.AdminKPIPeak)
		ar.Get("/api/v1/admin/kpi/frequency", a.AdminKPIFrequency)
		ar.Get("/api/v1/admin/kpi/top-spenders", a.AdminTopSpenders)
		ar.Get("/api/v1/admin/catalog/products", a.AdminListProducts)
		ar.Get("/api/v1/admin/catalog/packages", a.AdminListPackages)
		ar.Get("/api/v1/admin/catalog/categories", a.AdminListCategories)
		ar.Get("/api/v1/admin/promotions", a.AdminListPromotions)
		ar.Get("/api/v1/admin/milestones/events", a.AdminListMilestones)
		ar.Get("/api/v1/admin/redeem/catalog", a.AdminListRedeem)
		ar.Get("/api/v1/admin/monthly-reset/config", a.AdminGetMonthlyReset)
		ar.Get("/api/v1/admin/permissions", a.AdminPermissionsList)
		ar.Post("/api/v1/admin/products", a.AdminUpsertProduct)
		ar.Post("/api/v1/admin/packages", a.AdminUpsertPackage)
		ar.Post("/api/v1/admin/promotions", a.AdminUpsertPromotion)
		ar.Post("/api/v1/admin/milestones/events", a.AdminUpsertMilestone)
		ar.Post("/api/v1/admin/redeem/catalog", a.AdminUpsertRedeem)
		ar.Post("/api/v1/admin/banners", a.AdminUpsertBanner)
		ar.Post("/api/v1/admin/content/posts", a.AdminUpsertSitePost)
		ar.Post("/api/v1/admin/reset-accumulations", a.AdminReset)
		ar.Put("/api/v1/admin/monthly-reset/config", a.AdminSaveMonthlyReset)
		ar.Delete("/api/v1/admin/promotions/{id}", a.AdminDeletePromotion)

		ar.Group(func(dr chi.Router) {
			dr.Use(devOnly)
			dr.Put("/api/v1/admin/monitoring/config", a.AdminSaveMonitoring)
			dr.Put("/api/v1/admin/autoscale/config", a.AdminSaveAutoscale)
			dr.Post("/api/v1/admin/cache/invalidate", a.AdminInvalidateCache)
			dr.Get("/api/v1/admin/accounts", a.AdminListAccounts)
			dr.Post("/api/v1/admin/accounts", a.AdminCreateAccount)
			dr.Put("/api/v1/admin/accounts/{id}", a.AdminUpdateAccount)
		})
	})
}

func (a *API) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	token, acc, err := a.admin.Login(r.Context(), a.cfg, req.Username, req.Password, a.clientIP(r))
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	secure := strings.HasPrefix(a.cfg.FrontendURL, "https")
	http.SetCookie(w, &http.Cookie{
		Name: "raven_admin_token", Value: token, Path: "/",
		MaxAge: 28800, HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: secure,
	})
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"token": token, "username": acc.Username, "role": acc.Role, "display_name": acc.DisplayName, "permissions": acc.Permissions,
	})
}

func (a *API) AdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "raven_admin_token", Value: "", Path: "/", MaxAge: -1})
	response.JSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (a *API) AdminMe(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	acc, err := a.admin.Me(r.Context(), actor.Username)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, acc)
}

func (a *API) AdminHealth(w http.ResponseWriter, r *http.Request) {
	report := a.admin.HealthReport(r.Context(), a.cfg, a.redis)
	response.JSON(w, http.StatusOK, report)
}

func (a *API) AdminGetMonitoring(w http.ResponseWriter, r *http.Request) {
	cfg, err := a.admin.GetMonitoringConfig(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, cfg)
}

func (a *API) AdminSaveMonitoring(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	var cfg admin.MonitoringConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.SaveMonitoringConfig(r.Context(), *actor, a.clientIP(r), cfg); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) AdminGetAutoscale(w http.ResponseWriter, r *http.Request) {
	cfg, err := a.admin.GetAutoscaleConfig(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, cfg)
}

func (a *API) AdminSaveAutoscale(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	var req struct {
		Target string                `json:"target"`
		Config admin.AutoscaleConfig `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.SaveAutoscaleConfig(r.Context(), *actor, a.clientIP(r), req.Target, req.Config); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) AdminAutoscaleManifest(w http.ResponseWriter, r *http.Request) {
	manifest, err := a.admin.AutoscaleManifest(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, manifest)
}

func (a *API) AdminActivityLogs(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	actorID := r.URL.Query().Get("actor_id")
	logs, err := a.admin.ActivityLogs(r.Context(), category, actorID, 200)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, logs)
}

func (a *API) AdminPurchaseLogs(w http.ResponseWriter, r *http.Request) {
	discordID := r.URL.Query().Get("discord_id")
	status := r.URL.Query().Get("status")
	logs, err := a.admin.PurchaseLogs(r.Context(), discordID, status, 200)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, logs)
}

func (a *API) AdminPurchaseLogDetail(w http.ResponseWriter, r *http.Request) {
	ref := chi.URLParam(r, "orderRef")
	data, err := a.admin.PurchaseLogDetail(r.Context(), ref)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, data)
}

func (a *API) AdminListAccounts(w http.ResponseWriter, r *http.Request) {
	items, err := a.admin.ListAccounts(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) AdminCreateAccount(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	var req struct {
		Username    string   `json:"username"`
		Password    string   `json:"password"`
		Role        string   `json:"role"`
		DisplayName string   `json:"display_name"`
		DiscordID   string   `json:"discord_id"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.CreateAccount(r.Context(), *actor, a.clientIP(r), req.Username, req.Password, req.Role, req.DisplayName, req.DiscordID, req.Permissions); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "created"})
}

func (a *API) AdminUpdateAccount(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 64)
	var req struct {
		Password    string   `json:"password"`
		Role        string   `json:"role"`
		IsActive    *bool    `json:"is_active"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.UpdateAccount(r.Context(), *actor, a.clientIP(r), uint(id), req.Password, req.Role, req.IsActive, req.Permissions); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
