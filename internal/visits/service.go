package visits

import (
	"context"
	"database/sql"
	"go-project-278/internal/db"
)

// VisitService описывает операции над записями о посещениях ссылок.
type VisitService interface {
	// CreateVisit сохраняет запись о посещении ссылки (IP, user-agent, referrer, статус ответа).
	CreateVisit(ctx context.Context, params db.CreateVisitParams) (db.LinkVisit, error)
	// GetListVisits возвращает страницу посещений с учётом LIMIT/OFFSET.
	GetListVisits(ctx context.Context, params db.GetListVisitsParams) ([]db.LinkVisit, error)
	// CountVisits возвращает общее количество записей о посещениях.
	CountVisits(ctx context.Context) (int64, error)
}

type dbService struct{ q *db.Queries }

// NewService создаёт реализацию VisitService поверх переданного соединения с БД.
func NewService(conn *sql.DB) VisitService {
	return &dbService{q: db.New(conn)}
}

func (d *dbService) CreateVisit(ctx context.Context, params db.CreateVisitParams) (db.LinkVisit, error) {
	lv, err := d.q.CreateVisit(ctx, params)
	if err != nil {
		return db.LinkVisit{}, err
	}

	return lv, nil
}

func (d *dbService) GetListVisits(ctx context.Context, params db.GetListVisitsParams) ([]db.LinkVisit, error) {
	list, err := d.q.GetListVisits(ctx, params)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (d *dbService) CountVisits(ctx context.Context) (int64, error) {
	count, err := d.q.CountVisits(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}
