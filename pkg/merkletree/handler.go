package merkletree

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

// Handler handles incoming requests for merkletree service
type Handler struct {
	service *Service
}

// NewHandler returns a new handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// AddHandler handlers requests to add items to the tree
func (h *Handler) AddHandler(w http.ResponseWriter, r *http.Request) {
	var p = struct {
		Credential string `json:"credential"`
		Sender     string `json:"sender"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.service.AddEntry(p.Credential, p.Sender)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "ok")
}

// RevokeHandler handles requests to revoke a claim
func (h *Handler) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	var p = struct {
		Credential string `json:"credential"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.service.RevokeEntry(p.Credential)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "ok")
}

// GetProofHandler returns a proof for a given credential
func (h *Handler) GetProofHandler(w http.ResponseWriter, r *http.Request) {
	credential := chi.URLParam(r, "credential")
	proof, err := h.service.GenerateProof(credential)
	if err != nil {
		// TODO(walfly): make this more nuanced in the way it chooses a status code
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	js, err := json.Marshal(proof)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(js)
}
