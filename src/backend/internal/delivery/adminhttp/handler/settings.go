package handler

import (
	"encoding/json"
	"net/http"

	"github.com/J0hnLenin/67League/internal/application/ports"
	appdto "github.com/J0hnLenin/67League/internal/delivery/http/dto"
	"github.com/J0hnLenin/67League/internal/delivery/http/httputil"
	"github.com/J0hnLenin/67League/internal/domain/settings"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

type AdminSettingsHandler struct {
	settings ports.SettingsRepository
}

func NewAdminSettingsHandler(settings ports.SettingsRepository) *AdminSettingsHandler {
	return &AdminSettingsHandler{settings: settings}
}

func (h *AdminSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	all, err := h.settings.GetAll(r.Context())
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, all)
}

func (h *AdminSettingsHandler) Patch(w http.ResponseWriter, r *http.Request) {
	var req appdto.PatchSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	if len(req.Values) == 0 {
		httputil.WriteError(w, shared.ErrValidation)
		return
	}
	values := settings.SiteSettings{}
	for k, v := range req.Values {
		values[k] = v
	}
	if err := h.settings.SetMany(r.Context(), values); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"updated": len(values)})
}
