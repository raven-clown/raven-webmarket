package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/raven-clown/raven-webmarket/backend/internal/admin"
	"github.com/raven-clown/raven-webmarket/backend/internal/middleware"
	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	"github.com/raven-clown/raven-webmarket/backend/internal/rbac"
	"github.com/raven-clown/raven-webmarket/backend/internal/response"
)

func (a *API) AdminListProducts(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	items, err := a.admin.ListProducts(r.Context(), *actor)
	if err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) AdminListPackages(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	items, err := a.admin.ListPackages(r.Context(), *actor)
	if err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) AdminListCategories(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	items, err := a.admin.ListCategories(r.Context(), *actor)
	if err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) AdminUpsertProduct(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.UpsertProduct(r.Context(), *actor, a.clientIP(r), data); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	a.catalog.InvalidateCache(r.Context())
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) AdminUpsertPackage(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	var req struct {
		Package map[string]interface{}   `json:"package"`
		Items   []map[string]interface{} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.UpsertPackage(r.Context(), *actor, a.clientIP(r), req.Package, req.Items); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	a.catalog.InvalidateCache(r.Context())
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) AdminListPromotions(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	items, err := a.admin.ListPromotions(r.Context(), *actor)
	if err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) AdminUpsertPromotion(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	var p models.Promotion
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.UpsertPromotion(r.Context(), *actor, a.clientIP(r), p); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) AdminDeletePromotion(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 64)
	if err := a.admin.DeletePromotion(r.Context(), *actor, a.clientIP(r), uint(id)); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (a *API) AdminListMilestones(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	items, err := a.admin.ListMilestoneEvents(r.Context(), *actor)
	if err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) AdminUpsertMilestone(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	var req struct {
		Name      string                   `json:"name"`
		MonthYear string                   `json:"month_year"`
		IsActive  bool                     `json:"is_active"`
		Tiers     []map[string]interface{} `json:"tiers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.UpsertMilestoneEvent(r.Context(), *actor, a.clientIP(r), req.Name, req.MonthYear, req.IsActive, req.Tiers); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) AdminListRedeem(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	items, err := a.admin.ListRedeemCatalog(r.Context(), *actor)
	if err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) AdminUpsertRedeem(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.UpsertRedeemItem(r.Context(), *actor, a.clientIP(r), data); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) AdminGetMonthlyReset(w http.ResponseWriter, r *http.Request) {
	cfg, err := a.admin.GetMonthlyResetConfig(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, cfg)
}

func (a *API) AdminSaveMonthlyReset(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	var cfg admin.MonthlyResetConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.SaveMonthlyResetConfig(r.Context(), *actor, a.clientIP(r), cfg); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) AdminPermissionsList(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"all":     rbac.AllPermissions,
		"default": rbac.DefaultAdminPermissions,
	})
}

func (a *API) AdminGetPaymentSettings(w http.ResponseWriter, r *http.Request) {
	cfg, err := a.admin.GetPaymentSettings(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, cfg)
}

func (a *API) AdminSavePaymentSettings(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	var cfg admin.PaymentSettings
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.admin.SavePaymentSettings(r.Context(), *actor, a.clientIP(r), cfg); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}
