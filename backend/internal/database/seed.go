package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nwasiq/fieldops/backend/internal/models"
	"github.com/nwasiq/fieldops/backend/internal/services"
)

// SeedPassword is the password every seeded user logs in with.
const SeedPassword = "fieldops"

type seedUser struct {
	Email     string
	FirstName string
	LastName  string
	Role      string
	Licence   string
	Departed  bool
}

var seedUsers = []seedUser{
	{Email: "admin@fieldops.local", FirstName: "Ada", LastName: "Admin", Role: models.RoleAdmin},
	{Email: "dispatcher@fieldops.local", FirstName: "Dev", LastName: "Dispatcher", Role: models.RoleDispatcher},
	{Email: "tech1@fieldops.local", FirstName: "Tom", LastName: "Field", Role: models.RoleTechnician, Licence: "GS-284611"},
	{Email: "tech2@fieldops.local", FirstName: "Tara", LastName: "Wrench", Role: models.RoleTechnician, Licence: "GS-301877"},
	{Email: "tech3@fieldops.local", FirstName: "Tariq", LastName: "Bolt", Role: models.RoleTechnician, Licence: "GS-155042"},
	{Email: "tech4@fieldops.local", FirstName: "Dana", LastName: "Departed", Role: models.RoleTechnician, Licence: "GS-099120", Departed: true},
}

var seedSites = []services.SiteInput{
	{Name: "Riverside Business Park", Address: "Unit 4, Riverside Business Park, Kingston Road, Leatherhead KT22 7PL", ContactPhone: "01372 496120"},
	{Name: "Northgate Shopping Centre", Address: "Northgate Shopping Centre, Market Street, Chester CH1 2AB", ContactPhone: "01244 318550"},
	{Name: "St Mary's Primary School", Address: "St Mary's Primary School, Church Lane, Guildford GU1 3QS", ContactPhone: "01483 562204"},
	{Name: "Harbour View Apartments", Address: "Harbour View Apartments, Marine Parade, Brighton BN2 1TL", ContactPhone: "01273 604411"},
	{Name: "Oakwood Medical Centre", Address: "Oakwood Medical Centre, 12 Oak Road, Leeds LS6 4DP", ContactPhone: "0113 275 8890"},
	{Name: "Greenfield Distribution Depot", Address: "Greenfield Distribution Depot, Junction 9 Industrial Estate, Milton Keynes MK10 0BD", ContactPhone: "01908 447210"},
}

// slot is one visit in a day's book, in Europe/London wall-clock time.
// tech is an index into the technician emails below; -1 leaves it unassigned.
type slot struct {
	hour, minute int
	durationMin  int
	site         int
	tech         int
	cancelled    bool
}

var technicianEmails = []string{"tech1@fieldops.local", "tech2@fieldops.local", "tech3@fieldops.local"}

// daySlots is the book for each of yesterday, today and tomorrow.
var daySlots = []slot{
	{8, 0, 120, 0, 0, false},
	{8, 30, 90, 1, 1, false},
	{9, 0, 120, 2, 2, false},
	{10, 30, 90, 3, 0, false},
	{11, 0, 90, 4, 1, false},
	{11, 30, 120, 5, 2, false},
	{13, 0, 120, 0, 0, false},
	{13, 30, 90, 2, 1, false},
	{14, 0, 120, 4, 2, false},
	{15, 30, 90, 1, 0, false},
	{16, 0, 90, 3, -1, false},
	{16, 30, 90, 5, 2, false},
}

// Seed populates the demo data relative to today's Europe/London day. It is
// idempotent: users and sites are matched by email and name, visits by site
// and scheduled start, so a boot never duplicates a row it already wrote.
func Seed(ctx context.Context, db *gorm.DB, users *services.UserService, sites *services.SiteService, now time.Time) error {
	now = now.UTC()
	today := services.UKDay(now)

	userIDs, err := seedUserRows(ctx, db, users)
	if err != nil {
		return err
	}
	siteIDs, err := seedSiteRows(ctx, db, sites)
	if err != nil {
		return err
	}
	dispatcherID := userIDs["dispatcher@fieldops.local"]

	for offset := -1; offset <= 1; offset++ {
		day, err := services.UKDayAdd(today, offset)
		if err != nil {
			return err
		}
		for i, s := range daySlots {
			start, err := services.UKClock(day, s.hour, s.minute)
			if err != nil {
				return err
			}
			var tech *uint
			if s.tech >= 0 {
				id := userIDs[technicianEmails[s.tech]]
				tech = &id
			}
			// One cancellation yesterday and one today (§3.4).
			cancelled := (offset == -1 && i == 3) || (offset == 0 && i == 7)
			plan := visitPlan{
				siteID:    siteIDs[s.site],
				tech:      tech,
				start:     start,
				end:       start.Add(time.Duration(s.durationMin) * time.Minute),
				cancelled: cancelled,
			}
			if err := seedVisit(ctx, db, plan, now, dispatcherID); err != nil {
				return err
			}
		}
		// Instants either side of midnight UTC. In BST 23:30Z is already the
		// next Europe/London day, so these are the rows the UK-day helpers earn
		// their keep on (§3.3).
		if err := seedBoundaryVisits(ctx, db, day, siteIDs, userIDs, now, dispatcherID); err != nil {
			return err
		}
	}

	if err := seedDeparted(ctx, db, today, siteIDs, userIDs, now, dispatcherID); err != nil {
		return err
	}
	return nil
}

func seedUserRows(ctx context.Context, db *gorm.DB, users *services.UserService) (map[string]uint, error) {
	ids := map[string]uint{}
	for _, su := range seedUsers {
		var existing models.User
		err := db.WithContext(ctx).Unscoped().Where("email = ?", su.Email).First(&existing).Error
		switch {
		case err == nil:
			ids[su.Email] = existing.ID
			continue
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf("look up seed user %s: %w", su.Email, err)
		}
		password := SeedPassword
		licence := su.Licence
		created, err := users.Create(ctx, services.UserInput{
			Email:         su.Email,
			FirstName:     su.FirstName,
			LastName:      su.LastName,
			Role:          su.Role,
			LicenceNumber: &licence,
			Password:      &password,
		})
		if err != nil {
			return nil, fmt.Errorf("seed user %s: %w", su.Email, err)
		}
		ids[su.Email] = created.ID
	}
	return ids, nil
}

func seedSiteRows(ctx context.Context, db *gorm.DB, sites *services.SiteService) ([]uint, error) {
	ids := make([]uint, 0, len(seedSites))
	for _, in := range seedSites {
		var existing models.Site
		err := db.WithContext(ctx).Unscoped().Where("name = ?", in.Name).First(&existing).Error
		switch {
		case err == nil:
			ids = append(ids, existing.ID)
			continue
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf("look up seed site %s: %w", in.Name, err)
		}
		created, err := sites.Create(ctx, in)
		if err != nil {
			return nil, fmt.Errorf("seed site %s: %w", in.Name, err)
		}
		ids = append(ids, created.ID)
	}
	return ids, nil
}

type visitPlan struct {
	siteID    uint
	tech      *uint
	start     time.Time
	end       time.Time
	cancelled bool
}

// seedVisit writes one visit unless a visit at that site and start already
// exists. Status follows the clock: a visit whose window has passed is
// completed with clock events, one in flight is in progress, one still to
// come is scheduled. An unassigned visit has nobody to clock it and stays
// scheduled.
func seedVisit(ctx context.Context, db *gorm.DB, plan visitPlan, now time.Time, dispatcherID uint) error {
	var count int64
	err := db.WithContext(ctx).Model(&models.Visit{}).
		Where("site_id = ? AND scheduled_start = ?", plan.siteID, plan.start.UTC()).
		Count(&count).Error
	if err != nil {
		return fmt.Errorf("check seed visit: %w", err)
	}
	if count > 0 {
		return nil
	}

	visit := models.Visit{
		SiteID:         plan.siteID,
		TechnicianID:   plan.tech,
		ScheduledStart: plan.start.UTC(),
		ScheduledEnd:   plan.end.UTC(),
		Status:         models.VisitStatusScheduled,
	}
	var events []models.ClockEvent
	switch {
	case plan.cancelled:
		cancelledAt := plan.start.Add(-2 * time.Hour)
		visit.Status = models.VisitStatusCancelled
		visit.CancelledByID = &dispatcherID
		visit.CancelledAt = &cancelledAt
	case plan.tech == nil:
		// nobody to clock it
	case !plan.end.After(now):
		visit.Status = models.VisitStatusCompleted
		events = []models.ClockEvent{
			{Kind: models.ClockKindIn, OccurredAt: plan.start.Add(5 * time.Minute), RecordedByID: *plan.tech},
			{Kind: models.ClockKindOut, OccurredAt: plan.end.Add(-10 * time.Minute), RecordedByID: *plan.tech},
		}
	case !plan.start.After(now):
		visit.Status = models.VisitStatusInProgress
		events = []models.ClockEvent{
			{Kind: models.ClockKindIn, OccurredAt: plan.start.Add(5 * time.Minute), RecordedByID: *plan.tech},
		}
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&visit).Error; err != nil {
			return fmt.Errorf("seed visit: %w", err)
		}
		for i := range events {
			events[i].VisitID = visit.ID
			events[i].OccurredAt = events[i].OccurredAt.UTC()
			if err := tx.Create(&events[i]).Error; err != nil {
				return fmt.Errorf("seed clock event: %w", err)
			}
		}
		return nil
	})
}

func seedBoundaryVisits(ctx context.Context, db *gorm.DB, day string, siteIDs []uint, userIDs map[string]uint, now time.Time, dispatcherID uint) error {
	date, err := time.Parse(services.DayFormat, day)
	if err != nil {
		return fmt.Errorf("parse seed day: %w", err)
	}
	lateNight := time.Date(date.Year(), date.Month(), date.Day(), 23, 30, 0, 0, time.UTC)
	earlyMorning := time.Date(date.Year(), date.Month(), date.Day(), 0, 15, 0, 0, time.UTC)
	tech1 := userIDs["tech1@fieldops.local"]
	tech2 := userIDs["tech2@fieldops.local"]
	plans := []visitPlan{
		{siteID: siteIDs[0], tech: &tech1, start: lateNight, end: lateNight.Add(time.Hour)},
		{siteID: siteIDs[1], tech: &tech2, start: earlyMorning, end: earlyMorning.Add(time.Hour)},
	}
	for _, plan := range plans {
		if err := seedVisit(ctx, db, plan, now, dispatcherID); err != nil {
			return err
		}
	}
	return nil
}

// seedDeparted gives Dana visits and clock events on yesterday and the day
// before, then soft-deletes her, so every attribution path is exercised by a
// user the default scope can no longer see (§1.2, §4.3, §5.2).
func seedDeparted(ctx context.Context, db *gorm.DB, today string, siteIDs []uint, userIDs map[string]uint, now time.Time, dispatcherID uint) error {
	dana := userIDs["tech4@fieldops.local"]
	book := []struct {
		offset, hour, minute, duration, site int
	}{
		{-2, 9, 0, 120, 2},
		{-2, 13, 0, 120, 4},
		{-1, 9, 30, 90, 4},
		{-1, 14, 30, 90, 5},
	}
	for _, b := range book {
		day, err := services.UKDayAdd(today, b.offset)
		if err != nil {
			return err
		}
		start, err := services.UKClock(day, b.hour, b.minute)
		if err != nil {
			return err
		}
		id := dana
		plan := visitPlan{siteID: siteIDs[b.site], tech: &id, start: start, end: start.Add(time.Duration(b.duration) * time.Minute)}
		if err := seedVisit(ctx, db, plan, now, dispatcherID); err != nil {
			return err
		}
	}
	result := db.WithContext(ctx).Delete(&models.User{}, dana)
	if result.Error != nil {
		return fmt.Errorf("soft-delete departed technician: %w", result.Error)
	}
	return nil
}
