package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/uom/infrastructure"
)

type systemUoMSeed struct {
	Name          string
	Symbol        string
	Category      string
	DecimalPlaces int
}

var systemUoMs = []systemUoMSeed{
	{Name: "Milliliter", Symbol: "ml", Category: "VOLUME", DecimalPlaces: 2},
	{Name: "Liter", Symbol: "l", Category: "VOLUME", DecimalPlaces: 2},
	{Name: "Gram", Symbol: "g", Category: "WEIGHT", DecimalPlaces: 2},
	{Name: "Kilogram", Symbol: "kg", Category: "WEIGHT", DecimalPlaces: 3},
	{Name: "Piece", Symbol: "pcs", Category: "UNIT", DecimalPlaces: 0},
	{Name: "Pack", Symbol: "pack", Category: "UNIT", DecimalPlaces: 0},
	{Name: "Bottle", Symbol: "btl", Category: "UNIT", DecimalPlaces: 0},
	{Name: "Can", Symbol: "can", Category: "UNIT", DecimalPlaces: 0},
	{Name: "Carton / Box", Symbol: "ctn", Category: "UNIT", DecimalPlaces: 0},
}

func seedUoM(ctx context.Context, db *gorm.DB, log *zap.Logger) error {
	fmt.Println("\n⚖️  Seeding System Default UoMs & Metric Conversions...")

	now := time.Now().UTC()

	// 1. Seed System Units with Real UUID v4
	uomMap := make(map[string]string, len(systemUoMs))

	for _, u := range systemUoMs {
		var existing infrastructure.UoMModel
		err := db.WithContext(ctx).
			Where("symbol = ? AND organization_id IS NULL", u.Symbol).
			First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			newID := uuid.NewString() // Generate real RFC-4122 v4 UUID
			newUoM := infrastructure.UoMModel{
				ID:             newID,
				OrganizationID: nil,
				Name:           u.Name,
				Symbol:         u.Symbol,
				Category:       u.Category,
				DecimalPlaces:  u.DecimalPlaces,
				IsSystem:       true,
				Status:         "ACTIVE",
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := db.WithContext(ctx).Create(&newUoM).Error; err != nil {
				return fmt.Errorf("failed to insert system uom %s: %w", u.Symbol, err)
			}
			uomMap[u.Symbol] = newID
			fmt.Printf("   ✅ System UoM created: %s (%s) [%s] (ID: %s)\n", u.Name, u.Symbol, u.Category, newID)
		} else if err != nil {
			return fmt.Errorf("failed to query uom %s: %w", u.Symbol, err)
		} else {
			uomMap[u.Symbol] = existing.ID
			fmt.Printf("   ℹ️  System UoM exists: %s (%s) (ID: %s)\n", u.Name, u.Symbol, existing.ID)
		}
	}

	// 2. Seed Standard Global Conversions (kg -> g, l -> ml)
	systemConversions := []struct {
		FromSymbol string
		ToSymbol   string
		Factor     float64
	}{
		{FromSymbol: "kg", ToSymbol: "g", Factor: 1000.0},
		{FromSymbol: "l", ToSymbol: "ml", Factor: 1000.0},
	}

	for _, c := range systemConversions {
		fromID, okFrom := uomMap[c.FromSymbol]
		toID, okTo := uomMap[c.ToSymbol]
		if !okFrom || !okTo {
			return fmt.Errorf("missing UoM IDs for standard conversion: %s -> %s", c.FromSymbol, c.ToSymbol)
		}

		var existingConv infrastructure.UoMConversionModel
		err := db.WithContext(ctx).
			Where("from_uom_id = ? AND to_uom_id = ?", fromID, toID).
			First(&existingConv).Error

		if err == gorm.ErrRecordNotFound {
			newConvID := uuid.NewString() // Generate real RFC-4122 v4 UUID
			newConv := infrastructure.UoMConversionModel{
				ID:             newConvID,
				OrganizationID: nil,
				FromUoMID:      fromID,
				ToUoMID:        toID,
				Factor:         c.Factor,
				IsSystem:       true,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := db.WithContext(ctx).Create(&newConv).Error; err != nil {
				return fmt.Errorf("failed to insert standard conversion %s -> %s: %w", c.FromSymbol, c.ToSymbol, err)
			}
			fmt.Printf("   ✅ Standard Conversion created: %s -> %s (factor: %.0f) [ID: %s]\n", c.FromSymbol, c.ToSymbol, c.Factor, newConvID)
		} else if err != nil {
			return fmt.Errorf("failed to query conversion: %w", err)
		} else {
			fmt.Printf("   ℹ️  Standard Conversion exists: %s -> %s\n", c.FromSymbol, c.ToSymbol)
		}
	}

	return nil
}
