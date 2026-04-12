package links

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-project-278/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type stubService struct {
	getListFn    func(ctx context.Context) ([]db.Link, error)
	createLinkFn func(ctx context.Context, params db.NewLinkParams) (db.Link, error)
	getByIDFn    func(ctx context.Context, id int64) (db.Link, error)
	updateByIDFn func(ctx context.Context, params db.UpdateLinkByIDParams) (db.Link, error)
	deleteFn     func(ctx context.Context, id int64) error
}

func (s *stubService) GetListLinks(ctx context.Context) ([]db.Link, error) {
	if s.getListFn == nil {
		return nil, nil
	}
	return s.getListFn(ctx)
}

func (s *stubService) CreateLink(ctx context.Context, params db.NewLinkParams) (db.Link, error) {
	if s.createLinkFn == nil {
		return db.Link{}, nil
	}
	return s.createLinkFn(ctx, params)
}

func (s *stubService) GetLinkByID(ctx context.Context, id int64) (db.Link, error) {
	if s.getByIDFn == nil {
		return db.Link{}, nil
	}
	return s.getByIDFn(ctx, id)
}

func (s *stubService) UpdateLinkByID(ctx context.Context, params db.UpdateLinkByIDParams) (db.Link, error) {
	if s.updateByIDFn == nil {
		return db.Link{}, nil
	}
	return s.updateByIDFn(ctx, params)
}

func (s *stubService) DeleteLink(ctx context.Context, id int64) error {
	if s.deleteFn == nil {
		return nil
	}
	return s.deleteFn(ctx, id)
}

func newTestRouter(svc LinkService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(svc).Register(r)
	return r
}

func TestGetListLinks(t *testing.T) {
	svc := &stubService{
		getListFn: func(ctx context.Context) ([]db.Link, error) {
			return []db.Link{
				{ID: 1, OriginalUrl: "https://example.com", ShortName: "exmpl", ShortUrl: "https://short.io/r/exmpl"},
			}, nil
		},
	}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":1`)
	assert.Contains(t, w.Body.String(), `"short_name":"exmpl"`)
}

func TestGetListLinks_Empty(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetListLinks_DBError(t *testing.T) {
	svc := &stubService{
		getListFn: func(ctx context.Context) ([]db.Link, error) {
			return nil, errors.New("db error")
		},
	}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateLink(t *testing.T) {
	svc := &stubService{
		createLinkFn: func(ctx context.Context, params db.NewLinkParams) (db.Link, error) {
			assert.Equal(t, "https://example.com", params.OriginalUrl)
			assert.Equal(t, "exmpl", params.ShortName)
			return db.Link{ID: 1, OriginalUrl: params.OriginalUrl, ShortName: params.ShortName}, nil
		},
	}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{"original_url":"https://example.com","short_name":"exmpl"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"id":1`)
}

func TestCreateLink_Conflict(t *testing.T) {
	svc := &stubService{
		createLinkFn: func(ctx context.Context, params db.NewLinkParams) (db.Link, error) {
			return db.Link{}, ErrConflict
		},
	}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{"original_url":"https://example.com","short_name":"exmpl"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "short_name already exists")
}

func TestCreateLink_BadJSON(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{bad json`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetLinkByID(t *testing.T) {
	svc := &stubService{
		getByIDFn: func(ctx context.Context, id int64) (db.Link, error) {
			assert.Equal(t, int64(1), id)
			return db.Link{ID: 1, OriginalUrl: "https://example.com", ShortName: "exmpl"}, nil
		},
	}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":1`)
}

func TestGetLinkByID_NotFound(t *testing.T) {
	svc := &stubService{
		getByIDFn: func(ctx context.Context, id int64) (db.Link, error) {
			return db.Link{}, ErrNotFound
		},
	}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links/99", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetLinkByID_InvalidID(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateLinkByID(t *testing.T) {
	svc := &stubService{
		updateByIDFn: func(ctx context.Context, params db.UpdateLinkByIDParams) (db.Link, error) {
			assert.Equal(t, int64(1), params.ID)
			assert.Equal(t, "https://example.com/new", params.OriginalUrl)
			assert.Equal(t, "newname", params.ShortName)
			return db.Link{ID: params.ID, OriginalUrl: params.OriginalUrl, ShortName: params.ShortName}, nil
		},
	}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{"original_url":"https://example.com/new","short_name":"newname"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/links/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"short_name":"newname"`)
}

func TestUpdateLinkByID_NotFound(t *testing.T) {
	svc := &stubService{
		updateByIDFn: func(ctx context.Context, params db.UpdateLinkByIDParams) (db.Link, error) {
			return db.Link{}, ErrNotFound
		},
	}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{"original_url":"https://example.com","short_name":"exmpl"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/links/99", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateLinkByID_InvalidID(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{"original_url":"https://example.com","short_name":"exmpl"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/links/abc", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteLink(t *testing.T) {
	svc := &stubService{
		deleteFn: func(ctx context.Context, id int64) error {
			assert.Equal(t, int64(1), id)
			return nil
		},
	}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/links/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteLink_NotFound(t *testing.T) {
	svc := &stubService{
		deleteFn: func(ctx context.Context, id int64) error {
			return ErrNotFound
		},
	}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/links/99", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteLink_InvalidID(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/links/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
