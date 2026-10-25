// Package userhttp is the HTTP adapter of the User context.
package userhttp

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/httpjson"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/problem"
	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

// Handler serves the routes of the User context.
type Handler struct {
	registerUser *app.RegisterUser
	getUser      *app.GetUser
	log          *slog.Logger
}

// NewHandler builds the handler.
func NewHandler(registerUser *app.RegisterUser, getUser *app.GetUser, log *slog.Logger) *Handler {
	return &Handler{registerUser: registerUser, getUser: getUser, log: log}
}

// Routes adds the routes of the context to the group of the API.
func (h *Handler) Routes(api *gin.RouterGroup) {
	api.POST("/users", h.register)
	api.GET("/users/:id", h.getByID)
}

type registerUserRequest struct {
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
}

type userResponse struct {
	ID           uuid.UUID `json:"id"`
	EmailAddress string    `json:"emailAddress"`
	DisplayName  string    `json:"displayName"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (h *Handler) register(c *gin.Context) {
	var request registerUserRequest
	if !httpjson.Decode(c, &request) {
		return
	}

	user, err := h.registerUser.Execute(c.Request.Context(), app.RegisterUserInput{
		EmailAddress: request.EmailAddress,
		DisplayName:  request.DisplayName,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toUserResponse(user))
}

func (h *Handler) getByID(c *gin.Context) {
	raw := c.Param("id")
	// Accept only the canonical form with dashes. uuid.Parse also accepts other forms, for example with braces.
	id, err := uuid.Parse(raw)
	if err != nil || id.String() != strings.ToLower(raw) {
		problem.Write(c, http.StatusBadRequest,
			"The ID is not valid. Use the ID of a User, in the format of a UUID, for example 123e4567-e89b-12d3-a456-426614174000.")
		return
	}

	user, err := h.getUser.Execute(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

func toUserResponse(user domain.User) userResponse {
	return userResponse{
		ID:           user.ID,
		EmailAddress: string(user.EmailAddress),
		DisplayName:  string(user.DisplayName),
		CreatedAt:    user.CreatedAt,
	}
}

// writeError changes an error of the use case into a problem detail.
func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		problem.Write(c, http.StatusNotFound, "The User does not exist. Check the ID.")
	case errors.Is(err, domain.ErrEmailTaken):
		problem.Write(c, http.StatusConflict, "The email address is already registered. Use a different email address.")
	case errors.Is(err, domain.ErrInvalidEmailAddress):
		problem.Write(c, http.StatusUnprocessableEntity,
			"The email address is not valid. Enter an address in the format name@example.com.")
	case errors.Is(err, domain.ErrEmptyDisplayName):
		problem.Write(c, http.StatusUnprocessableEntity, "The display name is empty. Enter a display name.")
	case errors.Is(err, domain.ErrInvalidDisplayName):
		problem.Write(c, http.StatusUnprocessableEntity,
			"The display name has a character that is not allowed. Remove the character, for example a new line.")
	case errors.Is(err, domain.ErrDisplayNameTooLong):
		problem.Write(c, http.StatusUnprocessableEntity,
			fmt.Sprintf("The display name is too long. Use %d characters or fewer.", domain.MaxDisplayNameLength))
	default:
		h.log.ErrorContext(c.Request.Context(), "The application did not complete the request.",
			slog.String("method", c.Request.Method), slog.String("route", c.FullPath()), slog.Any("error", err))
		problem.WriteInternalError(c)
	}
}
