package database

import (
	"log"

	"github.com/anilbbsr/vedatri.com/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Println("✅ Database connected")
	DB = db
	return db
}

// Migrate runs GORM AutoMigrate for all models.
func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.ProductImage{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
	)
	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	log.Println("✅ Database migrated")
}
