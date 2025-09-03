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

func (pc *PollAdminController) Create(c *gin.Context) {
	var req domain.PollAdminRequest

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	hexSheetID, err := primitive.ObjectIDFromHex(req.SheetID)

	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Invalid sheet ID"})
		return
	}

	var poll domain.Poll

	poll = domain.Poll{
		ID:          primitive.NewObjectID(),
		SheetID:     hexSheetID,
		Title:       req.Title,
		Options:     req.Options,
		PollType:    req.PollType,
		Participant: 0,
		Votes:       make([]int, len(req.Options)),
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = pc.PollAdminUsecase.CreatePoll(c, &poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, domain.SuccessResponse{Message: "poll created successfully"})
}

func (pc *PollAdminController) Edit(c *gin.Context) {
	id := c.Param("id")

	var req domain.PollAdminRequest

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
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

	poll = domain.Poll{
		ID:          UID,
		SheetID:     hexSheetID,
		Title:       req.Title,
		Options:     req.Options,
		PollType:    req.PollType,
		Participant: 0,
		Votes:       make([]int, len(req.Options)),
		Description: req.Description,
		UpdatedAt:   time.Now(),
	}

	err = pc.PollAdminUsecase.EditPoll(c, &poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "poll updated successfully"})
}

func (pc *PollAdminController) GetBySheetID(c *gin.Context) {
	id := c.Param("id")

	var polls []domain.Poll

	polls, err := pc.PollAdminUsecase.GetBySheetID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	var response []domain.PollAdminResponse

	for _, poll := range polls {
		response = append(response, domain.PollAdminResponse{
			ID:          poll.ID.Hex(),
			Title:       poll.Title,
			Options:     poll.Options,
			PollType:    poll.PollType,
			Participant: poll.Participant,
			Votes:       poll.Votes,
			Description: poll.Description,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (pc *PollAdminController) Delete(c *gin.Context) {
	id := c.Param("id")

	err := pc.PollAdminUsecase.Delete(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "poll deleted successfully"})

}
