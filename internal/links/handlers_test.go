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
	getListFn        func(ctx context.Context, arg db.GetListLinksParams) ([]db.Link, error)
	countFn          func(ctx context.Context) (int64, error)
	createLinkFn     func(ctx context.Context, params db.NewLinkParams) (db.Link, error)
	getByIDFn        func(ctx context.Context, id int64) (db.Link, error)
	updateByIDFn     func(ctx context.Context, params db.UpdateLinkByIDParams) (db.Link, error)
	deleteFn         func(ctx context.Context, id int64) error
	getByShortNameFn func(ctx context.Context, shortName string) (db.Link, error)
}

func (s *stubService) GetListLinks(ctx context.Context, arg db.GetListLinksParams) ([]db.Link, error) {
	if s.getListFn == nil {
		return nil, nil
	}
	return s.getListFn(ctx, arg)
}

func (s *stubService) CountLinks(ctx context.Context) (int64, error) {
	if s.countFn == nil {
		return 0, nil
	}
	return s.countFn(ctx)
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

func (s *stubService) GetLinkByShortName(ctx context.Context, shortName string) (db.Link, error) {
	if s.getByShortNameFn == nil {
		return db.Link{}, nil
	}
	return s.getByShortNameFn(ctx, shortName)
}

func TestGetListLinks(t *testing.T) {
	svc := &stubService{
		getListFn: func(ctx context.Context, arg db.GetListLinksParams) ([]db.Link, error) {
			return []db.Link{
				{ID: 1, OriginalUrl: "https://example.com", ShortName: "exmpl", ShortUrl: "https://short.io/r/exmpl"},
			}, nil
		},
		countFn: func(ctx context.Context) (int64, error) {
			return 1, nil
		},
	}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links?range=[0,10]", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":1`)
	assert.Contains(t, w.Body.String(), `"short_name":"exmpl"`)
	assert.Equal(t, "links 0-10/1", w.Header().Get("Content-Range"))
}

func TestGetListLinks_Empty(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links?range=[0,10]", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links 0-10/0", w.Header().Get("Content-Range"))
}

func TestGetListLinks_DBError(t *testing.T) {
	svc := &stubService{
		getListFn: func(ctx context.Context, arg db.GetListLinksParams) ([]db.Link, error) {
			return nil, errors.New("db error")
		},
	}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links?range=[0,10]", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetListLinks_NoRange(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetListLinks_InvalidRange(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links?range=bad", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
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

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "short name already in use")
}

func TestGetListLinks_CountError(t *testing.T) {
	svc := &stubService{
		getListFn: func(ctx context.Context, arg db.GetListLinksParams) ([]db.Link, error) {
			return []db.Link{}, nil
		},
		countFn: func(ctx context.Context) (int64, error) {
			return 0, errors.New("count error")
		},
	}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links?range=[0,10]", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
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
	assert.Contains(t, w.Body.String(), `"error":"invalid request"`)
}

func TestCreateLink_InvalidURL(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{"original_url":"not-a-url","short_name":"exmpl"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), `"errors"`)
	assert.Contains(t, w.Body.String(), `"original_url"`)
}

func TestCreateLink_ShortNameTooShort(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{"original_url":"https://example.com","short_name":"ab"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), `"errors"`)
}

func TestUpdateLink_InvalidURL(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{"original_url":"not-a-url","short_name":"exmpl"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/links/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), `"errors"`)
}

func TestUpdateLink_Conflict(t *testing.T) {
	svc := &stubService{
		updateByIDFn: func(ctx context.Context, params db.UpdateLinkByIDParams) (db.Link, error) {
			return db.Link{}, ErrConflict
		},
	}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{"original_url":"https://example.com","short_name":"exmpl"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/links/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "short name already in use")
}

func TestCreateLink_DBError(t *testing.T) {
	svc := &stubService{
		createLinkFn: func(ctx context.Context, params db.NewLinkParams) (db.Link, error) {
			return db.Link{}, errors.New("db error")
		},
	}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{"original_url":"https://example.com","short_name":"exmpl"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
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

func TestGetLinkByID_DBError(t *testing.T) {
	svc := &stubService{
		getByIDFn: func(ctx context.Context, id int64) (db.Link, error) {
			return db.Link{}, errors.New("db error")
		},
	}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/links/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
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

func TestUpdateLinkByID_DBError(t *testing.T) {
	svc := &stubService{
		updateByIDFn: func(ctx context.Context, params db.UpdateLinkByIDParams) (db.Link, error) {
			return db.Link{}, errors.New("db error")
		},
	}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{"original_url":"https://example.com","short_name":"exmpl"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/links/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateLinkByID_BadJSON(t *testing.T) {
	svc := &stubService{}
	r := newTestRouter(svc)

	body := bytes.NewBufferString(`{bad json`)
	req := httptest.NewRequest(http.MethodPut, "/api/links/1", body)
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

func TestDeleteLink_DBError(t *testing.T) {
	svc := &stubService{
		deleteFn: func(ctx context.Context, id int64) error {
			return errors.New("db error")
		},
	}
	r := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/links/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
