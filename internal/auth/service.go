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
		return "", shared.NewErrorResponse(400, "google account email is not verified, please use another account!")
	}

	existingData, err := as.database.Query.GetUserAuthByProviderOpenID(
		ctx, generated.GetUserAuthByProviderOpenIDParams{
			Provider:       generated.AuthProviderGOOGLE,
			ProviderOpenid: googleUser.ID,
		})

	if err == nil {
		currentUser, getUserErr := as.database.Query.GetUserByID(ctx, existingData.ID)
		if getUserErr != nil {
			return "", shared.NewErrorResponse(500, "something wrong with the server right now, please try again another time")
		}

		return as.issueSession(ctx, existingData.ID, currentUser.Role, currentUser.Status)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return "", shared.NewErrorResponse(500, "something went wrong with the server")
	}

	var createdUser generated.User

	as.database.Transaction(ctx, func(q *generated.Queries) error {
		authData, err := q.CreateUserAuth(ctx, generated.CreateUserAuthParams{
			Provider:       generated.AuthProviderGOOGLE,
			ProviderOpenid: googleUser.ID,
			Email:          googleUser.Email,
		})

		if err != nil {
			return err
		}

		randStr, _ := utils.GenerateSecureString(12)

		userData, err := q.CreateUser(ctx, generated.CreateUserParams{
			ID:       authData.ID,
			Username: "USER-" + randStr,
			Role:     generated.UserRoleADMIN,
		})

		createdUser = userData

		if err != nil {
			return err
		}

		_, err = q.CreateProfile(ctx, generated.CreateProfileParams{
			UserID:      userData.ID,
			DisplayName: &googleUser.GivenName,
			AvatarUrl:   &googleUser.Picture,
		})

		if err != nil {
			return err
		}

		return nil
	})

	token, sessErr := as.issueSession(ctx, createdUser.ID, createdUser.Role, createdUser.Status)
	if sessErr != nil {
		return "", sessErr
	}

	return token, nil
}

func (as *AuthService) issueSession(ctx context.Context, userID uuid.UUID, role generated.UserRole, status generated.UserStatus) (string, *shared.ErrorResponse) {
	token, err := utils.GenerateSecureString(32)
	if err != nil {
		return "", shared.NewErrorResponse(500, "failed to generate token")
	}

	sess := shared.SessionsData{
		SessionId: token,
		UserId:    userID,
		Role:      string(role),
		Status:    string(status),
	}

	if err := as.CreateSessions(ctx, sess); err != nil {
		return "", shared.NewErrorResponse(500, "something went wrong with the server")
	}

	return token, nil
}

func (as *AuthService) RevokeSessions(ctx context.Context, token string) *shared.ErrorResponse {
	return as.ExpireSessions(ctx, token)
}
