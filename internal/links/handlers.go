package links

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"go-project-278/internal/db"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc LinkService
}

func NewHandler(svc LinkService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/api/links", h.handleGetListLinks)
	r.POST("/api/links", h.handleCreateLink)
	r.GET("/api/links/:id", h.handleGetLinkByID)
	r.PUT("/api/links/:id", h.handleUpdateLinkByID)
	r.DELETE("/api/links/:id", h.handleDelete)
}

func (h *Handler) handleGetListLinks(c *gin.Context) {
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

	query := db.GetListLinksParams{
		Offset: int32(offset),
		Limit:  int32(limit),
	}

	list, err := h.svc.GetListLinks(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total, err := h.svc.CountLinks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", offset, limit, total))

	c.JSON(http.StatusOK, list)
}

func (h *Handler) handleCreateLink(c *gin.Context) {
	var req db.NewLinkParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.svc.CreateLink(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "short_name already exists"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, link)
}

func (h *Handler) handleGetLinkByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.svc.GetLinkByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, link)
}

func (h *Handler) handleUpdateLinkByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req db.UpdateLinkByIDParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = id

	link, err := h.svc.UpdateLinkByID(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, link)
}

func (h *Handler) handleDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.DeleteLink(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
