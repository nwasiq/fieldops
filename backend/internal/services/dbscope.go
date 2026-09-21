package services

import "gorm.io/gorm"

// UnscopedUserAssoc preloads a User association without the soft-delete scope,
// so a departed user still resolves wherever they are shown for attribution
// (§1.2, §4.3, §5.2). A default-scoped preload would leave the association as
// its zero value and the name would render as "unknown".
func UnscopedUserAssoc(db *gorm.DB) *gorm.DB {
	return db.Unscoped()
}

// UnscopedSiteAssoc preloads a Site association without the soft-delete scope,
// so a deleted site's past visits stay readable (§2.1).
func UnscopedSiteAssoc(db *gorm.DB) *gorm.DB {
	return db.Unscoped()
}
