package handlers

import (
	"net/http"

	"github.com/anilbbsr/vedatri.com/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryHandler struct {
	DB *gorm.DB
}

// GET /api/categories
func (h *CategoryHandler) List(c *gin.Context) {
	var categories []models.Category
	if err := h.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// GET /api/categories/:slug
func (h *CategoryHandler) Get(c *gin.Context) {
	var cat models.Category
	if err := h.DB.Where("slug = ?", c.Param("slug")).First(&cat).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}
	c.JSON(http.StatusOK, cat)
}

// GET /api/categories/:slug/products
func (h *CategoryHandler) Products(c *gin.Context) {
	var cat models.Category
	if err := h.DB.Where("slug = ?", c.Param("slug")).First(&cat).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	var products []models.Product
	h.DB.
		Preload("Images").
		Where("category_id = ? AND published = true", cat.ID).
		Order("created_at DESC").
		Find(&products)

	c.JSON(http.StatusOK, gin.H{"category": cat, "data": products})
}

// POST /api/categories — admin only
func (h *CategoryHandler) Create(c *gin.Context) {
	var cat models.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Create(&cat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create category"})
		return
	}
	c.JSON(http.StatusCreated, cat)
}
