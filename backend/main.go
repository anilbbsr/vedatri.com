package main

import (
	"log"

	"github.com/anilbbsr/vedatri.com/config"
	"github.com/anilbbsr/vedatri.com/database"
	"github.com/anilbbsr/vedatri.com/handlers"
	"github.com/anilbbsr/vedatri.com/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// ── Config ──────────────────────────────────────────────────────────────
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	// ── Database ─────────────────────────────────────────────────────────────
	db := database.Connect(cfg.DSN())
	database.Migrate(db)

	// ── Router ───────────────────────────────────────────────────────────────
	r := gin.Default()
	r.Use(middleware.CORS())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ── Handlers ─────────────────────────────────────────────────────────────
	authH := &handlers.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	productH := &handlers.ProductHandler{DB: db}
	categoryH := &handlers.CategoryHandler{DB: db}
	cartH := &handlers.CartHandler{DB: db}
	orderH := &handlers.OrderHandler{DB: db}

	api := r.Group("/api")

	// ── Auth routes (public) ─────────────────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.GET("/me", middleware.AuthRequired(cfg.JWTSecret), authH.Me)
	}

	// ── Product routes ───────────────────────────────────────────────────────
	products := api.Group("/products")
	{
		products.GET("", productH.List)
		products.GET("/:slug", productH.GetBySlug)

		// Admin only
		adminProducts := products.Group("")
		adminProducts.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.AdminOnly())
		{
			adminProducts.POST("", productH.Create)
			adminProducts.PUT("/:slug", productH.Update)
			adminProducts.DELETE("/:slug", productH.Delete)
		}
	}

	// ── Category routes ──────────────────────────────────────────────────────
	categories := api.Group("/categories")
	{
		categories.GET("", categoryH.List)
		categories.GET("/:slug", categoryH.Get)
		categories.GET("/:slug/products", categoryH.Products)

		// Admin only
		adminCats := categories.Group("")
		adminCats.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.AdminOnly())
		{
			adminCats.POST("", categoryH.Create)
		}
	}

	// ── Cart routes (authenticated) ──────────────────────────────────────────
	cart := api.Group("/cart")
	cart.Use(middleware.AuthRequired(cfg.JWTSecret))
	{
		cart.GET("", cartH.Get)
		cart.POST("/items", cartH.AddItem)
		cart.PUT("/items/:id", cartH.UpdateItem)
		cart.DELETE("/items/:id", cartH.RemoveItem)
	}

	// ── Order routes (authenticated) ─────────────────────────────────────────
	orders := api.Group("/orders")
	orders.Use(middleware.AuthRequired(cfg.JWTSecret))
	{
		orders.POST("", orderH.PlaceOrder)
		orders.GET("", orderH.List)
		orders.GET("/:id", orderH.Get)
	}

	// ── Start ─────────────────────────────────────────────────────────────────
	addr := ":" + cfg.Port
	log.Printf("🚀 API running at http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
