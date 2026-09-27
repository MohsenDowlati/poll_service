package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/internal/validation"
	"github.com/gin-gonic/gin"
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
	if sheet.Status == domain.SheetStatusFinished {
		return domain.ErrPollFinished
	}
	if sheet.Status != domain.SheetStatusPublished {
		return domain.ErrPollNotPublished
	}

	name := strings.TrimSpace(payload.UserName)
	phone := strings.TrimSpace(payload.UserPhone)

	if sheet.IsPhoneRequired && (name == "" || phone == "") {
		return domain.ErrPhoneRequired
	}

	if phone != "" && !validation.Phone(phone) {
		return domain.ErrInvalidPhone
	}

	submission := &domain.PollSubmission{
		Name:        name,
		Phone:       phone,
		Key:         submissionKey(c, phone, payload.ParticipantKey),
		SubmittedAt: time.Now(),
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
		if len(inputs) > 1 || len(inputs[0]) > 2000 {
			return domain.ErrInvalidVote
		}
		return p.repository.AppendOpinionResponse(ctx, payload.ID, inputs, submission)
	default:
		if err := validateVotes(poll, payload.Votes); err != nil {
			return err
		}
		return p.repository.SubmitVote(ctx, payload.ID, payload.Votes, submission)
	}
}

func validateVotes(poll domain.Poll, votes []int) error {
	if len(votes) != len(poll.Options) {
		return domain.ErrInvalidVote
	}

	selected := 0
	for _, vote := range votes {
		switch poll.PollType {
		case domain.PollTypeSlide:
			if vote < 0 || vote > 5 {
				return domain.ErrInvalidVote
			}
			if vote > 0 {
				selected++
			}
		default:
			if vote < 0 || vote > 1 {
				return domain.ErrInvalidVote
			}
			if vote == 1 {
				selected++
			}
		}
	}

	if selected == 0 {
		return domain.ErrNoVotesSubmitted
	}
	if poll.PollType == domain.PollTypeSingleChoice && selected != 1 {
		return domain.ErrInvalidVote
	}
	return nil
}

func submissionKey(c context.Context, phone, participantKey string) string {
	if phone != "" {
		sum := sha256.Sum256([]byte("phone:" + phone))
		return hex.EncodeToString(sum[:])
	}
	if participantKey != "" {
		sum := sha256.Sum256([]byte("participant:" + participantKey))
		return hex.EncodeToString(sum[:])
	}
	if ginContext, ok := c.(*gin.Context); ok {
		raw := ginContext.GetHeader("Idempotency-Key")
		if raw == "" {
			raw = ginContext.GetHeader("X-Participant-Key")
		}
		if raw == "" {
			return ""
		}
		sum := sha256.Sum256([]byte(raw))
		return hex.EncodeToString(sum[:])
	}
	return ""
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
