package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/Agmer17/go-cosplay-backend/internal/db"
	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const avatarFolder = "avatar"
const bannerFolder = "banner"
const publicUploadsUrl = "http://localhost/uploads/public/"

type RedisUserCache interface {
	Cache(ctx context.Context, data shared.UserCredential) error
}

type UsersService struct {
	database      *db.Database
	serverStorage *storage.FileStorage
	userCache     RedisUserCache
}

func NewUsersService(db *db.Database, storage *storage.FileStorage, cache RedisUserCache) *UsersService {

	return &UsersService{
		database:      db,
		serverStorage: storage,
		userCache:     cache,
	}
}

func (us *UsersService) GetUserDataWithProfileByID(ctx context.Context, id uuid.UUID) (generated.GetUserWithProfileByIDRow, *shared.ErrorResponse) {

	data, err := us.database.Query.GetUserWithProfileByID(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return generated.GetUserWithProfileByIDRow{}, shared.NewErrorResponse(404, "account not found")
		}

		return generated.GetUserWithProfileByIDRow{}, shared.NewErrorResponse(500, "something wrong with the server")

	}

	data.AvatarUrl = AppendUploadsUrl(data.AvatarUrl)
	data.BannerUrl = AppendUploadsUrl(data.BannerUrl)

	return data, nil

}

func (us *UsersService) GetUsersDataWithProfilesByUsername(
	ctx context.Context, username string) (generated.GetPublicProfileByUsernameRow,
	*shared.ErrorResponse) {
	data, err := us.database.Query.GetPublicProfileByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return generated.GetPublicProfileByUsernameRow{},
				shared.NewErrorResponse(404, "profiles not found")
		}
		return generated.GetPublicProfileByUsernameRow{},
			shared.NewErrorResponse(500, "something was wrong with the server, please try again another time")
	}

	data.AvatarUrl = AppendUploadsUrl(data.AvatarUrl)
	data.BannerUrl = AppendUploadsUrl(data.BannerUrl)
	return data, nil
}

func (us *UsersService) UpdateUsersProfilesById(ctx context.Context, id uuid.UUID, dto UpdateProfilesDto) (generated.Profile, *shared.ErrorResponse) {

	var socialLinks json.RawMessage
	if dto.SocialLinks != nil {
		socialLinks = *dto.SocialLinks
	}

	newData, err := us.database.Query.UpdateProfile(ctx, generated.UpdateProfileParams{
		DisplayName: dto.DisplayName,
		Bio:         dto.Bio,
		SocialLinks: socialLinks,
		CosplayTags: dto.CosplayTags,
		UserID:      id,
	})
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return generated.Profile{}, shared.NewErrorResponse(404, "profile not found")
		}

		return generated.Profile{}, shared.NewErrorResponse(500,
			"something wrong while trying to update the profile, please try again another time")

	}

	newData.AvatarUrl = AppendUploadsUrl(newData.AvatarUrl)
	newData.BannerUrl = AppendUploadsUrl(newData.BannerUrl)

	return newData, nil
}

func (us *UsersService) UpdateProfilePicture(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (string, *shared.ErrorResponse) {
	oldData, err := us.database.Query.GetProfileByUserID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", shared.NewErrorResponse(404, "account was not found")

		}
		return "", shared.NewErrorResponse(500, "something wrong while trying to save files! please try again another time")
	}

	saved, _, err := us.serverStorage.SavePublicFile(file, avatarFolder)
	if err != nil {
		return "", shared.NewErrorResponse(500, "something wrong while trying to save files! please try again another time")
	}

	avatarUrl := avatarFolder + "/" + saved

	uptErr := us.database.Query.UpdateProfileAvatar(ctx, generated.UpdateProfileAvatarParams{
		UserID:    id,
		AvatarUrl: &avatarUrl,
	})

	if uptErr != nil {
		// edge case kalo akun keburu ilang tapi somehow sessionnya masih ada
		if errors.Is(uptErr, pgx.ErrNoRows) {
			return "", shared.NewErrorResponse(404, "account was not found")

		}

		us.serverStorage.DeletePublicFile(saved, avatarFolder)
		return "", shared.NewErrorResponse(500, "something wrong while trying to updating data! please try again another time")
	}

	if !strings.HasPrefix(*oldData.AvatarUrl, "https://") {
		filename := strings.Split(*oldData.AvatarUrl, "/")
		us.serverStorage.DeletePublicFile(filename[1], avatarFolder)
	}

	res := AppendUploadsUrl(&avatarUrl)

	return *res, nil
}

func (us *UsersService) UpdateBannerPicture(
	ctx context.Context,
	id uuid.UUID,
	bann *multipart.FileHeader) (string, *shared.ErrorResponse) {

	sv, _, err := us.serverStorage.SavePublicFile(bann, bannerFolder)

	if err != nil {
		return "", shared.NewErrorResponse(500, "something wrong while trying to save the banner image, please try again another time")
	}

	bannerUrl := bannerFolder + "/" + sv

	err = us.database.Query.UpdateProfileBanner(ctx, generated.UpdateProfileBannerParams{
		BannerUrl: &bannerUrl,
		UserID:    id,
	})

	if err != nil {
		us.serverStorage.DeletePublicFile(sv, bannerFolder)
		return "", shared.NewErrorResponse(500, "something wrong while trying to update data!")
	}

	resBanner := AppendUploadsUrl(&bannerUrl)
	return *resBanner, nil

}

func (us *UsersService) UpdateUsername(ctx context.Context, id uuid.UUID, newUsername string) (string, *shared.ErrorResponse) {

	oldData, err := us.database.Query.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", shared.NewErrorResponse(404, "account not found!")
		}

		return "", shared.NewErrorResponse(500, "something wrong with the server while trying to update your username")
	}

	if oldData.Username == newUsername {
		return "", shared.NewErrorResponse(400, "username can't be the same as before!")
	}

	exist, err := us.database.Query.IsUsernameTaken(ctx, newUsername)
	if err != nil {
		return "", shared.NewErrorResponse(500, "something wrong with the server while trying to update the username")
	}

	if exist {
		return "", shared.NewErrorResponse(409, "this username already exist")
	}

	updated, err := us.database.Query.UpdateUsername(ctx, generated.UpdateUsernameParams{
		Username: newUsername,
		ID:       id,
	})

	if err != nil {
		return "", shared.NewErrorResponse(500, "something wrong while trying to update the username")
	}

	return updated.Username, nil
}

func (us *UsersService) SubmitOnBoarding(ctx context.Context,
	id uuid.UUID,
	username string) (generated.GetPublicProfileByUsernameRow, *shared.ErrorResponse) {

	oldData, err := us.database.Query.GetUserWithProfileByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return generated.GetPublicProfileByUsernameRow{}, shared.NewErrorResponse(404, "account not found")
		}

		return generated.GetPublicProfileByUsernameRow{},
			shared.NewErrorResponse(500,
				"something wrong with the server right now! please try again another time")
	}

	if oldData.Status != generated.UserStatusONBOARDING {
		return generated.GetPublicProfileByUsernameRow{},
			shared.NewErrorResponse(409, "you can't submit onboarding!")
	}

	exist, err := us.database.Query.IsUsernameTaken(ctx, username)
	if err != nil {
		return generated.GetPublicProfileByUsernameRow{},
			shared.NewErrorResponse(500,
				"something wrong with the server right now! please try again another time")
	}

	if exist {
		return generated.GetPublicProfileByUsernameRow{},
			shared.NewErrorResponse(409, "username already exist")
	}

	uptErr := us.database.Query.UpdateOnBoarding(ctx, generated.UpdateOnBoardingParams{
		Username: username,
		ID:       id,
	})

	if uptErr != nil {
		return generated.GetPublicProfileByUsernameRow{},
			shared.NewErrorResponse(500,
				"something went wrong in onboarding process please try again another time!")
	}

	cacheErr := us.userCache.Cache(ctx, shared.UserCredential{
		UserId: id,
		Status: string(generated.UserStatusACTIVE),
		Role:   string(oldData.Role),
	})

	if cacheErr != nil {
		fmt.Println("CACHE SET MISS REASON : ", cacheErr.Error())
	}

	respData := generated.GetPublicProfileByUsernameRow{
		ID:                id,
		Username:          username,
		IsVerified:        oldData.IsVerified,
		Role:              oldData.Role,
		DisplayName:       oldData.DisplayName,
		Bio:               oldData.Bio,
		AvatarUrl:         oldData.AvatarUrl,
		BannerUrl:         oldData.BannerUrl,
		SocialLinks:       oldData.SocialLinks,
		CosplayTags:       oldData.CosplayTags,
		AccountVisibility: oldData.AccountVisibility,
	}

	return respData, nil
}

func (us *UsersService) IsUsernameAvaible(ctx context.Context, username string) (bool, *shared.ErrorResponse) {

	ex, err := us.database.Query.IsUsernameTaken(ctx, username)
	if err != nil {
		return false, shared.NewErrorResponse(500, "something wrong with the server! please try again later")
	}

	return ex, nil
}

func AppendUploadsUrl(imgUrl *string) *string {

	if imgUrl == nil {
		return nil
	}

	tmp := publicUploadsUrl + *imgUrl

	return &tmp
}

func (us *UsersService) UpdateAccountVisibility(ctx context.Context, id uuid.UUID, vis string) (generated.User, *shared.ErrorResponse) {

	var visEnum generated.UsersVisibility = generated.UsersVisibilityPUBLIC
	if vis == "PRIVATE" {
		visEnum = generated.UsersVisibilityPRIVATE
	}

	data, err := us.database.Query.UpdateUsersAccVisibility(ctx, generated.UpdateUsersAccVisibilityParams{
		Visibility: visEnum,
		ID:         id,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return generated.User{}, shared.NewErrorResponse(404, "account not found!")
		}
		return generated.User{}, shared.NewErrorResponse(500, "something wrong while trying to update your profile!")
	}

	return data, nil
}
