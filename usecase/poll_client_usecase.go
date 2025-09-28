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

func (p pollClientUsecase) GetBySheetID(c context.Context, sheetID string, pagination domain.PaginationQuery) ([]domain.Poll, int64, error) {
	ctx, cancel := context.WithTimeout(c, p.contextTimeout)
	defer cancel()

	return p.repository.GetPollBySheetID(ctx, sheetID, pagination)
}

func NewPollClientUsecase(repo domain.PollRepository, timeout time.Duration) domain.PollClientUsecase {
	return &pollClientUsecase{
		repository:     repo,
		contextTimeout: timeout,
	}
}
