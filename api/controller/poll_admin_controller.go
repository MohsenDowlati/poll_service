package controller

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"
)

type PollAdminController struct {
	PollAdminUsecase domain.PollAdminUsecase
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

	hexSheetID, err := primitive.ObjectIDFromHex(req.SheetID)

	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Invalid sheet ID"})
		return
	}

	var poll domain.Poll

	votes := make([]int, req.PollType.VoteSlots(len(req.Options)))
	if req.PollType == domain.PollTypeOpinion {
		votes = nil
	}

	poll = domain.Poll{
		ID:          primitive.NewObjectID(),
		SheetID:     hexSheetID,
		Title:       req.Title,
		Options:     req.Options,
		PollType:    req.PollType,
		Category:    categories,
		Participant: 0,
		Votes:       votes,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if req.PollType == domain.PollTypeOpinion {
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

	var poll domain.Poll

	votes := make([]int, req.PollType.VoteSlots(len(req.Options)))
	if req.PollType == domain.PollTypeOpinion {
		votes = nil
	}

	poll = domain.Poll{
		ID:          UID,
		SheetID:     hexSheetID,
		Title:       req.Title,
		Options:     req.Options,
		PollType:    req.PollType,
		Category:    categories,
		Participant: 0,
		Votes:       votes,
		Description: req.Description,
		UpdatedAt:   time.Now(),
	}

	if req.PollType == domain.PollTypeOpinion {
		poll.Responses = []string{}
	}

	err = pc.PollAdminUsecase.EditPoll(c, &poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
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

	pagination := extractPagination(c)

	polls, total, err := pc.PollAdminUsecase.GetBySheetID(c, id, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	var responseItems []domain.PollAdminResponse

	for _, poll := range polls {
		responseItems = append(responseItems, domain.PollAdminResponse{
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

	err := pc.PollAdminUsecase.Delete(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "poll deleted successfully"})

}
