package likes

import (
	"context"
	"errors"

	"github.com/Agmer17/go-cosplay-backend/internal/db"
	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostLikesService struct {
	database *db.Database
}

func NewPostLikesService(d *db.Database) *PostLikesService {
	return &PostLikesService{
		database: d,
	}
}

func (pl *PostLikesService) CreateLikes(ctx context.Context, postsID, userID uuid.UUID) (generated.PostLike, *shared.ErrorResponse) {

	data, err := pl.database.Query.CreatePostLike(ctx, generated.CreatePostLikeParams{
		ID:     uuid.Must(uuid.NewV7()),
		UserID: userID,
		PostID: postsID,
	})

	if err != nil {

		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case "23503": // Foreign Key Violation
				return generated.PostLike{}, shared.NewErrorResponse(404, "can't create likes, maybe the post or user was deleted")

			case "23505": // Unique Constraint Violation
				return generated.PostLike{}, shared.NewErrorResponse(409, "you already like this post")
			}
		}

		return generated.PostLike{}, shared.NewErrorResponse(500, "something wrong with the server : "+err.Error())
	}

	return data, nil
}

func (pl *PostLikesService) GetLikesDetailFromPosts(
	ctx context.Context,
	postID uuid.UUID) ([]generated.GetPostLikesWithDetailsRow, *shared.ErrorResponse) {
	data, err := pl.database.Query.GetPostLikesWithDetails(ctx, postID)
	if err != nil {
		return []generated.GetPostLikesWithDetailsRow{}, shared.NewErrorResponse(500, "something wrong while trying to get the likes data!")
	}
	return data, nil
}

func (pl *PostLikesService) Delete(ctx context.Context, postsID, userID uuid.UUID) *shared.ErrorResponse {

	row, err := pl.database.Query.DeletePostLike(ctx, generated.DeletePostLikeParams{
		PostID: postsID,
		UserID: userID,
	})

	if err != nil {
		return shared.NewErrorResponse(500, "something wrong while trying to delete the likes")
	}

	if row == 0 {
		return shared.NewErrorResponse(404, "you haven't likes this posts")

	}
	return nil
}

func (pl *PostLikesService) GetLikesWithPostsFromUsers(
	ctx context.Context, userID uuid.UUID) ([]generated.GetPostsByIDArrayRow, *shared.ErrorResponse) {

	likeData, err := pl.database.Query.GetUserLikes(ctx, userID)
	if err != nil {
		return []generated.GetPostsByIDArrayRow{}, shared.NewErrorResponse(500, "something wrong while trying toget likes data")

	}

	if len(likeData) == 0 {
		return []generated.GetPostsByIDArrayRow{}, nil
	}

	var postIDS []uuid.UUID = make([]uuid.UUID, len(likeData))

	for i, v := range likeData {
		postIDS[i] = v.PostID
	}

	postsLiked, err := pl.database.Query.GetPostsByIDArray(ctx, postIDS)
	if err != nil {
		return []generated.GetPostsByIDArrayRow{}, shared.NewErrorResponse(500, "something wrong while trying to get likes data")
	}

	for i := range postsLiked {
		postsLiked[i].IsLiked = true
	}

	return postsLiked, nil
}
