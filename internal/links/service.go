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

// ErrNotFound возвращается когда запрошенная ссылка не найдена в базе данных.
var ErrNotFound = errors.New("not found")

// ErrConflict возвращается при попытке создать ссылку с уже существующим short_name.
var ErrConflict = errors.New("conflict")

// LinkService описывает операции над сокращёнными ссылками.
type LinkService interface {
	// GetListLinks возвращает страницу ссылок с учётом LIMIT/OFFSET.
	GetListLinks(ctx context.Context, arg db.GetListLinksParams) ([]db.Link, error)
	// CreateLink создаёт новую ссылку. Если ShortName пустой — генерирует автоматически.
	// Возвращает ErrConflict если short_name уже занят.
	CreateLink(ctx context.Context, params db.NewLinkParams) (db.Link, error)
	// GetLinkByID возвращает ссылку по ID. Возвращает ErrNotFound если не существует.
	GetLinkByID(ctx context.Context, id int64) (db.Link, error)
	// UpdateLinkByID обновляет original_url и short_name ссылки.
	// Возвращает ErrNotFound или ErrConflict при соответствующих ошибках.
	UpdateLinkByID(ctx context.Context, params db.UpdateLinkByIDParams) (db.Link, error)
	// DeleteLink удаляет ссылку по ID. Возвращает ErrNotFound если не существует.
	DeleteLink(ctx context.Context, id int64) error
	// CountLinks возвращает общее количество ссылок в базе данных.
	CountLinks(ctx context.Context) (int64, error)
	// GetLinkByShortName находит ссылку по short_name. Возвращает ErrNotFound если не существует.
	GetLinkByShortName(ctx context.Context, shortName string) (db.Link, error)
}

type dbService struct {
	q *db.Queries
}

// NewService создаёт реализацию LinkService поверх переданного соединения с БД.
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

func (s *dbService) GetLinkByShortName(ctx context.Context, shortName string) (db.Link, error) {
	link, err := s.q.GetLinkByShortName(ctx, shortName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Link{}, ErrNotFound
		}
		return db.Link{}, err
	}

	return link, nil
}
