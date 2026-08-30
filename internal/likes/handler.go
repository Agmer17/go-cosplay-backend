package likes

import (
	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/middleware"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PostsLikesHandler struct {
	svc *PostLikesService
	mid *middleware.AuthMiddleware
}

func NewPostsLikesHandler(sv *PostLikesService, md *middleware.AuthMiddleware) *PostsLikesHandler {
	return &PostsLikesHandler{
		svc: sv,
		mid: md,
	}
}

func (plh *PostsLikesHandler) HanldeGetPostsLikes(c *gin.Context) {

	param := c.Param("id")
	if param == "" {
		c.JSON(400, shared.NewErrorResponse(400, "invalid id parameter"))
		return
	}

	postId, err := uuid.Parse(param)
	if err != nil {
		c.JSON(400, shared.NewErrorResponse(400, "invalid id parameter"))
		return
	}

	data, getErr := plh.svc.GetLikesDetailFromPosts(c.Request.Context(), postId)
	if getErr != nil {
		c.JSON(getErr.Code, getErr)
		return

	}

	c.JSON(200, shared.NewSuccessResponse(200, "successfully getting the likes data", data))
}

func (plh *PostsLikesHandler) HandleCreateLikes(c *gin.Context) {

	param := c.Param("id")
	if param == "" {
		c.JSON(400, shared.NewErrorResponse(400, "invalid id parameter"))
		return
	}

	postId, err := uuid.Parse(param)
	if err != nil {
		c.JSON(400, shared.NewErrorResponse(400, "invalid id parameter"))
		return
	}

	id, _ := middleware.GetUserIdFromContext(c)

	cr, lErr := plh.svc.CreateLikes(c.Request.Context(), postId, id)
	if lErr != nil {
		c.JSON(lErr.Code, lErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "created likes successfully", cr))

}

func (plh *PostsLikesHandler) HandleDeleteLikes(c *gin.Context) {

	param := c.Param("id")
	if param == "" {
		c.JSON(400, shared.NewErrorResponse(400, "invalid id parameter"))
		return
	}

	postId, err := uuid.Parse(param)
	if err != nil {
		c.JSON(400, shared.NewErrorResponse(400, "invalid id parameter"))
		return
	}

	id, _ := middleware.GetUserIdFromContext(c)

	delErr := plh.svc.Delete(c.Request.Context(), postId, id)
	if delErr != nil {
		c.JSON(delErr.Code, delErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "successfully deleting likes", nil))
}

func (plh *PostsLikesHandler) HandleGetAnotherUserLikes(c *gin.Context) {

	param := c.Param("id")
	if param == "" {
		c.JSON(400, shared.NewErrorResponse(400, "invalid id parameter"))
		return
	}

	userID, err := uuid.Parse(param)
	if err != nil {
		c.JSON(400, shared.NewErrorResponse(400, "invalid id parameter"))
		return
	}

	data, getErr := plh.svc.GetLikesWithPostsFromUsers(c.Request.Context(), userID)
	if getErr != nil {
		c.JSON(getErr.Code, getErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "successfully getting likes", data))

}

func (plh *PostsLikesHandler) RegisterRoutes(r gin.IRouter) {

	likesRoutes := r.Group("/likes")
	likesRoutes.GET("/posts/:id", plh.HanldeGetPostsLikes)

	protected := likesRoutes.Group("/")
	protected.Use(plh.mid.AuthenticatedUserOnly())

	protected.GET("/user/:id", plh.HandleGetAnotherUserLikes)

	protected.POST("/posts/:id",
		plh.mid.RequireUsersStatus([]generated.UserStatus{generated.UserStatusACTIVE}),
		plh.HandleCreateLikes)

	protected.DELETE("/posts/:id",
		plh.mid.RequireUsersStatus([]generated.UserStatus{generated.UserStatusACTIVE}),
		plh.HandleDeleteLikes)

}
