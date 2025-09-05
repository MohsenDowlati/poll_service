package usecase

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"time"
)

type sheetUseCase struct {
	repository     domain.SheetRepository
	contextTimeout time.Duration
}

func (s sheetUseCase) GetByID(c context.Context, id string) (domain.Sheet, error) {
	ctx, cancel := context.WithTimeout(c, s.contextTimeout)
	defer cancel()

	return s.repository.GetByID(ctx, id)
}

func (s sheetUseCase) Create(c context.Context, sheet domain.Sheet) error {
	ctx, cancel := context.WithTimeout(c, s.contextTimeout)
	defer cancel()

	return s.repository.Create(ctx, sheet)
}

func (s sheetUseCase) GetAll(c context.Context) ([]domain.Sheet, error) {
	ctx, cancel := context.WithTimeout(c, s.contextTimeout)
	defer cancel()

	return s.repository.GetAll(ctx)
}

func (s sheetUseCase) Delete(c context.Context, id string) error {
	ctx, cancel := context.WithTimeout(c, s.contextTimeout)
	defer cancel()

	return s.repository.Delete(ctx, id)
}

func NewSheetUseCase(repo domain.SheetRepository, timeout time.Duration) domain.SheetUseCase {
	return &sheetUseCase{
		repository:     repo,
		contextTimeout: timeout,
	}
}
