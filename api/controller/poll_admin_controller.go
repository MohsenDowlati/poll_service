package controller

import (
	"errors"
	"strings"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

type PollAdminController struct {
	PollAdminUsecase domain.PollAdminUsecase
	SheetUsecase     domain.SheetUseCase
}

// Create adds a new poll for the given sheet.
// @Summary Create poll
// @Description Create a poll associated with a sheet.
// @Tags Polls (Admin)
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param sheet_id formData string true "Sheet identifier"
// @Param title formData string true "Poll title"
// @Param options formData []string true "Poll options"
// @Param poll_type formData string true "Poll type"
// @Param category formData []string true "Poll categories (repeat parameter for multiple values)"
// @Param description formData string false "Poll description"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/create [post]
func (pc *PollAdminController) Create(c *gin.Context) {
	var req domain.PollAdminRequest

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	categories := normalizeCategories(req.Category)
	if len(categories) == 0 {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "at least one category is required"})
		return
	}
	pollType, err := domain.ParsePollType(strings.ToLower(strings.TrimSpace(string(req.PollType))))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "invalid poll type"})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Options = normalizeOptions(req.Options)
	if req.Title == "" || len(req.Options) < pollType.MinOptions() {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "invalid poll title or options"})
		return
	}

	hexSheetID, err := primitive.ObjectIDFromHex(req.SheetID)

	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Invalid sheet ID"})
		return
	}
	if err := pc.authorizeSheet(c, hexSheetID.Hex()); err != nil {
		pc.writeAuthorizationError(c, err)
		return
	}

	var poll domain.Poll

	votes := make([]int, pollType.VoteSlots(len(req.Options)))
	if pollType == domain.PollTypeOpinion {
		votes = nil
	}

	poll = domain.Poll{
		ID:          primitive.NewObjectID(),
		SheetID:     hexSheetID,
		Title:       req.Title,
		Options:     req.Options,
		PollType:    pollType,
		Category:    categories,
		Participant: 0,
		Votes:       votes,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if pollType == domain.PollTypeOpinion {
		poll.Responses = []string{}
	}

	err = pc.PollAdminUsecase.CreatePoll(c, &poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, domain.SuccessResponse{Message: "poll created successfully"})
}

// Edit updates an existing poll.
// @Summary Update poll
// @Description Update fields of an existing poll.
// @Tags Polls (Admin)
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param id query string true "Poll identifier"
// @Param sheet_id formData string true "Sheet identifier"
// @Param title formData string true "Poll title"
// @Param options formData []string true "Poll options"
// @Param poll_type formData string true "Poll type"
// @Param category formData []string true "Poll categories (repeat parameter for multiple values)"
// @Param description formData string false "Poll description"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/edit [post]
func (pc *PollAdminController) Edit(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		id = c.Query("id")
	}
	if strings.TrimSpace(id) == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "poll id is required"})
		return
	}

	var req domain.PollAdminRequest

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	categories := normalizeCategories(req.Category)
	if len(categories) == 0 {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "at least one category is required"})
		return
	}
	pollType, err := domain.ParsePollType(strings.ToLower(strings.TrimSpace(string(req.PollType))))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "invalid poll type"})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Options = normalizeOptions(req.Options)
	if req.Title == "" || len(req.Options) < pollType.MinOptions() {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "invalid poll title or options"})
		return
	}

	hexSheetID, err := primitive.ObjectIDFromHex(req.SheetID)

	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Invalid sheet ID"})
		return
	}

	UID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Invalid ID"})
		return
	}
	existing, err := pc.PollAdminUsecase.GetByID(c, id)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocuments) || errors.Is(err, domain.ErrPollNotFound) {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Message: "poll not found"})
		} else {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: "failed to load poll"})
		}
		return
	}
	if err := pc.authorizeSheet(c, existing.SheetID.Hex()); err != nil {
		pc.writeAuthorizationError(c, err)
		return
	}
	if existing.SheetID != hexSheetID {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "poll sheet cannot be changed"})
		return
	}

	var poll domain.Poll

	votes := make([]int, pollType.VoteSlots(len(req.Options)))
	if pollType == domain.PollTypeOpinion {
		votes = nil
	}

	poll = domain.Poll{
		ID:          UID,
		SheetID:     hexSheetID,
		Title:       req.Title,
		Options:     req.Options,
		PollType:    pollType,
		Category:    categories,
		Participant: 0,
		Votes:       votes,
		Description: req.Description,
		UpdatedAt:   time.Now(),
	}

	if pollType == domain.PollTypeOpinion {
		poll.Responses = []string{}
	}

	err = pc.PollAdminUsecase.EditPoll(c, &poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "poll updated successfully"})
}

// GetBySheetID lists polls registered for a sheet.
// @Summary List polls for sheet
// @Description Retrieve all polls created for a sheet.
// @Tags Polls (Admin)
// @Produce json
// @Security BearerAuth
// @Param id query string true "Sheet identifier"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} domain.PollAdminListResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/admin/fetch [get]
func (pc *PollAdminController) GetBySheetID(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		id = c.Param("id")
	}

	if id == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "sheet id is required"})
		return
	}
	if err := pc.authorizeSheet(c, id); err != nil {
		pc.writeAuthorizationError(c, err)
		return
	}

	pagination := extractPagination(c)

	polls, total, err := pc.PollAdminUsecase.GetBySheetID(c, id, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	var responseItems []domain.PollAdminResponse

	for _, poll := range polls {
		var responses []string
		if poll.PollType == domain.PollTypeOpinion {
			if len(poll.Responses) > 0 {
				responses = append([]string(nil), poll.Responses...)
			} else {
				responses = []string{}
			}
		}

		responseItems = append(responseItems, domain.PollAdminResponse{
			ID:          poll.ID.Hex(),
			Title:       poll.Title,
			Options:     poll.Options,
			PollType:    poll.PollType,
			Category:    poll.Category,
			Participant: poll.Participant,
			Votes:       poll.Votes,
			Responses:   responses,
			Description: poll.Description,
		})
	}

	response := domain.PollAdminListResponse{
		Data:       responseItems,
		Pagination: domain.NewPaginationResult(pagination, total),
	}

	c.JSON(http.StatusOK, response)
}

// Delete removes a poll.
// @Summary Delete poll
// @Description Delete a poll by identifier.
// @Tags Polls (Admin)
// @Produce json
// @Security BearerAuth
// @Param id query string true "Poll identifier"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/delete [put]
func (pc *PollAdminController) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		id = c.Query("id")
	}

	if id == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "id is required"})
		return
	}
	poll, err := pc.PollAdminUsecase.GetByID(c, id)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocuments) || errors.Is(err, domain.ErrPollNotFound) {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Message: "poll not found"})
		} else {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: "failed to load poll"})
		}
		return
	}
	if err := pc.authorizeSheet(c, poll.SheetID.Hex()); err != nil {
		pc.writeAuthorizationError(c, err)
		return
	}

	err = pc.PollAdminUsecase.Delete(c, id)

	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrPollNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "poll deleted successfully"})

}

func (pc *PollAdminController) authorizeSheet(c *gin.Context, sheetID string) error {
	if domain.UserType(c.GetString("x-user-type")) == domain.SuperAdmin {
		return nil
	}
	if pc.SheetUsecase == nil {
		return errors.New("sheet authorization unavailable")
	}
	sheet, err := pc.SheetUsecase.GetByID(c, sheetID)
	if err != nil {
		return err
	}
	if sheet.UserID.Hex() != c.GetString("x-user-id") {
		return errForbidden
	}
	return nil
}

var errForbidden = errors.New("forbidden")

func (pc *PollAdminController) writeAuthorizationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errForbidden):
		c.JSON(http.StatusForbidden, domain.ErrorResponse{Message: "forbidden"})
	case errors.Is(err, mongodriver.ErrNoDocuments), errors.Is(err, domain.ErrSheetNotFound):
		c.JSON(http.StatusNotFound, domain.ErrorResponse{Message: "sheet not found"})
	default:
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: "failed to authorize sheet"})
	}
}

func normalizeOptions(values []string) []string {
	options := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		options = append(options, value)
	}
	return options
}
