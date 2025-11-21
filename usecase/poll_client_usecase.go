package usecase

import (
	"context"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/internal/validation"
	"strings"
	"time"
)

type pollClientUsecase struct {
	repository      domain.PollRepository
	sheetRepository domain.SheetRepository
	contextTimeout  time.Duration
}

func (p pollClientUsecase) SubmitVote(c context.Context, payload domain.PollClientRequest) error {
	ctx, cancel := context.WithTimeout(c, p.contextTimeout)
	defer cancel()

	poll, err := p.repository.GetByID(ctx, payload.ID)
	if err != nil {
		return err
	}

	sheet, err := p.sheetRepository.GetByID(ctx, poll.SheetID.Hex())
	if err != nil {
		return err
	}

	name := strings.TrimSpace(payload.UserName)
	phone := strings.TrimSpace(payload.UserPhone)

	if sheet.IsPhoneRequired && (name == "" || phone == "") {
		return domain.ErrPhoneRequired
	}

	if phone != "" && !validation.Phone(phone) {
		return domain.ErrInvalidPhone
	}

	var submission *domain.PollSubmission
	if name != "" || phone != "" {
		submission = &domain.PollSubmission{
			Name:        name,
			Phone:       phone,
			SubmittedAt: time.Now(),
		}
	}

	switch poll.PollType {
	case domain.PollTypeOpinion:
		inputs := make([]string, 0, len(payload.Inputs))
		for _, input := range payload.Inputs {
			value := strings.TrimSpace(input)
			if value != "" {
				inputs = append(inputs, value)
			}
		}
		if len(inputs) == 0 {
			return domain.ErrNoOpinionSubmitted
		}
		return p.repository.AppendOpinionResponse(ctx, payload.ID, inputs, submission)
	default:
		if len(payload.Votes) == 0 {
			return domain.ErrNoVotesSubmitted
		}
		return p.repository.SubmitVote(ctx, payload.ID, payload.Votes, submission)
	}
}

func (p pollClientUsecase) GetBySheetID(c context.Context, sheetID string, pagination domain.PaginationQuery) ([]domain.Poll, int64, error) {
	ctx, cancel := context.WithTimeout(c, p.contextTimeout)
	defer cancel()

	return p.repository.GetPollBySheetID(ctx, sheetID, pagination)
}

func NewPollClientUsecase(repo domain.PollRepository, sheetRepo domain.SheetRepository, timeout time.Duration) domain.PollClientUsecase {
	return &pollClientUsecase{
		repository:      repo,
		sheetRepository: sheetRepo,
		contextTimeout:  timeout,
	}
}

func (p pollClientUsecase) GetSheet(c context.Context, sheetID string) (domain.Sheet, error) {
	ctx, cancel := context.WithTimeout(c, p.contextTimeout)
	defer cancel()

	return p.sheetRepository.GetByID(ctx, sheetID)
}
