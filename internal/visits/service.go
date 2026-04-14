package visits

import (
	"context"
	"database/sql"
	"go-project-278/internal/db"
)

type VisitService interface {
	CreateVisit(ctx context.Context, params db.CreateVisitParams) (db.LinkVisit, error)
	GetListVisits(ctx context.Context, params db.GetListVisitsParams) ([]db.LinkVisit, error)
	CountVisits(ctx context.Context) (int64, error)
}

type dbService struct{ q *db.Queries }

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
