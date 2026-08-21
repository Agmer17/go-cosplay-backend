package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Agmer17/go-cosplay-backend/internal/db"
	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthService struct {
	database    *db.Database
	oauthConfig *oauth2.Config
	rdb         *redis.Client
}

type CreateAuthServiceParams struct {
	GoogleAuthId string
	GoogleSecret string
	Database     *db.Database
	RedisClient  *redis.Client
}

func NewAuthService(params *CreateAuthServiceParams) *AuthService {
	return &AuthService{
		database: params.Database,
		oauthConfig: &oauth2.Config{
			ClientID:     params.GoogleAuthId,
			ClientSecret: params.GoogleSecret,
			RedirectURL:  "http://localhost/api/auth/google",
			Scopes:       []string{"email", "profile", "openid"},
			Endpoint:     google.Endpoint,
		},
		rdb: params.RedisClient,
	}
}

func (as *AuthService) AccuireGoogleLogin() (string, *shared.ErrorResponse) {
	rnd, err := utils.GenerateSecureString(32)
	if err != nil {
		return "", shared.NewErrorResponse(500, "something wrong with the server")
	}

	redirect := as.oauthConfig.AuthCodeURL(rnd)

	return redirect, nil
}

func (as *AuthService) createUsersAuth(
	ctx context.Context,
	data generated.CreateUserAuthParams,
) (generated.UsersAuth, *shared.ErrorResponse) {
	auth, err := as.database.Query.CreateUserAuth(ctx, data)
	if err != nil {
		return generated.UsersAuth{}, shared.NewErrorResponse(500, "something wrong while trying to create the account, please try again later")
	}
	return auth, nil
}

func (as *AuthService) AuthenticateGoogleLogin(ctx context.Context, code string) (string, *shared.ErrorResponse) {
	t, err := as.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", shared.NewErrorResponse(500, "something went wrong with the server")
	}

	client := as.oauthConfig.Client(ctx, t)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", shared.NewErrorResponse(500, "something went wrong with the server")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", shared.NewErrorResponse(500, "failed to fetch user data from google")
	}

	var googleUser googleUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return "", shared.NewErrorResponse(500, "failed to decode user data")
	}

	if !googleUser.VerifiedEmail {
		return "", shared.NewErrorResponse(400, "google account email is not verified")
	}

	existingData, err := as.database.Query.GetUserAuthByProviderOpenID(
		ctx, generated.GetUserAuthByProviderOpenIDParams{
			Provider:       generated.AuthProviderGOOGLE,
			ProviderOpenid: googleUser.ID,
		})

	if err == nil {
		// user sudah ada, langsung buat sesi
		return as.issueSession(ctx, existingData.ID, existingData.Role)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return "", shared.NewErrorResponse(500, "something went wrong with the server")
	}
	auth, cErr := as.createUsersAuth(
		ctx,
		generated.CreateUserAuthParams{
			Provider:       generated.AuthProviderGOOGLE,
			ProviderOpenid: googleUser.ID,
			Email:          googleUser.Email,
			Role:           generated.UserRoleADMIN,
		},
	)
	if cErr != nil {
		return "", cErr
	}

	return as.issueSession(ctx, auth.ID, auth.Role)
}

func (as *AuthService) issueSession(ctx context.Context, userID uuid.UUID, role generated.UserRole) (string, *shared.ErrorResponse) {
	token, err := utils.GenerateSecureString(32)
	if err != nil {
		return "", shared.NewErrorResponse(500, "failed to generate token")
	}

	sess := SessionsData{
		SessionId: token,
		UserId:    userID,
		Role:      string(role),
	}

	if err := as.CreateSessions(ctx, sess); err != nil {
		return "", shared.NewErrorResponse(500, "something went wrong with the server")
	}

	return token, nil
}
