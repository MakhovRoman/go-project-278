package visits

import (
	"errors"
	"fmt"
	"net/http"

	"go-project-278/internal/db"
	"go-project-278/internal/links"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	linkSvc  links.LinkService
	visitSvc VisitService
}

func NewHandler(linkSvc links.LinkService, visitSvc VisitService) *Handler {
	return &Handler{linkSvc: linkSvc, visitSvc: visitSvc}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/r/:code", h.handleRedirect)
	r.GET("/api/link_visits", h.handleGetListVisits)
}

func (h *Handler) handleRedirect(c *gin.Context) {
	code := c.Param("code")

	link, err := h.linkSvc.GetLinkByShortName(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, links.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, _ = h.visitSvc.CreateVisit(c.Request.Context(), db.CreateVisitParams{
		LinkID:    int32(link.ID),
		Ip:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
		Status:    http.StatusFound,
	})

	c.Redirect(http.StatusFound, link.OriginalUrl)
}

func (h *Handler) handleGetListVisits(c *gin.Context) {
	rangeStr := c.Query("range")
	if rangeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "range query param is required"})
		return
	}

	var offset, limit int64
	if _, err := fmt.Sscanf(rangeStr, "[%d,%d]", &offset, &limit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid range format, expected [offset,limit]"})
		return
	}

	list, err := h.visitSvc.GetListVisits(c.Request.Context(), db.GetListVisitsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total, err := h.visitSvc.CountVisits(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", offset, limit, total))
	c.JSON(http.StatusOK, list)
}
