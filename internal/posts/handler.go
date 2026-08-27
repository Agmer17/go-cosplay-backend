package posts

import (
	"github.com/Agmer17/go-cosplay-backend/internal/middleware"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/gin-gonic/gin"
)

type PostsHandler struct {
	svc *PostsService
	mid *middleware.AuthMiddleware
}

func NewPostsHandler(sv *PostsService, md *middleware.AuthMiddleware) *PostsHandler {
	return &PostsHandler{
		svc: sv,
		mid: md,
	}
}

func (ph *PostsHandler) HanldeCreatePosts(c *gin.Context) {

	id, _ := middleware.GetUserIdFromContext(c)

	var dto CreatePostsDto
	if err := c.ShouldBind(&dto); err != nil {
		c.JSON(400, shared.NewErrorResponse(400, err.Error()))
		return
	}

	posts, err := ph.svc.CreatePosts(c.Request.Context(), id, dto)

	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "successfully creating the posts!", posts))

}

func (ph *PostsHandler) RegisterRoutes(r gin.IRouter) {

	postsRoutes := r.Group("/posts")

	authPosts := postsRoutes.Group("/")
	authPosts.Use(ph.mid.AuthenticatedUserOnly())
	authPosts.POST("/create", ph.HanldeCreatePosts)
}
