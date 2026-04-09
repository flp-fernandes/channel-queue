package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flp-fernandes/product-views/internal/domain"
	"github.com/flp-fernandes/product-views/internal/queue"
)

type Handler struct {
	queue *queue.EventQueue
}

func NewHandler(q *queue.EventQueue) *Handler {
	return &Handler{
		queue: q,
	}
}

type productViewRequest struct {
	ProductID int64 `json:"product_id"`
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(`{"ok":"true"}`))
}

func (h *Handler) CreateProductView(w http.ResponseWriter, r *http.Request) {

	var req productViewRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.ProductID == 0 {
		http.Error(w, "product_id is required", http.StatusBadRequest)
		return
	}

	view := domain.ProductView{
		ProductID: req.ProductID,
		ViewedAt:  time.Now().UTC(),
	}

	ok := h.queue.Enqueue(view)
	if !ok {
		http.Error(w, "queue is full", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
