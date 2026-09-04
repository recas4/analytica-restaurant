package ingoing

import (
	"delivery-service/src/application/ports"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type HTTPHandler struct {
	service ports.DeliveryService
}

func NewHTTPHandler(service ports.DeliveryService) *HTTPHandler {
	return &HTTPHandler{service: service}
}

func (h *HTTPHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/orders/{orderID}", h.GetOrderStatus).Methods("GET", "OPTIONS")
	router.HandleFunc("/health", h.Health).Methods("GET")
}

func (h *HTTPHandler) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	orderID := vars["orderID"]

	if orderID == "" {
		http.Error(w, "order_id is required", http.StatusBadRequest)
		return
	}

	aggregate, err := h.service.GetOrderStatus(r.Context(), orderID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if aggregate == nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aggregate)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}
