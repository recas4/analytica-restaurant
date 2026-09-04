package rest

import "github.com/gin-gonic/gin"

// SetupRoutes konfiguriert alle HTTP-Routen fuer den Shopping-Service.
func SetupRoutes(r *gin.Engine, productHandler *ProductHandler, cartHandler *CartHandler) {
	// CORS Middleware
	r.Use(corsMiddleware())

	// CORS Preflight Handler
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(204)
	})

	// Oeffentliche Routen
	r.GET("/products", productHandler.ListProducts)
	r.GET("/health", healthCheck)

	// Admin-Routen (nur fuer Admins)
	adminGroup := r.Group("/")
	adminGroup.Use(JWTMiddleware())
	adminGroup.Use(RequireRole("admin"))
	{
		adminGroup.POST("/products", productHandler.CreateProduct)
	}

	// Benutzer-Routen (authentifiziert)
	userGroup := r.Group("/")
	userGroup.Use(JWTMiddleware())
	{
		userGroup.POST("/cart", cartHandler.AddToCart)
		userGroup.GET("/cart", cartHandler.GetCart)
		userGroup.PUT("/cart", cartHandler.UpdateCartItem)
		userGroup.DELETE("/cart/:product_id", cartHandler.RemoveFromCart)
		userGroup.POST("/checkout", cartHandler.Checkout)
	}
}

// corsMiddleware fuegt CORS-Header hinzu.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// healthCheck liefert den Gesundheitsstatus des Services.
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "healthy", "service": "shopping-service"})
}
