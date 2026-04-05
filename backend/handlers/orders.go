package handlers

import (
	"net/http"

	"github.com/anilbbsr/vedatri.com/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderHandler struct {
	DB *gorm.DB
}

// POST /api/orders
type PlaceOrderRequest struct {
	ShippingName    string `json:"shipping_name"    binding:"required"`
	ShippingAddress string `json:"shipping_address" binding:"required"`
	ShippingCity    string `json:"shipping_city"    binding:"required"`
	ShippingCountry string `json:"shipping_country" binding:"required"`
	ShippingZip     string `json:"shipping_zip"     binding:"required"`
	Notes           string `json:"notes"`
}

func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, _ := uuid.Parse(userIDStr.(string))

	var req PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Load cart
	var cart models.Cart
	err := h.DB.
		Preload("Items.Product").
		Where("user_id = ?", userID).
		First(&cart).Error
	if err != nil || len(cart.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
		return
	}

	// Build order inside a transaction
	var order models.Order
	txErr := h.DB.Transaction(func(tx *gorm.DB) error {
		var total float64
		var orderItems []models.OrderItem

		for _, item := range cart.Items {
			// Check stock
			if item.Product.Stock < item.Quantity {
				return gorm.ErrInvalidData
			}
			price := item.Product.Price
			if item.Product.SalePrice != nil {
				price = *item.Product.SalePrice
			}
			subtotal := price * float64(item.Quantity)
			total += subtotal

			orderItems = append(orderItems, models.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				UnitPrice: price,
				Subtotal:  subtotal,
			})

			// Decrement stock
			tx.Model(&item.Product).Update("stock", item.Product.Stock-item.Quantity)
		}

		order = models.Order{
			UserID:          userID,
			Status:          "pending",
			Total:           total,
			Items:           orderItems,
			ShippingName:    req.ShippingName,
			ShippingAddress: req.ShippingAddress,
			ShippingCity:    req.ShippingCity,
			ShippingCountry: req.ShippingCountry,
			ShippingZip:     req.ShippingZip,
			Notes:           req.Notes,
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// Clear cart
		tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{})
		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not place order"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GET /api/orders
func (h *OrderHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	var orders []models.Order
	h.DB.
		Preload("Items.Product").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders)
	c.JSON(http.StatusOK, gin.H{"data": orders})
}

// GET /api/orders/:id
func (h *OrderHandler) Get(c *gin.Context) {
	userID, _ := c.Get("userID")
	orderID := c.Param("id")

	var order models.Order
	err := h.DB.
		Preload("Items.Product.Images").
		Where("id = ? AND user_id = ?", orderID, userID).
		First(&order).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}
