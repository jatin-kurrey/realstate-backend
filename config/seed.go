package config

import (
	"log"
	"realstate-backend/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func SeedData() {
	// 1. Seed Users (Admin, Owners, Seekers)
	var userCount int64
	DB.Model(&models.User{}).Count(&userCount)
	if userCount < 6 {
		log.Println("Seeding core users...")
		userPass, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

		users := []models.User{
			{Name: "Alok Verma", Email: "alok@example.com", Password: string(userPass), Role: "owner", Phone: "+91 94252 12345", PublicPreference: "Full"},
			{Name: "Priya Singh", Email: "priya@example.com", Password: string(userPass), Role: "owner", Phone: "+91 94252 67890", PublicPreference: "Anonymized"},
			{Name: "Rahul Sharma", Email: "rahul@example.com", Password: string(userPass), Role: "seeker", Phone: "+91 91790 11223", PublicPreference: "Full"},
			{Name: "Amit Gupta", Email: "amit@example.com", Password: string(userPass), Role: "seeker", Phone: "+91 91790 44556", PublicPreference: "Anonymized"},
			{Name: "Siddharth Jain", Email: "sid@example.com", Password: string(userPass), Role: "owner", Phone: "+91 98271 77889", PublicPreference: "Full"},
		}

		for _, u := range users {
			var existing models.User
			if err := DB.Where("email = ?", u.Email).First(&existing).Error; err != nil {
				DB.Create(&u)
			}
		}
	}

	// Always ensure Admin exists and has correct credentials
	adminPass, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminUser := models.User{Name: "Master Admin", Email: "admin@rjg.com", Password: string(adminPass), Role: "admin"}

	var existingAdmin models.User
	if err := DB.Where("email = ?", adminUser.Email).First(&existingAdmin).Error; err != nil {
		log.Println("Creating Admin User...")
		DB.Create(&adminUser)
	} else {
		// FORCE UPDATE: Verify this user is strictly an ADMIN with known password
		// This fixes cases where user might have accidentally changed role in UI
		existingAdmin.Password = adminUser.Password
		existingAdmin.Role = "admin"
		DB.Save(&existingAdmin)
	}

	// Get some user IDs for seeding
	var owner1, owner2, owner3, seeker1 models.User
	DB.Where("email = ?", "alok@example.com").First(&owner1)
	DB.Where("email = ?", "priya@example.com").First(&owner2)
	DB.Where("email = ?", "sid@example.com").First(&owner3)
	DB.Where("email = ?", "rahul@example.com").First(&seeker1)

	// 2. Seed Properties (Diverse set)
	var propertyCount int64
	DB.Model(&models.Property{}).Count(&propertyCount)
	if propertyCount < 8 {
		log.Println("Seeding premium properties...")
		properties := []models.Property{
			{
				Title:  "Luxury 3BHK Villa with Garden",
				Status: "Sale", Type: "Residential", Area: 2500, Dimensions: "50x50 ft",
				Description: "Beautiful luxury villa in a gated community with modular kitchen and private garden.",
				Price:       7500000, Location: "Kacheri Chowk, Rajnandgaon",
				ImageUrl:   "https://images.unsplash.com/photo-1613490493576-7fde63acd811?auto=format&fit=crop&q=80&w=800",
				IsVerified: true, IsFeatured: true, IsActive: true, OwnerID: owner1.ID,
			},
			{
				Title:  "Prime Commercial Showroom",
				Status: "Rent", Type: "Commercial", Area: 1200, Dimensions: "40x30 ft",
				Description: "High-visibility showroom space on the ground floor. Ideal for retail brands.",
				Price:       45000, Location: "G.E. Road, Rajnandgaon",
				ImageUrl:   "https://images.unsplash.com/photo-1497366216548-37526070297c?auto=format&fit=crop&q=80&w=800",
				IsVerified: true, IsFeatured: true, IsActive: true, OwnerID: owner2.ID,
			},
			{
				Title:  "Residential Plot near New Bus Stand",
				Status: "Sale", Type: "Land", Area: 1500, Dimensions: "30x50 ft",
				Description: "Well-leveled residential plot in a fast-developing colony. East facing.",
				Price:       2800000, Location: "Lakholi, Rajnandgaon",
				ImageUrl:   "https://images.unsplash.com/photo-1500382017468-9049fed747ef?auto=format&fit=crop&q=80&w=800",
				IsVerified: false, IsFeatured: false, IsActive: true, OwnerID: owner3.ID,
			},
			{
				Title:  "Modern 2BHK Apartment",
				Status: "Rent", Type: "Residential", Area: 1100, Dimensions: "40x27 ft",
				Description: "Well-ventilated flat with lift, power backup, and dedicated parking.",
				Price:       12000, Location: "Basantpur, Rajnandgaon",
				ImageUrl:   "https://images.unsplash.com/photo-1522708323590-d24dbb6b0267?auto=format&fit=crop&q=80&w=800",
				IsVerified: true, IsFeatured: false, IsActive: true, OwnerID: owner1.ID,
			},
			{
				Title:  "Industrial Shed / Warehouse",
				Status: "Sale", Type: "Commercial", Area: 5000, Dimensions: "100x50 ft",
				Description: "Large industrial space with high ceiling and heavy power load capacity.",
				Price:       6500000, Location: "Tedezara Industrial Area",
				ImageUrl:   "https://images.unsplash.com/photo-1586528116311-ad861962bf3d?auto=format&fit=crop&q=80&w=800",
				IsVerified: true, IsActive: true, OwnerID: owner2.ID,
			},
			{
				Title:  "Commercial Plot on GE Road",
				Status: "Sale", Type: "Land", Area: 5000, Dimensions: "50x100 ft",
				Description: "High value commercial land on main highway. Suitable for hotel or hospital.",
				Price:       15000000, Location: "G.E. Road, Rajnandgaon",
				ImageUrl:   "https://images.unsplash.com/photo-1542253816-3e0e85295c5c?auto=format&fit=crop&q=80&w=800",
				IsVerified: false, IsFeatured: false, IsActive: true, OwnerID: owner3.ID,
			},
			{
				Title:  "2BHK Flat in Posu Colony",
				Status: "Rent", Type: "Residential", Area: 950, Dimensions: "30x32 ft",
				Description: "Affordable flat for small family. Near school and hospital.",
				Price:       8000, Location: "Posu Colony, Rajnandgaon",
				ImageUrl:   "https://images.unsplash.com/photo-1560448204-e02f11c3d0e2?auto=format&fit=crop&q=80&w=800",
				IsVerified: false, IsFeatured: false, IsActive: true, OwnerID: owner2.ID,
			},
		}
		for _, p := range properties {
			DB.Create(&p)
		}
	}

	// 3. Seed Requirements (Marketplace demand)
	var reqCount int64
	DB.Model(&models.Requirement{}).Count(&reqCount)
	if reqCount < 8 { // Increased count to ensure seed runs
		log.Println("Seeding community requirements...")
		requirements := []models.Requirement{
			{
				Purpose: "Buy", Type: "Residential", MinArea: 1200, MaxArea: 1800,
				MinBudget: 3500000, MaxBudget: 5000000, Location: "Basantpur / Kaurinbhata",
				Description:   "Looking for a ready-to-move 3BHK house with proper documentation.",
				ContactMethod: "In-app", UserID: seeker1.ID, IsVerified: true, IsActive: true,
			},
			{
				Purpose: "Rent", Type: "Commercial", MinArea: 300, MaxArea: 700,
				MinBudget: 10000, MaxBudget: 25000, Location: "Main Market / Ganj Line",
				Description:   "Searching for a small shop space for pharmacy business. Ground floor preferred.",
				ContactMethod: "Phone", UserID: seeker1.ID, IsVerified: true, IsActive: true,
			},
			{
				Purpose: "Buy", Type: "Land", MinArea: 2000, MaxArea: 4000,
				MinBudget: 1500000, MaxBudget: 3000000, Location: "Stadium Road",
				Description:   "Interested in residential plots in established colonies. Direct owners only.",
				ContactMethod: "Email", UserID: seeker1.ID, IsVerified: false, IsActive: true, // Pending
			},
			{
				Purpose: "Rent", Type: "Residential", MinArea: 800, MaxArea: 1200,
				MinBudget: 6000, MaxBudget: 10000, Location: "Lakholi",
				Description:   "Need a 2BHK flat for employee accommodation. Immediate joining possible.",
				ContactMethod: "WhatsApp", UserID: owner1.ID, IsVerified: true, IsActive: true,
			},
			{
				Purpose: "Buy", Type: "Commercial", MinArea: 1000, MaxArea: 2000,
				MinBudget: 5000000, MaxBudget: 8000000, Location: "Near Railway Station",
				Description:   "Looking for office space near railway station. Parking is a must.",
				ContactMethod: "Phone", UserID: owner3.ID, IsVerified: false, IsActive: true, // Pending
			},
		}
		for _, r := range requirements {
			DB.Create(&r)
		}
	}

	// 4. Seed Payments (Revenue visualization)
	var payCount int64
	DB.Model(&models.Payment{}).Count(&payCount)
	if payCount == 0 {
		log.Println("Seeding sample listing payments...")
		payments := []models.Payment{
			{UserID: owner1.ID, Amount: 100, Status: "Success", CreatedAt: time.Now().AddDate(0, 0, -2)},
			{UserID: owner2.ID, Amount: 100, Status: "Success", CreatedAt: time.Now().AddDate(0, 0, -5)},
			{UserID: owner3.ID, Amount: 100, Status: "Pending", CreatedAt: time.Now()},
			{UserID: seeker1.ID, Amount: 500, Status: "Success", CreatedAt: time.Now().AddDate(0, -1, 0)}, // Premium Seeker plan
		}
		for _, p := range payments {
			DB.Create(&p)
		}
	}

	// 5. Seed Site Config & Page Content
	var configCount int64
	DB.Model(&models.SiteConfig{}).Count(&configCount)
	if configCount == 0 {
		log.Println("Seeding CMS configurations...")
		configs := []models.SiteConfig{
			{Key: "site_name", Value: "RJG Property Connect", Group: "general", Type: "text"},
			{Key: "support_email", Value: "hq@rjgproperty.com", Group: "general", Type: "email"},
			{Key: "phone_number", Value: "+91 98765 43210", Group: "general", Type: "text"},
			{Key: "office_address", Value: "Main Road, Rajnandgaon, Chhattisgarh 491441", Group: "general", Type: "textarea"},
			{Key: "facebook_url", Value: "https://facebook.com", Group: "social", Type: "text"},
			{Key: "instagram_url", Value: "https://instagram.com", Group: "social", Type: "text"},
			{Key: "hero_title", Value: "Property Requirements", Group: "home", Type: "text"},
			{Key: "hero_subtitle", Value: "Browse what buyers and tenants are looking for in Rajnandgaon, or post your own requirement to connect with property owners.", Group: "home", Type: "textarea"},
			{Key: "about_text", Value: "The premier real estate bridge for Rajnandgaon and beyond. Verified community property intelligence.", Group: "about", Type: "textarea"},
		}
		for _, c := range configs {
			DB.Create(&c)
		}
	}

	// FORCE UPDATE: Ensure all unverified properties/requirements are Active (Direct Listing Fix)
	// This ensures existing items show up even if they were created before the default changed.
	DB.Model(&models.Property{}).Where("is_verified = ?", false).Update("is_active", true)
	DB.Model(&models.Requirement{}).Where("is_verified = ?", false).Update("is_active", true)

	log.Println("Data seeding successfully updated.")
}
