package usecase

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"time"
)

type pollClientUsecase struct {
	repository     domain.PollRepository
	contextTimeout time.Duration
}

func (p pollClientUsecase) SubmitVote(c context.Context, id string, votes []int) error {
	ctx, cancel := context.WithTimeout(c, p.contextTimeout)
	defer cancel()

	err := p.repository.SubmitVote(ctx, id, votes)
	return err
}

func (p pollClientUsecase) GetBySheetID(c context.Context, sheetID string) ([]domain.Poll, error) {
	ctx, cancel := context.WithTimeout(c, p.contextTimeout)
	defer cancel()

	var polls []domain.Poll
	polls, err := p.repository.GetPollBySheetID(ctx, sheetID)

	return polls, err
}

func NewPollClientUsecase(repo domain.PollRepository, timeout time.Duration) domain.PollClientUsecase {
	return &pollClientUsecase{
		repository:     repo,
		contextTimeout: timeout,
	}
}
