package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Agmer17/go-cosplay-backend/internal/db"
	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const loginSessionExpireTime = 7 * 24 * time.Hour

type AuthSessionStore interface {
	IssueSession(ctx context.Context, sess *shared.SessionsData, ttl time.Duration) error
	RevokeSession(ctx context.Context, key string) error
}

type UserDataRedisStore interface {
	Cache(ctx context.Context, data shared.UserCredential) error
}

type AuthService struct {
	database      *db.Database
	oauthConfig   *oauth2.Config
	sessionStore  AuthSessionStore
	userDataCache UserDataRedisStore
}

type CreateAuthServiceParams struct {
	GoogleAuthId string
	GoogleSecret string
	Database     *db.Database
	AuthSession  AuthSessionStore
	UserCache    UserDataRedisStore
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
		sessionStore:  params.AuthSession,
		userDataCache: params.UserCache,
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
			fmt.Println("ERROR : ", getUserErr)
			return "", shared.NewErrorResponse(500, "something wrong with the server right now, please try again another time")
		}

		as.userDataCache.Cache(ctx, shared.UserCredential{
			UserId: currentUser.ID,
			Status: string(currentUser.Status),
			Role:   string(currentUser.Role),
		})
		return as.issueSession(ctx, existingData.ID)
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

	token, sessErr := as.issueSession(ctx, createdUser.ID)
	if sessErr != nil {
		return "", sessErr
	}

	as.userDataCache.Cache(ctx, shared.UserCredential{
		UserId: createdUser.ID,
		Status: string(createdUser.Status),
		Role:   string(createdUser.Role),
	})

	return token, nil
}

func (as *AuthService) issueSession(ctx context.Context, userID uuid.UUID) (string, *shared.ErrorResponse) {
	token, err := utils.GenerateSecureString(32)
	if err != nil {
		return "", shared.NewErrorResponse(500, "failed to generate token")
	}

	sess := shared.SessionsData{
		SessionId: token,
		UserId:    userID,
	}

	if err := as.sessionStore.IssueSession(ctx, &sess, loginSessionExpireTime); err != nil {
		return "", shared.NewErrorResponse(500, "something went wrong with the server")
	}

	return token, nil
}

func (as *AuthService) RevokeSessions(ctx context.Context, token string) *shared.ErrorResponse {
	err := as.sessionStore.RevokeSession(ctx, token)

	if err != nil {
		return shared.NewErrorResponse(500, "something wrong while trying to remove your session, please try again another time")
	}

	return nil
}
