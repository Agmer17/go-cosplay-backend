package auth

import (
	"net/http"

	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/gin-gonic/gin"
)

const oneWeek = 7 * 24 * 60 * 60

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
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

	sessionToken, err := ah.service.AuthenticateGoogleLogin(c.Request.Context(), code)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"refresh_token",
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

func (ah *AuthHandler) RegisterRoutes(r gin.IRouter) {
	auth := r.Group("/auth")
	{
		auth.GET("/login/google", ah.HandleGoogleLogin)
		auth.GET("/google", ah.HandleGoogleCallback)
	}

	// privateAuth := auth.Group("/")
	// privateAuth.Use(middleware.AuthMiddlewareFromCookie())
	// {
	// 	privateAuth.GET("/logout", ah.LogoutHandler)
	// 	privateAuth.GET("/refresh-session", ah.HandleRefreshSession)
	// }
}
