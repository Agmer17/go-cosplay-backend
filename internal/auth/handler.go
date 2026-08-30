package auth

import (
	"net/http"

	"github.com/Agmer17/go-cosplay-backend/internal/middleware"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/gin-gonic/gin"
)

const oneWeek = 7 * 24 * 60 * 60

type AuthHandler struct {
	service *AuthService
	mid     *middleware.AuthMiddleware
}

func NewAuthHandler(service *AuthService, md *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		service: service,
		mid:     md,
	}
}

func (ah *AuthHandler) HandleGoogleLogin(c *gin.Context) {
	redirectURL, err := ah.service.AccuireGoogleLogin()
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (ah *AuthHandler) HandleGoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(400, shared.NewErrorResponse(400, "please provide a valid code exchange"))
		return
	}

	sessionToken, err := ah.service.AuthenticateGoogleLogin(c.Request.Context(), code)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"access_token",
		sessionToken,
		oneWeek,
		"/",
		"localhost",
		false,
		true,
	)

	// buat tetsting di postman or smth, masukin ini ke cookie
	// or smth, terus minta refresh session ke /auth/refresh-session biar dapet access tokennya
	// access tokennya di set ke header Bearer <token deez nut>
	// c.JSON(200, refreshToken)

	// kalo production ini uncomment duls
	c.JSON(200, shared.NewSuccessResponse(200, "successfully login with google", sessionToken))
}

func (ah *AuthHandler) HandleLogout(c *gin.Context) {
	access, err := c.Cookie("access_token")
	if err != nil {
		c.JSON(403, shared.NewErrorResponse(403, "you need to login to access this feature"))
		return
	}

	ah.service.RevokeSessions(c.Request.Context(), access)
	c.SetCookie("access_token", "", -1, "/", "localhost", false, true)
	c.JSON(200, shared.NewSuccessResponse(200, "successfully logout!", nil))

}

func (ah *AuthHandler) HandleGetMySessionData(c *gin.Context) {

	id, _ := middleware.GetUserIdFromContext(c)
	cookie, err := c.Cookie("access_token")
	if err != nil {
		c.JSON(401, shared.NewErrorResponse(401, "you need to login to access this feature"))
		return
	}

	c.JSON(200, shared.NewSuccessResponse(200, "successfully getting your session data", gin.H{
		"user_id":    id,
		"session_id": cookie,
	}))

}

func (ah *AuthHandler) RegisterRoutes(r gin.IRouter) {
	auth := r.Group("/auth")
	{
		auth.GET("/login/google", ah.HandleGoogleLogin)
		auth.GET("/google", ah.HandleGoogleCallback)
		auth.GET("/logout", ah.HandleLogout)
		auth.GET("/my-session", ah.mid.AuthenticatedUserOnly(), ah.HandleGetMySessionData)
	}

	// privateAuth := auth.Group("/")
	// privateAuth.Use(middleware.AuthMiddlewareFromCookie())
	// {
	// 	privateAuth.GET("/logout", ah.LogoutHandler)
	// 	privateAuth.GET("/refresh-session", ah.HandleRefreshSession)
	// }
}
