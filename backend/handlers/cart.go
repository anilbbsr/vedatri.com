package handlers

import (
	"net/http"

	"github.com/anilbbsr/vedatri.com/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CartHandler struct {
	DB *gorm.DB
}

// findOrCreateCart returns the cart for the authenticated user.
func (h *CartHandler) findOrCreateCart(userID string) (*models.Cart, error) {
	var cart models.Cart
	err := h.DB.
		Preload("Items.Product.Images").
		Where("user_id = ?", userID).
		First(&cart).Error

	if err == gorm.ErrRecordNotFound {
		uid, _ := uuid.Parse(userID)
		cart = models.Cart{UserID: &uid}
		if err := h.DB.Create(&cart).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &cart, nil
}

// GET /api/cart
func (h *CartHandler) Get(c *gin.Context) {
	userID, _ := c.Get("userID")
	cart, err := h.findOrCreateCart(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch cart"})
		return
	}
	c.JSON(http.StatusOK, cart)
}

// POST /api/cart/items
type AddItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity"   binding:"required,min=1"`
}

func (h *CartHandler) AddItem(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate product exists and has stock
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var product models.Product
	if err := h.DB.First(&product, "id = ? AND published = true", productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if product.Stock < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient stock"})
		return
	}

	cart, err := h.findOrCreateCart(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load cart"})
		return
	}

	// If item already in cart → increment
	var existing models.CartItem
	err = h.DB.Where("cart_id = ? AND product_id = ?", cart.ID, productID).First(&existing).Error
	if err == nil {
		existing.Quantity += req.Quantity
		h.DB.Save(&existing)
	} else {
		item := models.CartItem{
			CartID:    cart.ID,
			ProductID: productID,
			Quantity:  req.Quantity,
		}
		h.DB.Create(&item)
	}

	// Reload cart
	h.DB.Preload("Items.Product.Images").First(cart, cart.ID)
	c.JSON(http.StatusOK, cart)
}

// PUT /api/cart/items/:id
func (h *CartHandler) UpdateItem(c *gin.Context) {
	itemID := c.Param("id")
	var body struct {
		Quantity int `json:"quantity" binding:"required,min=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Quantity == 0 {
		h.DB.Delete(&models.CartItem{}, "id = ?", itemID)
		c.JSON(http.StatusOK, gin.H{"message": "item removed"})
		return
	}

	h.DB.Model(&models.CartItem{}).Where("id = ?", itemID).Update("quantity", body.Quantity)
	c.JSON(http.StatusOK, gin.H{"message": "cart updated"})
}

// DELETE /api/cart/items/:id
func (h *CartHandler) RemoveItem(c *gin.Context) {
	itemID := c.Param("id")
	h.DB.Delete(&models.CartItem{}, "id = ?", itemID)
	c.JSON(http.StatusOK, gin.H{"message": "item removed"})
}
