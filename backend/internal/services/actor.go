package services

import "github.com/nwasiq/fieldops/backend/internal/models"

// Actor is the authenticated caller a service acts on behalf of.
type Actor struct {
	ID   uint
	Role string
}

// IsTechnician reports whether the actor is limited to their own visits (§1.1).
func (a Actor) IsTechnician() bool {
	return a.Role == models.RoleTechnician
}

// CanSchedule reports whether the actor may manage sites and visits (§1.1).
func (a Actor) CanSchedule() bool {
	return a.Role == models.RoleAdmin || a.Role == models.RoleDispatcher
}
