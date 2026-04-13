package links

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"go-project-278/internal/db"
	"go-project-278/tools"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type LinkService interface {
	GetListLinks(ctx context.Context, arg db.GetListLinksParams) ([]db.Link, error)
	CreateLink(ctx context.Context, params db.NewLinkParams) (db.Link, error)
	GetLinkByID(ctx context.Context, id int64) (db.Link, error)
	UpdateLinkByID(ctx context.Context, params db.UpdateLinkByIDParams) (db.Link, error)
	DeleteLink(ctx context.Context, id int64) error
	CountLinks(ctx context.Context) (int64, error)
}

type dbService struct {
	q *db.Queries
}

func NewService(conn *sql.DB) LinkService {
	return &dbService{q: db.New(conn)}
}

func (s *dbService) GetListLinks(ctx context.Context, arg db.GetListLinksParams) ([]db.Link, error) {
	return s.q.GetListLinks(ctx, arg)
}

func (s *dbService) CreateLink(ctx context.Context, params db.NewLinkParams) (db.Link, error) {
	if params.ShortName == "" {
		params.ShortName = tools.GenerateShortName()
	}

	params.ShortUrl = os.Getenv("BASE_URL") + "/r/" + params.ShortName

	link, err := s.q.NewLink(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.Link{}, ErrConflict
		}

		return db.Link{}, err
	}

	return link, nil
}

func (s *dbService) GetLinkByID(ctx context.Context, id int64) (db.Link, error) {
	link, err := s.q.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Link{}, ErrNotFound
		}

		return db.Link{}, err
	}

	return link, nil
}

func (s *dbService) UpdateLinkByID(ctx context.Context, params db.UpdateLinkByIDParams) (db.Link, error) {
	params.ShortUrl = os.Getenv("BASE_URL") + "/r/" + params.ShortName

	link, err := s.q.UpdateLinkByID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Link{}, ErrNotFound
		}

		return db.Link{}, err
	}

	return link, nil
}

func (s *dbService) DeleteLink(ctx context.Context, id int64) error {
	if err := s.q.DeleteLink(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}

		return err
	}

	return nil
}

func (s *dbService) CountLinks(ctx context.Context) (int64, error) {
	count, err := s.q.CountLinks(ctx)
	if err != nil {
		return 0, err
	}

	return count, err
}
