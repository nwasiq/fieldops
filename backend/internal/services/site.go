package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/nwasiq/fieldops/backend/internal/models"
)

// SiteService manages customer sites (§2).
type SiteService struct {
	db *gorm.DB
}

// NewSiteService builds a SiteService.
func NewSiteService(db *gorm.DB) *SiteService {
	return &SiteService{db: db}
}

// SiteInput is the writable shape of a site.
type SiteInput struct {
	Name         string
	Address      string
	ContactPhone string
}

func (in *SiteInput) validate() error {
	in.Name = strings.TrimSpace(in.Name)
	in.Address = strings.TrimSpace(in.Address)
	in.ContactPhone = strings.TrimSpace(in.ContactPhone)
	switch {
	case in.Name == "":
		return invalidf("name is required")
	case in.Address == "":
		return invalidf("address is required")
	}
	return nil
}

// List returns live sites, paged, ordered by name.
func (s *SiteService) List(ctx context.Context, page Page) (PageResult[models.Site], error) {
	query := s.db.WithContext(ctx).Model(&models.Site{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PageResult[models.Site]{}, fmt.Errorf("count sites: %w", err)
	}
	var sites []models.Site
	if err := query.Order("name, id").Offset(page.Offset()).Limit(page.Size).Find(&sites).Error; err != nil {
		return PageResult[models.Site]{}, fmt.Errorf("list sites: %w", err)
	}
	return newPageResult(sites, total, page), nil
}

// Get returns one live site.
func (s *SiteService) Get(ctx context.Context, id uint) (*models.Site, error) {
	var site models.Site
	err := s.db.WithContext(ctx).First(&site, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, notFound("site")
	}
	if err != nil {
		return nil, fmt.Errorf("get site: %w", err)
	}
	return &site, nil
}

// Create adds a site.
func (s *SiteService) Create(ctx context.Context, in SiteInput) (*models.Site, error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	site := models.Site{Name: in.Name, Address: in.Address, ContactPhone: in.ContactPhone}
	if err := s.db.WithContext(ctx).Create(&site).Error; err != nil {
		return nil, fmt.Errorf("create site: %w", err)
	}
	return &site, nil
}

// Update replaces a site's fields.
func (s *SiteService) Update(ctx context.Context, id uint, in SiteInput) (*models.Site, error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	site, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	site.Name = in.Name
	site.Address = in.Address
	site.ContactPhone = in.ContactPhone
	if err := s.db.WithContext(ctx).Save(site).Error; err != nil {
		return nil, fmt.Errorf("update site: %w", err)
	}
	return site, nil
}

// Delete soft-deletes a site; its past visits remain readable (§2.1).
func (s *SiteService) Delete(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&models.Site{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete site: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return notFound("site")
	}
	return nil
}
