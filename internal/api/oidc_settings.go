package api

import (
	"encoding/json"
	"net/http"

	"github.com/lusoris/lurkarr/internal/database"
)

// OIDCSettingsHandler handles OIDC settings endpoints.
type OIDCSettingsHandler struct {
	DB Store
}

// HandleGetSettings handles GET /api/oidc/settings.
func (h *OIDCSettingsHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetOIDCSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load OIDC settings"))
		return
	}
	settings.ClientSecret = settings.MaskedClientSecret()
	writeJSON(w, http.StatusOK, settings)
}

// HandleUpdateSettings handles PUT /api/oidc/settings.
func (h *OIDCSettingsHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)

	var req database.OIDCSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	// Preserve existing client secret if the masked value was sent back.
	if req.ClientSecret == "" || (len(req.ClientSecret) >= 4 && req.ClientSecret[:4] == "****") {
		existing, err := h.DB.GetOIDCSettings(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load current settings"))
			return
		}
		req.ClientSecret = existing.ClientSecret
	}

	// Validate required fields when enabling.
	if req.Enabled {
		if req.IssuerURL == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse("issuer_url is required when OIDC is enabled"))
			return
		}
		if req.ClientID == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse("client_id is required when OIDC is enabled"))
			return
		}
		if req.RedirectURL == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse("redirect_url is required when OIDC is enabled"))
			return
		}
	}

	if err := h.DB.UpdateOIDCSettings(r.Context(), &req); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to save OIDC settings"))
		return
	}

	req.ClientSecret = req.MaskedClientSecret()
	writeJSON(w, http.StatusOK, req)
}
