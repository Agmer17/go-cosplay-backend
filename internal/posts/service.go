package posts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Agmer17/go-cosplay-backend/internal/db"
	"github.com/Agmer17/go-cosplay-backend/internal/db/generated"
	"github.com/Agmer17/go-cosplay-backend/internal/shared"
	"github.com/Agmer17/go-cosplay-backend/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const posts_folder = "posts"

type PostsService struct {
	db            *db.Database
	serverStorage *storage.FileStorage
}

func NewPostsService(d *db.Database, sr *storage.FileStorage) *PostsService {

	return &PostsService{
		db:            d,
		serverStorage: sr,
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

	if len(dto.Media) < 1 {
		return generated.GetPostByIDRow{}, shared.NewErrorResponse(400, "please provide at least one media for posts!")
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
		case "image":
			allMediaTypes[i] = strings.ToUpper("image")
		case "video":
			allMediaTypes[i] = strings.ToUpper("video")
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

	// 1. Siapkan penampung untuk menangkap data hasil eksekusi dari database
	var createdPost generated.Post           // Asumsi nama struct dari sqlc adalah 'Post'
	var createdMedia []generated.PostsMedium // Asumsi struct array dari tabel media

	psErr := ps.db.Transaction(ctx, func(qtx *generated.Queries) error {
		var txErr error

		// 2. Jangan pakai '_', tangkap data hasil buatannya
		createdPost, txErr = qtx.CreatePost(ctx, generated.CreatePostParams{
			ID:         postId,
			UserID:     userId,
			Caption:    dto.Caption,
			Location:   dto.Location,
			Visibility: postVis,
		})
		if txErr != nil {
			return fmt.Errorf("CreatePost error: %w", txErr) // Wrap error biar jelas
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
		// LOG INI SANGAT PENTING BIAR KAMU TAU PENYEBAB 500-NYA
		fmt.Printf("DB Transaction failed on CreatePosts: %v\n", psErr)
		return generated.GetPostByIDRow{}, shared.NewErrorResponse(500, "cannot save the posts data")
	}

	// 3. Skip query ulang, langsung rakit JSON media dan bangun response
	mediaBytes, err := json.Marshal(createdMedia)
	if err != nil {
		mediaBytes = []byte("[]") // fallback
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
