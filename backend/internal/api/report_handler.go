package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DailyReport returns the summary for one Europe/London day (§5).
func (h *Handlers) DailyReport(c *gin.Context) {
	day := c.Query("date")
	if day == "" {
		respondError(c, http.StatusBadRequest, "date is required (YYYY-MM-DD)")
		return
	}
	report, err := h.reports.Daily(c.Request.Context(), day)
	if err != nil {
		fail(c, h.logger, err)
		return
	}
	respond(c, http.StatusOK, newDailyReportView(report))
}
