package database

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/nwasiq/fieldops/backend/internal/models"
	"github.com/nwasiq/fieldops/backend/internal/services"
)

const testKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

// rule: §1.2 — the seed is idempotent and leaves a soft-deleted technician with visits and clock events behind
func TestSeed_IdempotentAndShapedForToday(t *testing.T) {
	db, err := ConnectSQLite("seed_test")
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	cipher, err := services.NewCipher(testKey)
	if err != nil {
		t.Fatal(err)
	}
	auth := services.NewAuthService(db, "seed-test-secret-at-least-32-chars-long")
	auth.SetHashCost(bcrypt.MinCost)
	users := services.NewUserService(db, cipher, auth)
	sites := services.NewSiteService(db)

	now, _ := time.Parse(time.RFC3339, "2026-07-14T13:00:00Z") // 14:00 BST
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if err := Seed(ctx, db, users, sites, now); err != nil {
			t.Fatalf("seed run %d: %v", i+1, err)
		}
	}

	var userCount, siteCount, visitCount, eventCount int64
	db.Unscoped().Model(&models.User{}).Count(&userCount)
	db.Model(&models.Site{}).Count(&siteCount)
	db.Model(&models.Visit{}).Count(&visitCount)
	db.Model(&models.ClockEvent{}).Count(&eventCount)
	if userCount != 6 || siteCount != 6 {
		t.Errorf("users %d sites %d, want 6 and 6 after two runs", userCount, siteCount)
	}
	if want := int64(3*(len(daySlots)+2) + 4); visitCount != want {
		t.Errorf("visits = %d, want %d after two runs", visitCount, want)
	}
	if eventCount == 0 {
		t.Error("no clock events seeded")
	}

	for _, email := range []string{"admin@fieldops.local", "tech1@fieldops.local"} {
		if _, _, err := auth.Login(ctx, email, SeedPassword); err != nil {
			t.Errorf("login %s: %v", email, err)
		}
	}
	if _, _, err := auth.Login(ctx, "tech4@fieldops.local", SeedPassword); services.KindOf(err) != services.KindUnauthorized {
		t.Errorf("departed technician login err = %v, want unauthorized", err)
	}

	var dana models.User
	if err := db.Unscoped().Where("email = ?", "tech4@fieldops.local").First(&dana).Error; err != nil {
		t.Fatal(err)
	}
	if !dana.DeletedAt.Valid {
		t.Fatal("Dana is not soft-deleted")
	}
	var danaVisits int64
	db.Model(&models.Visit{}).Where("technician_id = ? AND status = ?", dana.ID, models.VisitStatusCompleted).Count(&danaVisits)
	if danaVisits != 4 {
		t.Errorf("Dana has %d completed visits, want 4", danaVisits)
	}
	var statuses []string
	db.Model(&models.Visit{}).Distinct("status").Pluck("status", &statuses)
	if len(statuses) != len(models.AllVisitStatuses) {
		t.Errorf("seeded statuses = %v, want all four", statuses)
	}
	var cancelled []models.Visit
	db.Where("status = ?", models.VisitStatusCancelled).Find(&cancelled)
	if len(cancelled) != 2 || cancelled[0].CancelledByID == nil || cancelled[0].CancelledAt == nil {
		t.Errorf("cancelled visits = %d with attribution %v", len(cancelled), cancelled[0].CancelledByID != nil)
	}

	// The 14th's 23:30Z visit is 00:30 BST on the 15th and must list there (§3.3).
	late, _ := time.Parse(time.RFC3339, "2026-07-14T23:30:00Z")
	var boundary models.Visit
	if err := db.Where("scheduled_start = ?", late).First(&boundary).Error; err != nil {
		t.Fatalf("23:30Z boundary visit missing: %v", err)
	}
	onDay := func(day string) bool {
		start, end, err := services.UKDayRange(day)
		if err != nil {
			t.Fatal(err)
		}
		var count int64
		db.Model(&models.Visit{}).Where("id = ? AND scheduled_start >= ? AND scheduled_start < ?", boundary.ID, start, end).Count(&count)
		return count == 1
	}
	if onDay("2026-07-14") || !onDay("2026-07-15") {
		t.Errorf("23:30Z on 14 July: on 14th=%v on 15th=%v, want only the 15th", onDay("2026-07-14"), onDay("2026-07-15"))
	}
}
