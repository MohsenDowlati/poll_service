package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type sheetUseCase struct {
	repository     domain.SheetRepository
	userRepository domain.UserRepository
	contextTimeout time.Duration
}

func (s sheetUseCase) GetByUserID(c context.Context, userID string, pagination domain.PaginationQuery) ([]domain.SheetListItem, int64, error) {
	ctx, cancel := context.WithTimeout(c, s.contextTimeout)
	defer cancel()

	sheets, total, err := s.repository.GetByUserID(ctx, userID, pagination)
	if err != nil {
		return nil, 0, err
	}

	items, err := s.buildSheetListItems(ctx, sheets)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
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

func (s sheetUseCase) GetAll(c context.Context, pagination domain.PaginationQuery) ([]domain.SheetListItem, int64, error) {
	ctx, cancel := context.WithTimeout(c, s.contextTimeout)
	defer cancel()

	sheets, total, err := s.repository.GetAll(ctx, pagination)
	if err != nil {
		return nil, 0, err
	}

	items, err := s.buildSheetListItems(ctx, sheets)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s sheetUseCase) Delete(c context.Context, id string) error {
	ctx, cancel := context.WithTimeout(c, s.contextTimeout)
	defer cancel()

	return s.repository.Delete(ctx, id)
}

func (s sheetUseCase) UpdateStatus(c context.Context, id string, status domain.SheetStatus, approvedBy primitive.ObjectID, approvedAt time.Time) error {
	ctx, cancel := context.WithTimeout(c, s.contextTimeout)
	defer cancel()

	return s.repository.UpdateStatus(ctx, id, status, approvedBy, approvedAt)
}

func (s sheetUseCase) buildSheetListItems(ctx context.Context, sheets []domain.Sheet) ([]domain.SheetListItem, error) {
	if len(sheets) == 0 {
		return []domain.SheetListItem{}, nil
	}

	ownerIDs := make(map[string]struct{}, len(sheets))
	for _, sheet := range sheets {
		if sheet.UserID.IsZero() {
			continue
		}
		ownerIDs[sheet.UserID.Hex()] = struct{}{}
	}

	ownerNames := make(map[string]string, len(ownerIDs))
	for ownerID := range ownerIDs {
		user, err := s.userRepository.GetByID(ctx, ownerID)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				continue
			}
			return nil, err
		}
		ownerNames[ownerID] = user.Name
	}

	items := make([]domain.SheetListItem, 0, len(sheets))
	for _, sheet := range sheets {
		item := domain.SheetListItem{
			ID:        sheet.ID.Hex(),
			Title:     sheet.Title,
			Venue:     sheet.Venue,
			Status:    sheet.Status,
			UpdatedAt: sheet.UpdatedAt,
		}

		if !sheet.UserID.IsZero() {
			if name, ok := ownerNames[sheet.UserID.Hex()]; ok {
				item.UserName = name
			}
		}

		if item.UpdatedAt.IsZero() {
			item.UpdatedAt = sheet.CreatedAt
		}

		items = append(items, item)
	}

	return items, nil
}

func NewSheetUseCase(repo domain.SheetRepository, userRepo domain.UserRepository, timeout time.Duration) domain.SheetUseCase {
	return &sheetUseCase{
		repository:     repo,
		userRepository: userRepo,
		contextTimeout: timeout,
	}
}
