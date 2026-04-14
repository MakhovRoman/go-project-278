package links

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"reflect"

	"go-project-278/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = func() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})
	return v
}()

type createLinkPayload struct {
	OriginalUrl string `json:"original_url" validate:"required,url"`
	ShortName   string `json:"short_name"   validate:"omitempty,min=3,max=32"`
}

type updateLinkPayload struct {
	OriginalUrl string `json:"original_url" validate:"required,url"`
	ShortName   string `json:"short_name"   validate:"omitempty,min=3,max=32"`
}

// validationErrors форматирует ошибки валидатора в {"errors": {"field": "message"}}
func validationErrors(err error) gin.H {
	errs := make(map[string]string)
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, e := range ve {
			errs[e.Field()] = e.Error()
		}
	}
	return gin.H{"errors": errs}
}

// Handler обрабатывает HTTP-запросы для маршрутов /api/links.
type Handler struct {
	svc LinkService
}

// NewHandler создаёт Handler с переданным сервисом ссылок.
func NewHandler(svc LinkService) *Handler {
	return &Handler{svc: svc}
}

// Register регистрирует маршруты CRUD для ссылок на переданном роутере.
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
	var payload createLinkPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusUnprocessableEntity, validationErrors(err))
		return
	}

	link, err := h.svc.CreateLink(c.Request.Context(), db.NewLinkParams{
		OriginalUrl: payload.OriginalUrl,
		ShortName:   payload.ShortName,
	})
	if err != nil {
		if errors.Is(err, ErrConflict) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{"short_name": "short name already in use"}})
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

	var payload updateLinkPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusUnprocessableEntity, validationErrors(err))
		return
	}

	link, err := h.svc.UpdateLinkByID(c.Request.Context(), db.UpdateLinkByIDParams{
		ID:          id,
		OriginalUrl: payload.OriginalUrl,
		ShortName:   payload.ShortName,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		if errors.Is(err, ErrConflict) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{"short_name": "short name already in use"}})
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
