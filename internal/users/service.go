package users

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/Agmer17/go-cosplay-backend/internal/db"
	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const avatarFolder = "avatar"
const bannerFolder = "banner"

type UsersService struct {
	database      *db.Database
	serverStorage *storage.FileStorage
}

func NewUsersService(db *db.Database, storage *storage.FileStorage) *UsersService {

	return &UsersService{
		database:      db,
		serverStorage: storage,
	}
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
	return data, nil
}

func (us *UsersService) UpdateUsersProfilesById(ctx context.Context, id uuid.UUID, dto UpdateProfilesDto) (generated.Profile, *shared.ErrorResponse) {

	newData, err := us.database.Query.UpdateProfile(ctx, generated.UpdateProfileParams{
		DisplayName: dto.DisplayName,
		Bio:         dto.Bio,
		SocialLinks: *dto.SocialLinks,
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

	return newData, nil
}

func (us *UsersService) UpdateProfilePicture(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (string, *shared.ErrorResponse) {

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

	return avatarUrl, nil
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
	return bannerUrl, nil

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

func (us *UsersService) SubmitOnBoarding(ctx context.Context, id uuid.UUID, username string) *shared.ErrorResponse {
	return nil
}
