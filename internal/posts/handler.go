package posts

import (
	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/middleware"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leg100/surl/v2"
)

type PostsHandler struct {
	svc  *PostsService
	mid  *middleware.AuthMiddleware
	sign *surl.Signer
}

func NewPostsHandler(sv *PostsService, md *middleware.AuthMiddleware, su *surl.Signer) *PostsHandler {
	return &PostsHandler{
		svc:  sv,
		mid:  md,
		sign: su,
	}
}

func (ph *PostsHandler) HanldeCreatePosts(c *gin.Context) {

	id, _ := middleware.GetUserIdFromContext(c)

	var dto CreatePostsDto
	if err := c.ShouldBind(&dto); err != nil {
		c.JSON(400, shared.NewErrorResponse(400, err.Error()))
		return
	}

	if len(dto.Media) < 1 {
		c.JSON(400, shared.NewErrorResponse(400, "please provide at least one media for posts!"))
		return
	}

	posts, err := ph.svc.CreatePosts(c.Request.Context(), id, dto)

	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "successfully creating the posts!", posts))

}

func (ph *PostsHandler) HandleServePostsMedia(c *gin.Context) {
	param := c.Param("filename")

	if param == "" {
		c.JSON(400, shared.NewErrorResponse(400, "invalid url parameter!"))
		return
	}

	path := ph.svc.ResolvePostsMediaPath(param)

	c.File(path)
}

func (ph *PostsHandler) HandleGetPostsById(c *gin.Context) {

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

	var userId uuid.UUID = uuid.Nil
	id, exist := middleware.GetUserIdFromContext(c)

	if exist {
		userId = id
	}

	data, getErr := ph.svc.GetPostsWithMediaById(c.Request.Context(), postId, userId)
	if getErr != nil {
		c.JSON(getErr.Code, getErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "successfully getting the posts data", data))
}

func (ph *PostsHandler) HandleDeletePosts(c *gin.Context) {

	param := c.Param("id")
	postsID, err := uuid.Parse(param)
	if err != nil {
		c.JSON(400, shared.NewErrorResponse(400, "invalid id parameter!"))
		return
	}

	id, _ := middleware.GetUserIdFromContext(c)

	delErr := ph.svc.DeletePosts(c.Request.Context(), postsID, id)

	if delErr != nil {
		c.JSON(delErr.Code, delErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "sucessfully deleting the posts", nil))
}

func (ph *PostsHandler) HandlePatchPosts(c *gin.Context) {

	param := c.Param("id")
	postsID, err := uuid.Parse(param)
	if err != nil {
		c.JSON(400, shared.NewErrorResponse(400, "invalid id parameter!"))
		return
	}

	id, _ := middleware.GetUserIdFromContext(c)
	var dto UpdatePostsDataDto
	if err := c.ShouldBindBodyWithJSON(&dto); err != nil {
		vldMsg, ok := utils.ParseValidationErrors(err)
		if !ok {
			c.JSON(400, shared.NewErrorResponse(400, "invalid request body"))
			return
		}
		c.JSON(400, shared.NewErrorResponse(400, vldMsg))
		return
	}

	updatedData, uptErr := ph.svc.UpdatePostsData(c.Request.Context(), id, postsID, dto)
	if uptErr != nil {
		c.JSON(uptErr.Code, uptErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "sucessfully updating the posts", updatedData))
}

func (ph *PostsHandler) HandleSharingPosts(c *gin.Context) {

	param := c.Param("id")
	postsID, err := uuid.Parse(param)
	if err != nil {
		c.JSON(400, shared.NewErrorResponse(400, "invalid id parameter!"))
		return
	}

	data, incErr := ph.svc.IncrementShareCount(c.Request.Context(), postsID)
	if incErr != nil {
		c.JSON(incErr.Code, incErr)
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "sucessfully sharing the posts", data))
}

func (ph *PostsHandler) RegisterRoutes(r gin.IRouter) {

	postsRoutes := r.Group("/posts")

	postsRoutes.GET("/id/:id", ph.HandleGetPostsById)
	postsRoutes.POST("/share/:id", ph.HandleSharingPosts)
	postsRoutes.GET("/media/:filename", ph.mid.RequireSignedUrl(ph.sign), ph.HandleServePostsMedia)

	authPosts := postsRoutes.Group("/")

	// middlewarenya
	authPosts.Use(ph.mid.AuthenticatedUserOnly())
	authPosts.Use(ph.mid.RequireUsersStatus([]generated.UserStatus{generated.UserStatusACTIVE}))

	authPosts.POST("/create", ph.HanldeCreatePosts)
	authPosts.DELETE("/delete/:id", ph.HandleDeletePosts)
	authPosts.PATCH("/update/:id", ph.HandlePatchPosts)

}
