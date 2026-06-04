package infrastructure

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/antonio-villamarin/vat-validator/internal/vatvalidation/application"
)

// ProblemJSON is a RFC 7807 Problem Details response.
type ProblemJSON struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

// ValidationResponse is the success envelope.
type ValidationResponse struct {
	Data interface{} `json:"data"`
	Meta struct {
		ValidatedAt string `json:"validated_at"`
	} `json:"meta"`
}

// HTTPHandler handles VAT validation HTTP requests.
type HTTPHandler struct {
	vies application.VIESChecker
	log  *slog.Logger
}

// NewHTTPHandler creates a new HTTP handler for VAT validation.
func NewHTTPHandler(vies application.VIESChecker, log *slog.Logger) *HTTPHandler {
	return &HTTPHandler{
		vies: vies,
		log:  log,
	}
}

// HandleVAT handles GET /v1/vat/{number} — format validation only.
func (h *HTTPHandler) HandleVAT(w http.ResponseWriter, r *http.Request) {
	h.handleValidation(w, r, false)
}

// HandleVIES handles GET /v1/vies/{number} — format validation + VIES lookup.
func (h *HTTPHandler) HandleVIES(w http.ResponseWriter, r *http.Request) {
	h.handleValidation(w, r, true)
}

func (h *HTTPHandler) handleValidation(w http.ResponseWriter, r *http.Request, includeVIES bool) {
	if r.Method != http.MethodGet {
		writeProblemJSON(w, ProblemJSON{
			Type:     "https://vat-validator.dev/errors/method-not-allowed",
			Title:    "Method not allowed",
			Status:   http.StatusMethodNotAllowed,
			Detail:   "Only GET method is supported for this endpoint",
			Instance: r.URL.Path,
		})
		return
	}

	number := extractPathNumber(r.URL.Path)
	if number == "" {
		writeProblemJSON(w, ProblemJSON{
			Type:     "https://vat-validator.dev/errors/missing-vat-number",
			Title:    "Missing VAT number",
			Status:   http.StatusBadRequest,
			Detail:   "The VAT number path parameter is required",
			Instance: r.URL.Path,
		})
		return
	}

	h.log.Info("validating VAT number",
		slog.String("number", number),
		slog.Bool("include_vies", includeVIES),
		slog.String("path", r.URL.Path),
	)

	resp, vErr := application.ValidateVAT(
		application.ValidateVATRequest{
			VATNumber:   number,
			IncludeVIES: includeVIES,
		},
		h.vies,
	)

	if vErr != nil {
		if resp == nil {
			writeProblemJSON(w, ProblemJSON{
				Type:     vErr.Type,
				Title:    vErr.Title,
				Status:   vErr.Status,
				Detail:   vErr.Detail,
				Instance: r.URL.Path,
			})
			return
		}

		h.log.Warn("VIES lookup failed, returning partial result",
			slog.String("number", number),
			slog.String("error", vErr.Title),
		)
		writeJSON(w, vErr.Status, ValidationResponse{
			Data: resp,
			Meta: struct {
				ValidatedAt string `json:"validated_at"`
			}{ValidatedAt: resp.ValidatedAt},
		})
		return
	}

	writeJSON(w, http.StatusOK, ValidationResponse{
		Data: resp,
		Meta: struct {
			ValidatedAt string `json:"validated_at"`
		}{ValidatedAt: resp.ValidatedAt},
	})
}

// HandleHealthz handles GET /healthz
func (h *HTTPHandler) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// extractPathNumber extracts the VAT number from paths like /v1/vat/ES12345678Z
// or /v1/vies/ES12345678Z.
func extractPathNumber(path string) string {
	// Try /v1/vat/ prefix
	for _, prefix := range []string{"/v1/vat/", "/v1/vies/"} {
		if strings.HasPrefix(path, prefix) {
			number := strings.TrimPrefix(path, prefix)
			// Remove trailing slashes or :validate suffix
			if idx := strings.Index(number, ":"); idx >= 0 {
				number = number[:idx]
			}
			number = strings.TrimSuffix(number, "/")
			if number != "" {
				return number
			}
		}
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", slog.String("error", err.Error()))
	}
}

func writeProblemJSON(w http.ResponseWriter, p ProblemJSON) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	if err := json.NewEncoder(w).Encode(p); err != nil {
		slog.Error("failed to encode problem JSON", slog.String("error", err.Error()))
	}
}
