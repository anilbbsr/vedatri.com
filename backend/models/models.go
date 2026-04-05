package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ────────────────────────────────────────────────────────────────────────────
// Base
// ────────────────────────────────────────────────────────────────────────────

type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"                json:"id"`
	CreatedAt time.Time      `                                           json:"created_at"`
	UpdatedAt time.Time      `                                           json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                               json:"-"`
}

func (b *Base) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// ────────────────────────────────────────────────────────────────────────────
// User
// ────────────────────────────────────────────────────────────────────────────

type User struct {
	Base
	Name         string  `gorm:"not null"              json:"name"`
	Email        string  `gorm:"uniqueIndex;not null"  json:"email"`
	PasswordHash string  `gorm:"not null"              json:"-"`
	Role         string  `gorm:"default:'customer'"    json:"role"` // customer | admin
	Orders       []Order `                             json:"orders,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Category
// ────────────────────────────────────────────────────────────────────────────

type Category struct {
	Base
	Name        string    `gorm:"not null"             json:"name"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	Description string    `                            json:"description"`
	ImageURL    string    `                            json:"image_url"`
	Products    []Product `                            json:"products,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Product
// ────────────────────────────────────────────────────────────────────────────

type Product struct {
	Base
	Name        string         `gorm:"not null"             json:"name"`
	Slug        string         `gorm:"uniqueIndex;not null" json:"slug"`
	Description string         `gorm:"type:text"            json:"description"`
	Price       float64        `gorm:"not null"             json:"price"`
	SalePrice   *float64       `                            json:"sale_price,omitempty"`
	Stock       int            `gorm:"default:0"            json:"stock"`
	CategoryID  uuid.UUID      `                            json:"category_id"`
	Category    Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Images      []ProductImage `                            json:"images,omitempty"`
	MetaTitle   string         `                            json:"meta_title"`
	MetaDesc    string         `                            json:"meta_description"`
	Featured    bool           `gorm:"default:false"        json:"featured"`
	Published   bool           `gorm:"default:true"         json:"published"`
}

type ProductImage struct {
	Base
	ProductID uuid.UUID `json:"product_id"`
	URL       string    `gorm:"not null" json:"url"`
	AltText   string    `json:"alt_text"`
	IsPrimary bool      `gorm:"default:false" json:"is_primary"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
}

// ────────────────────────────────────────────────────────────────────────────
// Cart
// ────────────────────────────────────────────────────────────────────────────

type Cart struct {
	Base
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	SessionID string     `gorm:"index" json:"session_id"`
	Items     []CartItem `json:"items,omitempty"`
}

type CartItem struct {
	Base
	CartID    uuid.UUID `json:"cart_id"`
	ProductID uuid.UUID `json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int       `gorm:"default:1" json:"quantity"`
}

// ────────────────────────────────────────────────────────────────────────────
// Order
// ────────────────────────────────────────────────────────────────────────────

// Status values: pending | processing | shipped | delivered | cancelled
type Order struct {
	Base
	UserID          uuid.UUID   `                             json:"user_id"`
	User            User        `gorm:"foreignKey:UserID"     json:"user,omitempty"`
	Status          string      `gorm:"default:'pending'"     json:"status"`
	Total           float64     `                             json:"total"`
	Items           []OrderItem `                             json:"items,omitempty"`
	ShippingName    string      `                             json:"shipping_name"`
	ShippingAddress string      `                             json:"shipping_address"`
	ShippingCity    string      `                             json:"shipping_city"`
	ShippingCountry string      `                             json:"shipping_country"`
	ShippingZip     string      `                             json:"shipping_zip"`
	Notes           string      `                             json:"notes"`
}

type OrderItem struct {
	Base
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	Subtotal  float64   `json:"subtotal"`
}
