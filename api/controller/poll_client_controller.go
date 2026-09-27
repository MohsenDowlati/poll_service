package controller

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type PollClientController struct {
	PollClientUsecse domain.PollClientUsecase
}

// Submit records votes for a poll.
// @Summary Submit poll votes
// @Description Submit votes for a poll.
// @Tags Polls
// @Accept json
// @Produce json
// @Param payload body domain.PollClientRequest true "Votes payload"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/submit [post]
func (pcc *PollClientController) Submit(c *gin.Context) {
	var req domain.PollClientRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}
	participantKey, cookieErr := c.Cookie("participant_id")
	if cookieErr != nil || participantKey == "" {
		buf := make([]byte, 24)
		if _, randErr := rand.Read(buf); randErr == nil {
			participantKey = hex.EncodeToString(buf)
			c.SetSameSite(http.SameSiteLaxMode)
			// The participant key is an opaque anti-duplicate identifier; it is not
			// needed by browser JavaScript and should not be exposed to XSS.
			c.SetCookie("participant_id", participantKey, 365*24*60*60, "/", "", c.Request.TLS != nil, true)
		}
	}
	req.ParticipantKey = participantKey
	req.ID = strings.TrimSpace(req.ID)
	if req.ID == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "poll id is required"})
		return
	}

	err = pcc.PollClientUsecse.SubmitVote(c, req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrNoVotesSubmitted) ||
			errors.Is(err, domain.ErrNoOpinionSubmitted) ||
			errors.Is(err, domain.ErrPhoneRequired) ||
			errors.Is(err, domain.ErrInvalidPhone) ||
			errors.Is(err, domain.ErrInvalidVote) {
			status = http.StatusBadRequest
		} else if errors.Is(err, domain.ErrDuplicateSubmission) {
			status = http.StatusConflict
		} else if errors.Is(err, domain.ErrPollNotPublished) || errors.Is(err, domain.ErrPollFinished) {
			status = http.StatusForbidden
		} else if errors.Is(err, mongo.ErrNoDocuments) {
			status = http.StatusNotFound
		}
		message := err.Error()
		if status >= 500 {
			message = "unable to submit vote"
		}
		c.JSON(status, domain.ErrorResponse{Message: message})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "Votes are submitted."})
}

// Fetch returns polls for a sheet.
// @Summary Get polls for sheet
// @Description Retrieve published polls for a sheet.
// @Tags Polls
// @Produce json
// @Param id query string true "Sheet identifier"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} domain.PollClientListResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/client/fetch [get]
func (pcc *PollClientController) Fetch(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		id = c.Param("id")
	}

	if id == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "sheet id is required"})
		return
	}
	id = strings.TrimSpace(id)

	pagination := extractPagination(c)

	sheet, err := pcc.PollClientUsecse.GetSheet(c, id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			status = http.StatusNotFound
		}
		message := "unable to load sheet"
		if status == http.StatusNotFound {
			message = "sheet not found"
		}
		c.JSON(status, domain.ErrorResponse{Message: message})
		return
	}

	if sheet.Status != domain.SheetStatusPublished {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{Message: "sheet not found"})
		return
	}

	polls, total, err := pcc.PollClientUsecse.GetBySheetID(c, id, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: "unable to retrieve polls"})
		return
	}

	result := make([]domain.PollClientResponse, 0, len(polls))

	for _, poll := range polls {
		result = append(result, domain.PollClientResponse{
			ID:          poll.ID.Hex(),
			Title:       poll.Title,
			Options:     poll.Options,
			PollType:    poll.PollType,
			Description: poll.Description,
		})
	}

	response := domain.PollClientListResponse{
		Data: result,
		Sheet: domain.PollClientSheetMeta{
			ID:              sheet.ID.Hex(),
			Title:           sheet.Title,
			IsPhoneRequired: sheet.IsPhoneRequired,
		},
		Pagination: domain.NewPaginationResult(pagination, total),
	}

	c.JSON(http.StatusOK, response)
}
