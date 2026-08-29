package users

import (
	"strings"

	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/middleware"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

const maxAvatarSize = 1 << 20

type UserHandler struct {
	svc *UsersService
	mid *middleware.AuthMiddleware
}

func NewUserHandler(svc *UsersService, m *middleware.AuthMiddleware) *UserHandler {
	return &UserHandler{
		svc: svc,
		mid: m,
	}
}

func (uh *UserHandler) GetMyProfile(c *gin.Context) {

	id, _ := middleware.GetUserIdFromContext(c)

	data, err := uh.svc.GetUserDataWithProfileByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "successfully getting your profiles data", data))
}

func (uh *UserHandler) HandleGetUserProfile(c *gin.Context) {

	param := c.Param("username")

	if param == "" {
		c.JSON(400, shared.NewErrorResponse(400, "please provide a valid username!"))
		return
	}

	username := strings.TrimPrefix(param, "@")

	data, err := uh.svc.GetUsersDataWithProfilesByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "successfully getting the user data", data))
}

func (uh *UserHandler) HandlePatchMyProfile(c *gin.Context) {

	id, _ := middleware.GetUserIdFromContext(c)
	var req UpdateProfilesDto
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		vldMsg, ok := utils.ParseValidationErrors(err)
		if !ok {
			c.JSON(400, shared.NewErrorResponse(400, "invalid request body"))
			return
		}
		c.JSON(400, shared.NewErrorResponse(400, vldMsg))
		return
	}

	data, uptErr := uh.svc.UpdateUsersProfilesById(c.Request.Context(), id, req)
	if uptErr != nil {
		c.JSON(uptErr.Code, uptErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "sucessfully update your profiles", data))

}

func (uh *UserHandler) HandlePostOnBoarding(c *gin.Context) {

	id, _ := middleware.GetUserIdFromContext(c)
	var req UsernamePostsDto
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		vldMsg, ok := utils.ParseValidationErrors(err)
		if !ok {
			c.JSON(400, shared.NewErrorResponse(400, "invalid request body"))
			return
		}
		c.JSON(400, shared.NewErrorResponse(400, vldMsg))
		return
	}

	data, obErr := uh.svc.SubmitOnBoarding(c.Request.Context(), id, req.Username)
	if obErr != nil {
		c.JSON(obErr.Code, obErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "success submiting your onboarding", data))

}

func (uh *UserHandler) HandleUpdateMyUsername(c *gin.Context) {

	id, _ := middleware.GetUserIdFromContext(c)
	var req UsernamePostsDto
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		vldMsg, ok := utils.ParseValidationErrors(err)
		if !ok {
			c.JSON(400, shared.NewErrorResponse(400, "invalid request body"))
			return
		}
		c.JSON(400, shared.NewErrorResponse(400, vldMsg))
		return
	}

	newData, uptErr := uh.svc.UpdateUsername(c.Request.Context(), id, req.Username)
	if uptErr != nil {
		c.JSON(uptErr.Code, uptErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "success updating your username", newData))
}

func (uh *UserHandler) HandlePostProfilePicture(c *gin.Context) {

	id, _ := middleware.GetUserIdFromContext(c)

	file, err := c.FormFile("profile_picture")
	if err != nil {
		c.JSON(400, shared.NewErrorResponse(400, "missing the required image file! please provide a valid image"))
		return
	}

	if file.Size > maxAvatarSize {
		c.JSON(200, shared.NewErrorResponse(400, "file to large, maximum file size for avatar is 1 MB"))
		return
	}

	name, uptErr := uh.svc.UpdateProfilePicture(c.Request.Context(), id, file)
	if uptErr != nil {
		c.JSON(uptErr.Code, uptErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "profile picture update sucesss", name))

}

func (uh *UserHandler) HandlePostBannerPicture(c *gin.Context) {

	id, _ := middleware.GetUserIdFromContext(c)

	file, err := c.FormFile("banner_picture")
	if err != nil {
		c.JSON(400, shared.NewErrorResponse(400, "missing the required image file! please provide a valid image"))
		return
	}

	if file.Size > maxAvatarSize {
		c.JSON(200, shared.NewErrorResponse(400, "file to large, maximum file size for avatar is 1 MB"))
		return
	}

	name, uptErr := uh.svc.UpdateBannerPicture(c.Request.Context(), id, file)
	if uptErr != nil {
		c.JSON(uptErr.Code, uptErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "profile picture update sucesss", name))

}

func (uh *UserHandler) HandleCheckUsername(c *gin.Context) {

	param := c.Param("username")

	data, err := uh.svc.IsUsernameAvaible(c.Request.Context(), param)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "successfully checking the username", data))
}

func (uh *UserHandler) HandleUpdateUsersPrivacy(c *gin.Context) {

	id, _ := middleware.GetUserIdFromContext(c)

	var dto UpdateUsersPrvacy
	if err := c.ShouldBindBodyWithJSON(&dto); err != nil {
		vldMsg, ok := utils.ParseValidationErrors(err)
		if !ok {
			c.JSON(400, shared.NewErrorResponse(400, "invalid request body"))
			return
		}
		c.JSON(400, shared.NewErrorResponse(400, vldMsg))
		return
	}

	data, uptErr := uh.svc.UpdateAccountVisibility(c.Request.Context(), id, dto.Visibility)
	if uptErr != nil {
		c.JSON(uptErr.Code, shared.NewErrorResponse(uptErr.Code, uptErr))
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "successfully updating your privacy status", data))
}

func (uh *UserHandler) RegisterRoutes(r gin.IRouter) {
	users := r.Group("/users")

	spec := users.Group("/:username")
	spec.GET("/", uh.HandleGetUserProfile)
	spec.GET("/availability", uh.HandleCheckUsername)

	me := r.Group("/me")
	me.Use(uh.mid.AuthenticatedUserOnly())

	me.GET("/", uh.GetMyProfile)
	me.PATCH("/", uh.HandlePatchMyProfile)

	me.POST("/onboarding", uh.HandlePostOnBoarding)

	activeUsersRoutes := me.Group("/")
	activeUsersRoutes.Use(uh.mid.RequireUsersStatus([]generated.UserStatus{generated.UserStatusACTIVE}))
	activeUsersRoutes.PATCH("/username", uh.HandleUpdateMyUsername)

	activeUsersRoutes.POST("/profile-picture", uh.HandlePostProfilePicture)
	activeUsersRoutes.POST("/banner-picture", uh.HandlePostBannerPicture)

	activeUsersRoutes.PATCH("/account_privacy", uh.HandleUpdateUsersPrivacy)
}
