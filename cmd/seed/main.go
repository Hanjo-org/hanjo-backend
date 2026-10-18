package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/hanjo-erp/hanjo/internal/config"
	orgdomain "github.com/hanjo-erp/hanjo/internal/modules/organization/domain"
	orginfra "github.com/hanjo-erp/hanjo/internal/modules/organization/infrastructure"
	userdomain "github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	userinfra "github.com/hanjo-erp/hanjo/internal/modules/user/infrastructure"
	applogger "github.com/hanjo-erp/hanjo/pkg/logger"
	"github.com/hanjo-erp/hanjo/pkg/password"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
)

func main() {
	cfg := config.Load()

	log := applogger.Init(applogger.Config{
		Env:       cfg.AppEnv,
		FilePath:  cfg.LogFilePath,
		MaxSizeMB: cfg.LogMaxSizeMB,
	})
	defer applogger.Sync()

	db := config.ConnectDatabase(cfg, log)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgRepo := orginfra.NewOrganizationRepository(db)
	brandRepo := orginfra.NewBrandRepository(db)
	outletRepo := orginfra.NewOutletRepository(db)
	userRepo := userinfra.NewUserRepository(db)

	fmt.Println("\n🚀 Seeding Hanjo ERP initial multi-tenant data...")

	// 1. Tier 1: Organization (Holding Management)
	orgSlug := "harmoni-rasa"
	org, err := orgRepo.GetBySlug(ctx, orgSlug)
	if err != nil {
		org = orgdomain.NewOrganization("PT Harmoni Rasa Nusantara", orgSlug)
		if err := orgRepo.Create(ctx, org); err != nil {
			log.Fatal("failed to seed organization", zap.Error(err))
		}
		fmt.Printf("✅ Organization created: %s (ID: %s)\n", org.Name, org.ID)
	} else {
		fmt.Printf("ℹ️  Organization already exists: %s\n", org.Name)
	}

	// 2. Tier 2: Brands (Coffee Shop Concepts)
	brandSenja, err := brandRepo.GetBySlug(ctx, org.ID, "senja-coffee")
	if err != nil {
		brandSenja = orgdomain.NewBrand(org.ID, "Senja Coffee & Roastery", "senja-coffee")
		brandSenja.Tagline = ptr.Ptr("Specialty Beans & Artisan Manual Brew")
		if err := brandRepo.Create(ctx, brandSenja); err != nil {
			log.Fatal("failed to seed brand senja", zap.Error(err))
		}
		fmt.Printf("✅ Brand created: %s\n", brandSenja.Name)
	} else {
		fmt.Printf("ℹ️  Brand already exists: %s\n", brandSenja.Name)
	}

	brandDaily, err := brandRepo.GetBySlug(ctx, org.ID, "daily-coffee")
	if err != nil {
		brandDaily = orgdomain.NewBrand(org.ID, "Daily Coffee", "daily-coffee")
		brandDaily.Tagline = ptr.Ptr("Good Coffee Everyday - Grab & Go")
		if err := brandRepo.Create(ctx, brandDaily); err != nil {
			log.Fatal("failed to seed brand daily", zap.Error(err))
		}
		fmt.Printf("✅ Brand created: %s\n", brandDaily.Name)
	} else {
		fmt.Printf("ℹ️  Brand already exists: %s\n", brandDaily.Name)
	}

	// 3. Tier 3: Outlets (Physical Stores)
	outletSenopati, err := outletRepo.GetByCode(ctx, org.ID, "SC-SNP")
	if err != nil {
		outletSenopati = orgdomain.NewOutlet(org.ID, brandSenja.ID, "Senja Coffee - Senopati", "SC-SNP", "Jl. Senopati No. 12, Kebayoran Baru, Jakarta Selatan")
		outletSenopati.Phone = ptr.Ptr("08123456701")
		if err := outletRepo.Create(ctx, outletSenopati); err != nil {
			log.Fatal("failed to seed outlet senopati", zap.Error(err))
		}
		fmt.Printf("✅ Outlet created: %s (%s)\n", outletSenopati.Name, outletSenopati.Code)
	} else {
		fmt.Printf("ℹ️  Outlet already exists: %s\n", outletSenopati.Name)
	}

	outletKemang, err := outletRepo.GetByCode(ctx, org.ID, "DC-KMG")
	if err != nil {
		outletKemang = orgdomain.NewOutlet(org.ID, brandDaily.ID, "Daily Coffee - Kemang", "DC-KMG", "Jl. Kemang Raya No. 5, Mampang, Jakarta Selatan")
		outletKemang.Phone = ptr.Ptr("08123456702")
		if err := outletRepo.Create(ctx, outletKemang); err != nil {
			log.Fatal("failed to seed outlet kemang", zap.Error(err))
		}
		fmt.Printf("✅ Outlet created: %s (%s)\n", outletKemang.Name, outletKemang.Code)
	} else {
		fmt.Printf("ℹ️  Outlet already exists: %s\n", outletKemang.Name)
	}

	outletSudirman, err := outletRepo.GetByCode(ctx, org.ID, "DC-SDR")
	if err != nil {
		outletSudirman = orgdomain.NewOutlet(org.ID, brandDaily.ID, "Daily Coffee - Sudirman", "DC-SDR", "Gedung FX Sudirman Lt. 1, Jakarta Pusat")
		outletSudirman.Phone = ptr.Ptr("08123456703")
		if err := outletRepo.Create(ctx, outletSudirman); err != nil {
			log.Fatal("failed to seed outlet sudirman", zap.Error(err))
		}
		fmt.Printf("✅ Outlet created: %s (%s)\n", outletSudirman.Name, outletSudirman.Code)
	}

	outletBSD, err := outletRepo.GetByCode(ctx, org.ID, "DC-BSD")
	if err != nil {
		outletBSD = orgdomain.NewOutlet(org.ID, brandDaily.ID, "Daily Coffee - BSD", "DC-BSD", "Ruko Foresta Business Loft 2 No. 10, BSD City, Tangerang")
		outletBSD.Phone = ptr.Ptr("08123456704")
		if err := outletRepo.Create(ctx, outletBSD); err != nil {
			log.Fatal("failed to seed outlet bsd", zap.Error(err))
		}
		fmt.Printf("✅ Outlet created: %s (%s)\n", outletBSD.Name, outletBSD.Code)
	}

	// 4. Users & Staff
	defaultPassword, err := password.Hash("Password123!")
	if err != nil {
		log.Fatal("failed to hash password", zap.Error(err))
	}
	defaultCashierPIN, err := password.Hash("123456")
	if err != nil {
		log.Fatal("failed to hash PIN", zap.Error(err))
	}
	defaultBaristaPIN, err := password.Hash("654321")
	if err != nil {
		log.Fatal("failed to hash PIN", zap.Error(err))
	}

	usersToSeed := []struct {
		username string
		email    string
		brandID  *string
		outletID *string
		pin      *string
		roleDesc string
	}{
		{
			username: "developer",
			email:    "dev@hanjo.id",
			brandID:  nil, // Global / Full Access
			outletID: nil,
			roleDesc: "Developer / Super Admin (Full Access)",
		},
		{
			username: "owner",
			email:    "owner@harmoni.id",
			brandID:  nil, // Global HQ scope
			outletID: nil,
			roleDesc: "Owner / Group Management",
		},
		{
			username: "manager.daily",
			email:    "manager.daily@harmoni.id",
			brandID:  &brandDaily.ID, // Scoped to Daily Coffee
			outletID: nil,
			roleDesc: "Brand Manager Daily Coffee",
		},
		{
			username: "manager.kemang",
			email:    "manager.kemang@harmoni.id",
			brandID:  &brandDaily.ID,
			outletID: &outletKemang.ID, // Scoped to Kemang outlet
			roleDesc: "Branch Manager Kemang",
		},
		{
			username: "kasir.kemang",
			email:    "kasir.kemang@harmoni.id",
			brandID:  &brandDaily.ID,
			outletID: &outletKemang.ID,
			pin:      &defaultCashierPIN, // PIN: 123456
			roleDesc: "Kasir POS Kemang",
		},
		{
			username: "barista.kemang",
			email:    "barista.kemang@harmoni.id",
			brandID:  &brandDaily.ID,
			outletID: &outletKemang.ID,
			pin:      &defaultBaristaPIN, // PIN: 654321
			roleDesc: "Barista / Kitchen Display Kemang",
		},
	}

	for _, u := range usersToSeed {
		user, err := userRepo.GetByEmail(ctx, u.email)
		if err != nil {
			newUser := userdomain.NewUser(org.ID, u.username, u.email, defaultPassword)
			newUser.BrandID = u.brandID
			newUser.OutletID = u.outletID
			newUser.Pin = u.pin

			if err := userRepo.Create(ctx, newUser); err != nil {
				log.Fatal("failed to seed user "+u.username, zap.Error(err))
			}
			fmt.Printf("✅ User created: %s (%s) - %s\n", u.username, u.email, u.roleDesc)
		} else {
			user.Password = defaultPassword
			user.Pin = u.pin
			if err := userRepo.Update(ctx, user); err != nil {
				log.Fatal("failed to update user "+u.username, zap.Error(err))
			}
			fmt.Printf("🔄 User updated with bcrypt: %s (%s)\n", user.Username, user.Email)
		}
	}

	fmt.Println("\n=========================================================================")
	fmt.Println("🎉 Seeding Completed Successfully!")
	fmt.Println("=========================================================================")
	fmt.Println("Demo Credentials:")
	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println("0. Developer / Super Admin (Full Access)")
	fmt.Println("   Email       : dev@hanjo.id")
	fmt.Println("   Password    : Password123!")
	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println("1. Owner (Dashboard Web)")
	fmt.Println("   Email       : owner@harmoni.id")
	fmt.Println("   Password    : Password123!")
	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println("2. Kasir Kemang (POS Tablet Login)")
	fmt.Printf("   Outlet ID   : %s\n", outletKemang.ID)
	fmt.Println("   Username    : kasir.kemang")
	fmt.Println("   PIN         : 123456")
	fmt.Println("   Email/Pass  : kasir.kemang@harmoni.id / Password123!")
	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println("3. Barista Kemang (Kitchen/Bar Display)")
	fmt.Printf("   Outlet ID   : %s\n", outletKemang.ID)
	fmt.Println("   Username    : barista.kemang")
	fmt.Println("   PIN         : 654321")
	fmt.Println("=========================================================================")
}
