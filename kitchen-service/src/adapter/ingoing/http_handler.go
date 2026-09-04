package ingoing

import (
	"encoding/json"
	"kitchen-service/src/application/ports"
	"net/http"

	"github.com/gorilla/mux"
)

type HTTPHandler struct {
	kitchenService ports.KitchenService
}

func NewHTTPHandler(kitchenService ports.KitchenService) *HTTPHandler {
	return &HTTPHandler{
		kitchenService: kitchenService,
	}
}

func (kh *HTTPHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/kitchen/orders", kh.GetAllOrders).Methods("GET")
	router.HandleFunc("/kitchen/orders/{orderID}", kh.GetOrder).Methods("GET")
	router.HandleFunc("/kitchen/orders/{orderID}/start", kh.StartPreparation).Methods("POST")
	router.HandleFunc("/kitchen/orders/{orderID}/complete", kh.CompleteOrder).Methods("POST")
	router.HandleFunc("/kitchen/orders/{orderID}/pickup", kh.MarkPickedUpByDriver).Methods("POST")
	router.HandleFunc("/kitchen/orders/{orderID}/cancel", kh.CancelOrder).Methods("POST")
	router.HandleFunc("/kitchen/stats", kh.GetKitchenStats).Methods("GET")
	router.HandleFunc("/kitchen/process-queue", kh.ProcessQueue).Methods("POST")
	router.HandleFunc("/kitchen/dashboard", kh.ServeDashboard).Methods("GET")
}

func (kh *HTTPHandler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := kh.kitchenService.GetAllOrders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(orders)
}

func (kh *HTTPHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["orderID"]

	order, err := kh.kitchenService.GetOrderStatus(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(order)
}

func (kh *HTTPHandler) StartPreparation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["orderID"]

	err := kh.kitchenService.StartPreparation(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Preparation started",
		"orderID": orderID,
	})
}

func (kh *HTTPHandler) CompleteOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["orderID"]

	err := kh.kitchenService.CompleteOrder(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Order completed",
		"orderID": orderID,
	})
}

func (kh *HTTPHandler) MarkPickedUpByDriver(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["orderID"]

	err := kh.kitchenService.MarkPickedUpByDriver(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Order picked up by driver",
		"orderID": orderID,
	})
}

func (kh *HTTPHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["orderID"]

	err := kh.kitchenService.CancelOrder(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Order cancelled",
		"orderID": orderID,
	})
}

func (kh *HTTPHandler) GetKitchenStats(w http.ResponseWriter, r *http.Request) {
	stats, err := kh.kitchenService.GetKitchenStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(stats)
}

func (kh *HTTPHandler) ProcessQueue(w http.ResponseWriter, r *http.Request) {
	err := kh.kitchenService.ProcessOrderQueue(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Queue processed successfully",
	})
}

func (kh *HTTPHandler) ServeDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"message": "Kitchen dashboard moved to frontend"}`))
}
