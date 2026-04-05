package handlers

import (
	"net/http"
	"strconv"

	"github.com/anilbbsr/vedatri.com/backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	DB *gorm.DB
}

// ── List Products (GET /api/products) ───────────────────────────────────────
// Query params: page, limit, category, search, featured, sort (price_asc|price_desc|newest)

func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 12
	}
	offset := (page - 1) * limit

	query := h.DB.Model(&models.Product{}).
		Preload("Category").
		Preload("Images").
		Where("published = true")

	if cat := c.Query("category"); cat != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.slug = ?", cat)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("products.name ILIKE ? OR products.description ILIKE ?",
			"%"+search+"%", "%"+search+"%")
	}
	if c.Query("featured") == "true" {
		query = query.Where("featured = true")
	}

	switch c.DefaultQuery("sort", "newest") {
	case "price_asc":
		query = query.Order("price ASC")
	case "price_desc":
		query = query.Order("price DESC")
	default:
		query = query.Order("products.created_at DESC")
	}

	var total int64
	query.Count(&total)

	var products []models.Product
	if err := query.Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       products,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (int(total) + limit - 1) / limit,
	})
}

// ── Single Product (GET /api/products/:slug) ─────────────────────────────────

func (h *ProductHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var product models.Product
	err := h.DB.
		Preload("Category").
		Preload("Images").
		Where("slug = ? AND published = true", slug).
		First(&product).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

// ── Create Product (POST /api/products) — admin only ────────────────────────

type CreateProductRequest struct {
	Name        string   `json:"name"        binding:"required"`
	Slug        string   `json:"slug"        binding:"required"`
	Description string   `json:"description"`
	Price       float64  `json:"price"       binding:"required,gt=0"`
	SalePrice   *float64 `json:"sale_price"`
	Stock       int      `json:"stock"`
	CategoryID  string   `json:"category_id" binding:"required"`
	MetaTitle   string   `json:"meta_title"`
	MetaDesc    string   `json:"meta_description"`
	Featured    bool     `json:"featured"`
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Price:       req.Price,
		SalePrice:   req.SalePrice,
		Stock:       req.Stock,
		MetaTitle:   req.MetaTitle,
		MetaDesc:    req.MetaDesc,
		Featured:    req.Featured,
		Published:   true,
	}

	if err := h.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create product"})
		return
	}
	c.JSON(http.StatusCreated, product)
}

// ── Update Product (PUT /api/products/:slug) — admin only ───────────────────

func (h *ProductHandler) Update(c *gin.Context) {
	slug := c.Param("slug")
	var product models.Product
	if err := h.DB.Where("slug = ?", slug).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.DB.Save(&product)
	c.JSON(http.StatusOK, product)
}

// ── Delete Product (DELETE /api/products/:slug) — admin only ────────────────

func (h *ProductHandler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	if err := h.DB.Where("slug = ?", slug).Delete(&models.Product{}).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
