package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SheetController struct {
	SheetuseCase        domain.SheetUseCase
	NotificationUsecase domain.NotificationUsecase
	PollUsecase         domain.PollAdminUsecase
}

// Create registers a new sheet.
// @Summary Create sheet
// @Description Create a new sheet (verified admin or super admin).
// @Tags Sheets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body domain.SheetCreateRequest true "Sheet creation payload"
// @Success 201 {object} domain.SheetCreateResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/sheet/create [post]
func (sc *SheetController) Create(c *gin.Context) {
	userID := c.GetString("x-user-id")
	userType := domain.UserType(c.GetString("x-user-type"))

	if userID == "" || (userType != domain.VerifiedAdmin && userType != domain.SuperAdmin) {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	var payload domain.SheetCreateRequest

	if err := c.ShouldBind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	title := strings.TrimSpace(payload.EffectiveTitle())
	if title == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "title is required"})
		return
	}

	venue := strings.TrimSpace(payload.Venue)
	if venue == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "venue is required"})
		return
	}

	validatedPolls := make([]domain.Poll, 0, len(payload.Polls))
	for idx, pollReq := range payload.Polls {
		pollTitle := strings.TrimSpace(pollReq.Title)
		if pollTitle == "" {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: fmt.Sprintf("poll %d title is required", idx+1)})
			return
		}

		trimmedOptions := make([]string, 0, len(pollReq.Options))
		for _, opt := range pollReq.Options {
			optionValue := strings.TrimSpace(opt)
			if optionValue != "" {
				trimmedOptions = append(trimmedOptions, optionValue)
			}
		}

		pollType, err := domain.ParsePollType(strings.ToLower(strings.TrimSpace(pollReq.PollType)))
		if err != nil {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: fmt.Sprintf("poll %d has invalid type", idx+1)})
			return
		}

		minOptions := pollType.MinOptions()
		if len(trimmedOptions) < minOptions {
			message := fmt.Sprintf("poll %d requires at least %d option", idx+1, minOptions)
			if minOptions > 1 {
				message += "s"
			}
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: message})
			return
		}

		categories := normalizeCategories(pollReq.Category)
		if len(categories) == 0 {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: fmt.Sprintf("poll %d requires at least one category", idx+1)})
			return
		}

		validatedPolls = append(validatedPolls, domain.Poll{
			Title:       pollTitle,
			Description: strings.TrimSpace(pollReq.Description),
			Options:     trimmedOptions,
			PollType:    pollType,
			Category:    categories,
		})
	}

	actorID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "invalid user identifier"})
		return
	}

	ownerID := actorID

	now := time.Now()

	sheet := domain.Sheet{
		ID:              primitive.NewObjectID(),
		UserID:          ownerID,
		Title:           title,
		Venue:           venue,
		IsPhoneRequired: payload.IsPhoneRequired,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if userType == domain.SuperAdmin {
		sheet.Status = domain.SheetStatusPublished
		sheet.ApprovedBy = actorID
		sheet.ApprovedAt = now
	} else {
		sheet.Status = domain.SheetStatusPending
	}

	if err = sc.SheetuseCase.Create(c, sheet); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	var (
		createdPolls   []domain.PollAdminResponse
		createdPollIDs []string
	)

	if len(validatedPolls) > 0 {
		if sc.PollUsecase == nil {
			_ = sc.SheetuseCase.Delete(c, sheet.ID.Hex())
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: "poll creation not configured"})
			return
		}

		for _, pollTemplate := range validatedPolls {
			poll := pollTemplate
			poll.ID = primitive.NewObjectID()
			poll.SheetID = sheet.ID
			poll.Participant = 0
			if poll.PollType == domain.PollTypeOpinion {
				poll.Votes = nil
				poll.Responses = []string{}
			} else {
				poll.Votes = make([]int, poll.PollType.VoteSlots(len(poll.Options)))
			}
			poll.CreatedAt = now
			poll.UpdatedAt = now

			if err = sc.PollUsecase.CreatePoll(c, &poll); err != nil {
				for _, pollID := range createdPollIDs {
					_ = sc.PollUsecase.Delete(c, pollID)
				}
				_ = sc.SheetuseCase.Delete(c, sheet.ID.Hex())
				c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
				return
			}

			createdPollIDs = append(createdPollIDs, poll.ID.Hex())
			createdPolls = append(createdPolls, domain.PollAdminResponse{
				ID:          poll.ID.Hex(),
				Title:       poll.Title,
				Options:     poll.Options,
				PollType:    poll.PollType,
				Category:    poll.Category,
				Participant: poll.Participant,
				Votes:       poll.Votes,
				Responses:   poll.Responses,
				Description: poll.Description,
			})
		}
	}

	if userType == domain.VerifiedAdmin && sc.NotificationUsecase != nil {
		if err = sc.NotificationUsecase.CreateForSheet(c, &sheet); err != nil {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
			return
		}
	}

	message := "sheet created!"
	if sheet.Status == domain.SheetStatusPending {
		message = "sheet submitted for approval"
	} else {
		message = "sheet published"
	}

	c.JSON(http.StatusCreated, domain.SheetCreateResponse{
		Message: message,
		Sheet:   sheet,
		Polls:   createdPolls,
	})
}

// Fetch lists sheets for the authenticated admin.
// @Summary List sheets
// @Description Retrieve sheets depending on admin role.
// @Tags Sheets
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} domain.SheetListResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/sheet/fetch [get]
func (sc *SheetController) Fetch(c *gin.Context) {
	userID := c.GetString("x-user-id")
	userType := domain.UserType(c.GetString("x-user-type"))

	if userID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	pagination := extractPagination(c)

	var sheets []domain.SheetListItem
	var total int64
	var err error

	if userType == domain.SuperAdmin {
		sheets, total, err = sc.SheetuseCase.GetAll(c, pagination)
	} else if userType == domain.VerifiedAdmin {
		sheets, total, err = sc.SheetuseCase.GetByUserID(c, userID, pagination)
	} else {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	response := domain.SheetListResponse{
		Data:       sheets,
		Pagination: domain.NewPaginationResult(pagination, total),
	}

	c.JSON(http.StatusOK, response)
}

func (sc *SheetController) FetchByID(c *gin.Context) {
	identifier := c.Param("id")
	if identifier == "" {
		identifier = c.Query("id")
	}

	if identifier == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "id is required"})
		return
	}

	sheet, err := sc.SheetuseCase.GetByID(c, identifier)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			status = http.StatusNotFound
		}
		c.JSON(status, domain.ErrorResponse{Message: err.Error()})
		return
	}

	userID := c.GetString("x-user-id")
	userType := domain.UserType(c.GetString("x-user-type"))

	if userID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	if userType != domain.SuperAdmin && sheet.UserID.Hex() != userID {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, sheet)
}

// Delete removes a sheet owned by the authenticated admin.
// @Summary Delete sheet
// @Description Delete a sheet by identifier.
// @Tags Sheets
// @Produce json
// @Security BearerAuth
// @Param id query string true "Sheet identifier"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/sheet/delete [put]
func (sc *SheetController) Delete(c *gin.Context) {
	identifier := c.Param("id")
	if identifier == "" {
		identifier = c.Query("id")
	}

	if identifier == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "id is required"})
		return
	}

	sheet, err := sc.SheetuseCase.GetByID(c, identifier)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			status = http.StatusNotFound
		}
		c.JSON(status, domain.ErrorResponse{Message: err.Error()})
		return
	}

	userID := c.GetString("x-user-id")
	userType := domain.UserType(c.GetString("x-user-type"))

	if userID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	if userType != domain.SuperAdmin && sheet.UserID.Hex() != userID {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	if err = sc.SheetuseCase.Delete(c, identifier); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "sheet deleted!"})
}

// Finish marks a sheet as finished.
// @Summary Finish sheet
// @Description Mark a sheet as finished (super admin or sheet owner).
// @Tags Sheets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query string true "Sheet identifier"
// @Param payload body domain.SheetStatusUpdateRequest true "Status payload (status must be 'finished')"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/sheet/finish [put]
func (sc *SheetController) Finish(c *gin.Context) {
	identifier := c.Param("id")
	if identifier == "" {
		identifier = c.Query("id")
	}

	if identifier == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "id is required"})
		return
	}

	var payload domain.SheetStatusUpdateRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	if payload.Status != domain.SheetStatusFinished {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "status must be 'finished'"})
		return
	}

	sheet, err := sc.SheetuseCase.GetByID(c, identifier)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			status = http.StatusNotFound
		}
		c.JSON(status, domain.ErrorResponse{Message: err.Error()})
		return
	}

	userID := c.GetString("x-user-id")
	userType := domain.UserType(c.GetString("x-user-type"))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	if userType != domain.SuperAdmin && sheet.UserID.Hex() != userID {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	if sheet.Status == domain.SheetStatusFinished {
		c.JSON(http.StatusOK, domain.SuccessResponse{Message: "sheet already finished"})
		return
	}

	actorID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "invalid user identifier"})
		return
	}

	now := time.Now()
	if err = sc.SheetuseCase.UpdateStatus(c, identifier, domain.SheetStatusFinished, actorID, now); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "sheet marked as finished"})
}
