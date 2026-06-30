package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/raven-clown/raven-webmarket/backend/internal/middleware"
	"github.com/raven-clown/raven-webmarket/backend/internal/response"
)

func (a *API) GetSitePosts(w http.ResponseWriter, r *http.Request) {
	postType := r.URL.Query().Get("type")
	placement := r.URL.Query().Get("placement")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.cms.ListPosts(r.Context(), postType, placement, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) GetForumThreads(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.cms.ListThreads(r.Context(), limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (a *API) GetForumThread(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	thread, replies, err := a.cms.GetThread(r.Context(), uint(id))
	if err != nil {
		response.Error(w, http.StatusNotFound, "thread not found")
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"thread":  thread,
		"replies": replies,
	})
}

func (a *API) CreateForumThread(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "login required")
		return
	}
	var req struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	profile, _ := a.auth.Me(r.Context(), user.DiscordID)
	authorName := user.DiscordID
	if name, ok := profile["display_name"].(string); ok && name != "" {
		authorName = name
	}
	id, err := a.cms.CreateThread(r.Context(), user.DiscordID, authorName, req.Title, req.Body)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (a *API) CreateForumReply(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "login required")
		return
	}
	threadID, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	profile, _ := a.auth.Me(r.Context(), user.DiscordID)
	authorName := user.DiscordID
	if name, ok := profile["display_name"].(string); ok && name != "" {
		authorName = name
	}
	if err := a.cms.CreateReply(r.Context(), uint(threadID), user.DiscordID, authorName, req.Body); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, map[string]string{"status": "posted"})
}

func (a *API) AdminUpsertSitePost(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetAdmin(r)
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.cms.UpsertPost(r.Context(), data); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	a.admin.LogDetailed(r.Context(), actor.DiscordIDValue(), actor.Username, actor.Role, "cms", "upsert_site_post", "site_post", strMap(data, "title_en"), "{}", a.clientIP(r))
	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func strMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
