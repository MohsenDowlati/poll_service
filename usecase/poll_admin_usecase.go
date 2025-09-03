package usecase

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"time"
)

type pollAdminUsecase struct {
	repository     domain.PollRepository
	contextTimeout time.Duration
}

func (p pollAdminUsecase) Delete(c context.Context, id string) error {
	ctx, cancel := context.WithTimeout(c, p.contextTimeout)
	defer cancel()

	err := p.repository.Delete(ctx, id)

	return err
}

func (p pollAdminUsecase) CreatePoll(c context.Context, poll *domain.Poll) error {
	ctx, cancel := context.WithTimeout(c, p.contextTimeout)
	defer cancel()

	err := p.repository.Create(ctx, poll)

	return err
}

func (p pollAdminUsecase) GetBySheetID(c context.Context, sheetID string) ([]domain.Poll, error) {
	ctx, cancel := context.WithTimeout(c, p.contextTimeout)
	defer cancel()

	var polls []domain.Poll
	polls, err := p.repository.GetPollBySheetID(ctx, sheetID)

	return polls, err
}

func (p pollAdminUsecase) EditPoll(c context.Context, poll *domain.Poll) error {
	ctx, cancel := context.WithTimeout(c, p.contextTimeout)
	defer cancel()

	err := p.repository.EditPoll(ctx, poll)

	return err
}

func NewPollAdminUsecase(repository domain.PollRepository, timeout time.Duration) domain.PollAdminUsecase {
	return &pollAdminUsecase{
		repository:     repository,
		contextTimeout: timeout,
	}
}
