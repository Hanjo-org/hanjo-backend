package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/supplier/infrastructure"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
)

type demoSupplierSeed struct {
	Name          string
	Code          string
	ContactPerson string
	Phone         string
	Email         string
	Address       string
}

var demoSuppliers = []demoSupplierSeed{
	{
		Name:          "PT Diamond Cold Storage",
		Code:          "SUPP-DIAMOND",
		ContactPerson: "Budi Hartono (Key Account Manager)",
		Phone:         "+6281234567890",
		Email:         "order@diamond.co.id",
		Address:       "Kawasan Industri MM2100 Blok C, Cikarang Barat",
	},
	{
		Name:          "Toffin Indonesia",
		Code:          "SUPP-TOFFIN",
		ContactPerson: "Jessica Tan (Sales Representative)",
		Phone:         "+6281987654321",
		Email:         "sales@toffin.id",
		Address:       "Jl. Pluit Permai Raya No. 12, Pluit, Jakarta Utara",
	},
	{
		Name:          "CV Kemasan Prima Makmur",
		Code:          "SUPP-KEMASAN",
		ContactPerson: "Agus Setiawan",
		Phone:         "+6281345678901",
		Email:         "kemasanprima@gmail.com",
		Address:       "Jl. Daan Mogot KM 14 No. 8, Kalideres, Jakarta Barat",
	},
}

func seedSuppliers(ctx context.Context, db *gorm.DB, orgID string, log *zap.Logger) error {
	fmt.Println("\n🚚 Seeding Demo Suppliers for Organization...")

	now := time.Now().UTC()

	for _, s := range demoSuppliers {
		var existing infrastructure.SupplierModel
		err := db.WithContext(ctx).
			Where("organization_id = ? AND code = ?", orgID, s.Code).
			First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			newSupplier := infrastructure.SupplierModel{
				ID:             uuid.NewString(),
				OrganizationID: orgID,
				Name:           s.Name,
				Code:           ptr.Ptr(s.Code),
				ContactPerson:  ptr.Ptr(s.ContactPerson),
				Phone:          ptr.Ptr(s.Phone),
				Email:          ptr.Ptr(s.Email),
				Address:        ptr.Ptr(s.Address),
				Status:         "ACTIVE",
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := db.WithContext(ctx).Create(&newSupplier).Error; err != nil {
				log.Error("failed to seed supplier", zap.String("name", s.Name), zap.Error(err))
				return err
			}
			fmt.Printf("✅ Supplier created: %s [%s]\n", s.Name, s.Code)
		} else if err != nil {
			log.Error("failed checking supplier existence", zap.String("name", s.Name), zap.Error(err))
			return err
		} else {
			fmt.Printf("ℹ️  Supplier already exists: %s [%s]\n", existing.Name, *existing.Code)
		}
	}

	return nil
}
