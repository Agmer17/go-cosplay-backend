package posts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Agmer17/go-cosplay-backend/internal/db"
	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leg100/surl/v2"
)

const posts_folder = "posts"
const mediaAccessUrlPrefix = "http://localhost/api/posts/media/"

type PostsService struct {
	db            *db.Database
	serverStorage *storage.FileStorage
	signer        *surl.Signer
}

func NewPostsService(d *db.Database, sr *storage.FileStorage, sn *surl.Signer) *PostsService {

	return &PostsService{
		db:            d,
		serverStorage: sr,
		signer:        sn,
	}
}

func (ps *PostsService) CreatePosts(ctx context.Context, userId uuid.UUID, dto CreatePostsDto) (generated.GetPostByIDRow, *shared.ErrorResponse) {
	userData, err := ps.db.Query.GetUserByID(ctx, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return generated.GetPostByIDRow{}, shared.NewErrorResponse(404, "your account was not found!")
		}
		return generated.GetPostByIDRow{}, shared.NewErrorResponse(500, "something wrong with the server! try again another time!")
	}

	postVis := generated.PostVisibilityPUBLIC
	if userData.AccountVisibility == generated.UsersVisibilityPRIVATE {
		postVis = generated.PostVisibilityPRIVATE
	}

	postId := uuid.Must(uuid.NewV7())

	medFilenames, rawMediaTypes, err := ps.serverStorage.SaveAllPrivateFile(ctx, dto.Media, posts_folder)
	if err != nil {
		return generated.GetPostByIDRow{}, shared.NewErrorResponse(500, "something wrong while trying to save the media! please try again another time")
	}

	allMediaTypes := make([]string, len(rawMediaTypes))
	for i, v := range rawMediaTypes {
		switch v {
		case storage.TypeImage:
			allMediaTypes[i] = strings.ToUpper(storage.TypeImage)
		case storage.TypeVideo:
			allMediaTypes[i] = strings.ToUpper(storage.TypeVideo)
		default:
			ps.serverStorage.DeleteAllPrivateFiles(medFilenames, posts_folder)
			return generated.GetPostByIDRow{}, shared.NewErrorResponse(400, "unsupported media type detected")
		}
	}

	medUrls := make([]string, len(medFilenames))
	mediaOrders := make([]int16, len(medFilenames))
	for i := range medUrls {
		medUrls[i] = posts_folder + "/" + medFilenames[i]
		mediaOrders[i] = int16(i)
	}
	var createdPost generated.Post
	var createdMedia []generated.PostsMedium

	psErr := ps.db.Transaction(ctx, func(qtx *generated.Queries) error {
		var txErr error

		createdPost, txErr = qtx.CreatePost(ctx, generated.CreatePostParams{
			ID:         postId,
			UserID:     userId,
			Caption:    dto.Caption,
			Location:   dto.Location,
			Visibility: postVis,
		})
		if txErr != nil {
			return fmt.Errorf("CreatePost error: %w", txErr)
		}

		createdMedia, txErr = qtx.CreatePostMediaBatch(ctx, generated.CreatePostMediaBatchParams{
			PostID:        postId,
			MediaTypes:    allMediaTypes,
			MediaUrls:     medUrls,
			DisplayOrders: mediaOrders,
		})
		if txErr != nil {
			return fmt.Errorf("CreatePostMediaBatch error: %w", txErr)
		}

		return nil
	})

	if psErr != nil {
		ps.serverStorage.DeleteAllPrivateFiles(medFilenames, posts_folder)
		fmt.Printf("DB Transaction failed on CreatePosts: %v\n", psErr)
		return generated.GetPostByIDRow{}, shared.NewErrorResponse(500, "cannot save the posts data")
	}

	ps.GenerateMediaUrl(createdMedia)
	mediaBytes, err := json.Marshal(createdMedia)
	if err != nil {
		mediaBytes = []byte("[]")
	}

	return generated.GetPostByIDRow{
		ID:            createdPost.ID,
		UserID:        createdPost.UserID,
		Caption:       createdPost.Caption,
		Location:      createdPost.Location,
		Visibility:    createdPost.Visibility,
		LikeCount:     0,
		CommentCount:  0,
		BookmarkCount: 0,
		ShareCount:    0,
		CreatedAt:     createdPost.CreatedAt,
		UpdatedAt:     createdPost.UpdatedAt,
		Media:         mediaBytes,
	}, nil
}

func (ps *PostsService) signMediaUrl(filename string, ttl time.Duration) string {
	unsignedUrl := mediaAccessUrlPrefix + filename
	signed, _ := ps.signer.Sign(unsignedUrl, time.Now().Add(ttl))

	return signed
}

func (ps *PostsService) GenerateMediaUrl(el []generated.PostsMedium) {
	for i := range el {
		filename := strings.Split(el[i].MediaUrl, "/")[1]

		var ttl time.Duration = 15 * time.Minute

		if el[i].MediaType == generated.PostMediaTypeVIDEO {
			ttl = 1 * time.Hour
		}

		el[i].MediaUrl = ps.signMediaUrl(filename, ttl)
	}
}

func (ps *PostsService) GetPostsWithMediaById(ctx context.Context,
	pID uuid.UUID, usrID uuid.UUID) (generated.GetPostByIDRow, *shared.ErrorResponse) {

	post, err := ps.db.Query.GetPostByID(ctx, pID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return generated.GetPostByIDRow{}, shared.NewErrorResponse(404, "posts not found")
		}

		return generated.GetPostByIDRow{}, shared.NewErrorResponse(500, "Internal server error! please try again another time")

	}

	if post.Visibility == generated.PostVisibilityPRIVATE {

		if usrID == uuid.Nil {
			return generated.GetPostByIDRow{}, shared.NewErrorResponse(403, "you don't have permission to see this posts")
		}

		cv, err := ps.isPostVisibleToUser(usrID, post.UserID)
		if err != nil {
			return generated.GetPostByIDRow{}, shared.NewErrorResponse(500, "Internal server error! please try again another time")
		}

		if !cv {
			return generated.GetPostByIDRow{}, shared.NewErrorResponse(403, "you don't have permission to see this posts")
		}
	}

	var postMedia []generated.PostsMedium
	mErr := json.Unmarshal(post.Media, &postMedia)

	if mErr != nil {
		return generated.GetPostByIDRow{}, shared.NewErrorResponse(500, "Internal server error! please try again another time")
	}

	ps.GenerateMediaUrl(postMedia)
	mediaBytes, err := json.Marshal(postMedia)
	if err != nil {
		return generated.GetPostByIDRow{}, shared.NewErrorResponse(500, "Internal server error! please try again another time")
	}
	post.Media = mediaBytes

	return post, nil
}

func (ps *PostsService) isPostVisibleToUser(userId, authorId uuid.UUID) (bool, *shared.ErrorResponse) {
	// nanti impl ini jika sistem follow selesai!
	return true, nil
}

func (ps *PostsService) DeletePosts(ctx context.Context, postID, userID uuid.UUID) *shared.ErrorResponse {

	postsData, err := ps.db.Query.GetPostsDataOnlyById(ctx, postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return shared.NewErrorResponse(404, "posts not found")
		}
		return shared.NewErrorResponse(500, "something wrong while retrieving posts data")
	}

	if postsData.UserID != userID {
		return shared.NewErrorResponse(403, "access forbidden you can't delete this posts")
	}

	err = ps.db.Query.DeletePost(ctx, postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return shared.NewErrorResponse(404, "posts not found")
		}
		return shared.NewErrorResponse(500, "something wrong while trying to delete posts data")
	}

	return nil
}

func (ps *PostsService) ResolvePostsMediaPath(filename string) string {

	path := filepath.Join(
		ps.serverStorage.PrivatePath,
		posts_folder,
		filename,
	)

	return path

}
