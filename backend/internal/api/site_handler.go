package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nwasiq/fieldops/backend/internal/middleware"
	"github.com/nwasiq/fieldops/backend/internal/services"
)

type siteRequest struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	ContactPhone string `json:"contact_phone"`
}

func (r siteRequest) input() services.SiteInput {
	return services.SiteInput{Name: r.Name, Address: r.Address, ContactPhone: r.ContactPhone}
}

// ListSites pages through live sites.
func (h *Handlers) ListSites(c *gin.Context) {
	result, err := h.sites.List(c.Request.Context(), pageFromQuery(c))
	if err != nil {
		fail(c, h.logger, err)
		return
	}
	sites := make([]siteView, 0, len(result.Items))
	for i := range result.Items {
		sites = append(sites, newSiteView(&result.Items[i]))
	}
	respond(c, http.StatusOK, siteListView{Sites: sites, pageView: newPageView(result)})
}

// GetSite returns one site.
func (h *Handlers) GetSite(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	site, err := h.sites.Get(c.Request.Context(), id)
	if err != nil {
		fail(c, h.logger, err)
		return
	}
	respond(c, http.StatusOK, newSiteView(site))
}

// CreateSite adds a site.
func (h *Handlers) CreateSite(c *gin.Context) {
	var req siteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	site, err := h.sites.Create(c.Request.Context(), req.input())
	if err != nil {
		fail(c, h.logger, err)
		return
	}
	middleware.SetAuditResource(c, site.ID)
	respond(c, http.StatusCreated, newSiteView(site))
}

// UpdateSite replaces a site's fields.
func (h *Handlers) UpdateSite(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	var req siteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	site, err := h.sites.Update(c.Request.Context(), id, req.input())
	if err != nil {
		fail(c, h.logger, err)
		return
	}
	respond(c, http.StatusOK, newSiteView(site))
}

// DeleteSite soft-deletes a site.
func (h *Handlers) DeleteSite(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.sites.Delete(c.Request.Context(), id); err != nil {
		fail(c, h.logger, err)
		return
	}
	respond(c, http.StatusOK, gin.H{"id": id})
}
