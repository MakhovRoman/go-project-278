package visits

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-project-278/internal/db"
	"go-project-278/internal/links"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// stubLinkService реализует links.LinkService
type stubLinkService struct {
	getByShortNameFn func(ctx context.Context, shortName string) (db.Link, error)
}

func (s *stubLinkService) GetListLinks(ctx context.Context, arg db.GetListLinksParams) ([]db.Link, error) {
	return nil, nil
}
func (s *stubLinkService) CreateLink(ctx context.Context, params db.NewLinkParams) (db.Link, error) {
	return db.Link{}, nil
}
func (s *stubLinkService) GetLinkByID(ctx context.Context, id int64) (db.Link, error) {
	return db.Link{}, nil
}
func (s *stubLinkService) UpdateLinkByID(ctx context.Context, params db.UpdateLinkByIDParams) (db.Link, error) {
	return db.Link{}, nil
}
func (s *stubLinkService) DeleteLink(ctx context.Context, id int64) error { return nil }
func (s *stubLinkService) CountLinks(ctx context.Context) (int64, error)  { return 0, nil }
func (s *stubLinkService) GetLinkByShortName(ctx context.Context, shortName string) (db.Link, error) {
	if s.getByShortNameFn == nil {
		return db.Link{}, nil
	}
	return s.getByShortNameFn(ctx, shortName)
}

// stubVisitService реализует VisitService
type stubVisitService struct {
	createVisitFn func(ctx context.Context, params db.CreateVisitParams) (db.LinkVisit, error)
	getListFn     func(ctx context.Context, params db.GetListVisitsParams) ([]db.LinkVisit, error)
	countFn       func(ctx context.Context) (int64, error)
}

func (s *stubVisitService) CreateVisit(ctx context.Context, params db.CreateVisitParams) (db.LinkVisit, error) {
	if s.createVisitFn == nil {
		return db.LinkVisit{}, nil
	}
	return s.createVisitFn(ctx, params)
}
func (s *stubVisitService) GetListVisits(ctx context.Context, params db.GetListVisitsParams) ([]db.LinkVisit, error) {
	if s.getListFn == nil {
		return nil, nil
	}
	return s.getListFn(ctx, params)
}
func (s *stubVisitService) CountVisits(ctx context.Context) (int64, error) {
	if s.countFn == nil {
		return 0, nil
	}
	return s.countFn(ctx)
}

func newTestRouter(linkSvc links.LinkService, visitSvc VisitService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(linkSvc, visitSvc).Register(r)
	return r
}

// --- Redirect ---

func TestRedirect_OK(t *testing.T) {
	var capturedParams db.CreateVisitParams

	linkSvc := &stubLinkService{
		getByShortNameFn: func(ctx context.Context, shortName string) (db.Link, error) {
			assert.Equal(t, "google", shortName)
			return db.Link{ID: 1, OriginalUrl: "https://google.com", ShortName: "google"}, nil
		},
	}
	visitSvc := &stubVisitService{
		createVisitFn: func(ctx context.Context, params db.CreateVisitParams) (db.LinkVisit, error) {
			capturedParams = params
			return db.LinkVisit{}, nil
		},
	}
	r := newTestRouter(linkSvc, visitSvc)

	req := httptest.NewRequest(http.MethodGet, "/r/google", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://google.com", w.Header().Get("Location"))
	assert.Equal(t, int32(1), capturedParams.LinkID)
	assert.Equal(t, int32(http.StatusFound), capturedParams.Status)
}

func TestRedirect_NotFound(t *testing.T) {
	visitCalled := false

	linkSvc := &stubLinkService{
		getByShortNameFn: func(ctx context.Context, shortName string) (db.Link, error) {
			return db.Link{}, links.ErrNotFound
		},
	}
	visitSvc := &stubVisitService{
		createVisitFn: func(ctx context.Context, params db.CreateVisitParams) (db.LinkVisit, error) {
			visitCalled = true
			return db.LinkVisit{}, nil
		},
	}
	r := newTestRouter(linkSvc, visitSvc)

	req := httptest.NewRequest(http.MethodGet, "/r/notexist", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.False(t, visitCalled, "CreateVisit не должен вызываться если ссылка не найдена")
}

func TestRedirect_DBError(t *testing.T) {
	linkSvc := &stubLinkService{
		getByShortNameFn: func(ctx context.Context, shortName string) (db.Link, error) {
			return db.Link{}, errors.New("db error")
		},
	}
	r := newTestRouter(linkSvc, &stubVisitService{})

	req := httptest.NewRequest(http.MethodGet, "/r/google", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRedirect_VisitErrorDoesNotBlock(t *testing.T) {
	linkSvc := &stubLinkService{
		getByShortNameFn: func(ctx context.Context, shortName string) (db.Link, error) {
			return db.Link{ID: 1, OriginalUrl: "https://google.com"}, nil
		},
	}
	visitSvc := &stubVisitService{
		createVisitFn: func(ctx context.Context, params db.CreateVisitParams) (db.LinkVisit, error) {
			return db.LinkVisit{}, errors.New("analytics db down")
		},
	}
	r := newTestRouter(linkSvc, visitSvc)

	req := httptest.NewRequest(http.MethodGet, "/r/google", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://google.com", w.Header().Get("Location"))
}

// --- GetListVisits ---

func TestGetListVisits_OK(t *testing.T) {
	visitSvc := &stubVisitService{
		getListFn: func(ctx context.Context, params db.GetListVisitsParams) ([]db.LinkVisit, error) {
			assert.Equal(t, int32(0), params.Offset)
			assert.Equal(t, int32(10), params.Limit)
			return []db.LinkVisit{{ID: 1, LinkID: 1, Status: 302}}, nil
		},
		countFn: func(ctx context.Context) (int64, error) {
			return 42, nil
		},
	}
	r := newTestRouter(&stubLinkService{}, visitSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/link_visits?range=[0,10]", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits 0-10/42", w.Header().Get("Content-Range"))
	assert.Contains(t, w.Body.String(), `"id":1`)
}

func TestGetListVisits_Empty(t *testing.T) {
	r := newTestRouter(&stubLinkService{}, &stubVisitService{})

	req := httptest.NewRequest(http.MethodGet, "/api/link_visits?range=[0,10]", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits 0-10/0", w.Header().Get("Content-Range"))
}

func TestGetListVisits_NoRange(t *testing.T) {
	r := newTestRouter(&stubLinkService{}, &stubVisitService{})

	req := httptest.NewRequest(http.MethodGet, "/api/link_visits", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetListVisits_InvalidRange(t *testing.T) {
	r := newTestRouter(&stubLinkService{}, &stubVisitService{})

	req := httptest.NewRequest(http.MethodGet, "/api/link_visits?range=bad", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetListVisits_DBError(t *testing.T) {
	visitSvc := &stubVisitService{
		getListFn: func(ctx context.Context, params db.GetListVisitsParams) ([]db.LinkVisit, error) {
			return nil, errors.New("db error")
		},
	}
	r := newTestRouter(&stubLinkService{}, visitSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/link_visits?range=[0,10]", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
