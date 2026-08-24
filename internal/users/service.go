package users

import (
	"context"
	"errors"

	"github.com/Agmer17/go-cosplay-backend/internal/db"
	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

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

func (us *UsersService) UpdateProfilePicture(ctx context.Context) {

}
