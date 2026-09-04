package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"shopping-service/internal/application/ports"
	"shopping-service/internal/domain/models"
)

// ProductHandler verarbeitet HTTP-Anfragen fuer Produkte.
type ProductHandler struct {
	productService ports.ProductService
	eventPublisher ports.EventPublisher
}

// NewProductHandler erstellt einen neuen ProductHandler.
func NewProductHandler(productService ports.ProductService, eventPublisher ports.EventPublisher) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		eventPublisher: eventPublisher,
	}
}

// CreateProduct verarbeitet POST /products Anfragen.
// Erfordert Admin-Rolle.
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product domain.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungueltige Anfrage"})
		return
	}

	// Benutzer-ID aus JWT holen
	if userID, exists := c.Get(ContextUserIDKey); exists {
		product.UserID = userID.(string)
	}

	if err := h.productService.CreateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Speichern"})
		return
	}

	// Event an Kafka senden
	if err := h.eventPublisher.PublishMessage(context.Background(), product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kafka Sendefehler"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// ListProducts verarbeitet GET /products Anfragen.
// öffentlicher Endpunkt.
func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.productService.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Laden der Produkte"})
		return
	}
	c.JSON(http.StatusOK, products)
}

// CartHandler verarbeitet HTTP-Anfragen fuer den Warenkorb.
type CartHandler struct {
	cartService    ports.CartService
	productService ports.ProductService
	eventPublisher ports.EventPublisher
}

// NewCartHandler erstellt einen neuen CartHandler.
func NewCartHandler(cartService ports.CartService, productService ports.ProductService, eventPublisher ports.EventPublisher) *CartHandler {
	return &CartHandler{
		cartService:    cartService,
		productService: productService,
		eventPublisher: eventPublisher,
	}
}

// AddToCartRequest definiert die Anfrage zum Hinzufuegen eines Artikels.
type AddToCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Qty       int    `json:"qty" binding:"required"`
}

// AddToCart verarbeitet POST /cart Anfragen.
func (h *CartHandler) AddToCart(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kein Benutzer"})
		return
	}

	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.cartService.AddToCart(userID, req.ProductID, req.Qty); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hinzufuegen fehlgeschlagen"})
		return
	}

	// Event an Kafka senden
	event := map[string]interface{}{
		"event_type": "item_added_to_cart",
		"user_id":    userID,
		"product_id": req.ProductID,
		"quantity":   req.Qty,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	_ = h.eventPublisher.PublishMessage(context.Background(), event)

	c.JSON(http.StatusOK, gin.H{"status": "hinzugefuegt", "product_id": req.ProductID, "qty": req.Qty})
}

// GetCart verarbeitet GET /cart Anfragen.
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, _ := GetUserID(c)

	cart, err := h.cartService.GetCart(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Laden"})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// UpdateCartItem verarbeitet PUT /cart Anfragen.
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kein Benutzer"})
		return
	}

	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Qty <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Menge muss groesser als 0 sein"})
		return
	}

	if err := h.cartService.UpdateCartItem(userID, req.ProductID, req.Qty); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aktualisierung fehlgeschlagen"})
		return
	}

	// Event senden
	event := map[string]interface{}{
		"event_type":   "cart_item_updated",
		"user_id":      userID,
		"product_id":   req.ProductID,
		"new_quantity": req.Qty,
		"timestamp":    time.Now().Format(time.RFC3339),
	}
	_ = h.eventPublisher.PublishMessage(context.Background(), event)

	c.JSON(http.StatusOK, gin.H{"message": "Warenkorb aktualisiert"})
}

// RemoveFromCart verarbeitet DELETE /cart/:product_id Anfragen.
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kein Benutzer"})
		return
	}

	productID := c.Param("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id erforderlich"})
		return
	}

	if err := h.cartService.RemoveFromCart(userID, productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Entfernen fehlgeschlagen"})
		return
	}

	// Event senden
	event := map[string]interface{}{
		"event_type": "item_removed_from_cart",
		"user_id":    userID,
		"product_id": productID,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	_ = h.eventPublisher.PublishMessage(context.Background(), event)

	c.JSON(http.StatusOK, gin.H{"message": "Artikel entfernt"})
}

// CheckoutRequest definiert die Anfrage fuer den Checkout.
type CheckoutRequest struct {
	OrderType    string                 `json:"order_type"`
	DeliveryInfo map[string]interface{} `json:"delivery_info,omitempty"`
}

// Checkout verarbeitet POST /checkout Anfragen.
func (h *CartHandler) Checkout(c *gin.Context) {
	userID, _ := GetUserID(c)

	// Optionale Checkout-Daten aus Request Body lesen
	var checkoutReq CheckoutRequest
	_ = c.ShouldBindJSON(&checkoutReq)

	// Default order_type ist "delivery"
	orderType := checkoutReq.OrderType
	if orderType == "" {
		orderType = "delivery"
	}

	cart, err := h.cartService.GetCart(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Laden des Warenkorbs"})
		return
	}

	if len(cart.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Warenkorb ist leer"})
		return
	}

	// Gesamtsumme berechnen
	var totalAmount float64
	orderItems := []map[string]interface{}{}
	productIDs := []string{}

	for _, item := range cart.Items {
		productIDs = append(productIDs, item.ProductID)

		product, err := h.productService.GetProductByID(item.ProductID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Produkt %s nicht gefunden", item.ProductID),
			})
			return
		}

		itemPrice := product.Price * float64(item.Qty)
		totalAmount += itemPrice

		orderItems = append(orderItems, map[string]interface{}{
			"product_id":   item.ProductID,
			"product_name": product.Name,
			"quantity":     item.Qty,
			"unit_price":   product.Price,
			"total_price":  itemPrice,
		})
	}

	// Bestellung erstellen
	order := map[string]interface{}{
		"event_type":   "order_created",
		"order_id":     fmt.Sprintf("ord-%d", time.Now().UnixNano()),
		"user_id":      userID,
		"items":        orderItems,
		"product_ids":  productIDs,
		"total_amount": totalAmount,
		"currency":     "EUR",
		"status":       "pending",
		"timestamp":    time.Now().Format(time.RFC3339),
		"order_type":   orderType,
	}

	// DeliveryInfo hinzufuegen falls vorhanden
	if checkoutReq.DeliveryInfo != nil {
		order["delivery_info"] = checkoutReq.DeliveryInfo
	}

	if err := h.eventPublisher.PublishMessage(context.Background(), order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bestellung konnte nicht erstellt werden"})
		return
	}

	// Warenkorb leeren
	_ = h.cartService.ClearCart(userID)

	c.JSON(http.StatusOK, gin.H{
		"status":       "order_created",
		"order_id":     order["order_id"],
		"total_amount": totalAmount,
		"items_count":  len(cart.Items),
	})
}
