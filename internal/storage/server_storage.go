package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Agmer17/go-cosplay-backend/internal/utils"
	"github.com/disintegration/imaging"
	"golang.org/x/sync/errgroup"
)

var errorsFileNotSupported = errors.New("This file type currently not supported")
var allowedMimes = map[string]bool{
	// image
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,

	// video
	"video/mp4":       true,
	"video/webm":      true,
	"video/quicktime": true,

	// audio
	"audio/mpeg": true,
	"audio/wav":  true,
	"audio/ogg":  true,

	// document
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true, // .xlsx

	"application/vnd.ms-powerpoint":                                             true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true, // .pptx

	"text/plain": true, // .txt
}

const (
	TypeAudio    = "audio"
	TypeImage    = "image"
	TypeVideo    = "video"
	TypeDocument = "document"
)

type FileStorage struct {
	Root        string
	PublicPath  string
	PrivatePath string
	maxWorkers  int
}

func mustGetProjectRoot() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(exe)
}

func mustCreateDir(path string) {
	if err := os.MkdirAll(path, 0755); err != nil {
		panic(err)
	}
}

func NewFileStorage(worker int) *FileStorage {
	root := mustGetProjectRoot()
	uploadsDir := filepath.Join(root, "uploads")
	privateDir := filepath.Join(uploadsDir, "private")
	publicDir := filepath.Join(uploadsDir, "public")
	mustCreateDir(uploadsDir)
	mustCreateDir(privateDir)
	mustCreateDir(publicDir)

	var maxWork int = 0

	if worker <= 0 {
		maxWork = 5
	} else {
		maxWork = worker
	}

	fileStoreage := FileStorage{
		Root:        root,
		PublicPath:  publicDir,
		PrivatePath: privateDir,
		maxWorkers:  maxWork,
	}

	return &fileStoreage
}

func (storage *FileStorage) validateFile(f multipart.File) (string, string, error) {
	buf := make([]byte, 512)
	_, err := f.Read(buf)
	if err != nil {
		return "", "", err
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return "", "", err
	}

	mimeType := http.DetectContentType(buf)

	if !allowedMimes[mimeType] {
		return "", "", errorsFileNotSupported
	}

	_, after, _ := strings.Cut(mimeType, "/")

	fileExt := "." + after

	mediaType := storage.GetMediaType(mimeType)

	return fileExt, mediaType, nil
}

func (storage *FileStorage) GetMediaType(mime string) string {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return TypeImage
	case strings.HasPrefix(mime, "video/"):
		return TypeVideo
	case strings.HasPrefix(mime, "audio/"):
		return TypeAudio
	case strings.HasPrefix(mime, "application/"):
		return TypeDocument
	default:
		return ""
	}
}

func (storage *FileStorage) compressImage(src multipart.File, dst io.Writer) error {
	img, err := imaging.Decode(src)
	if err != nil {
		return err
	}

	// resize
	img = imaging.Resize(img, 1024, 0, imaging.Lanczos)

	// encode ke JPEG (compressed)
	err = imaging.Encode(dst, img, imaging.JPEG, imaging.JPEGQuality(70))
	if err != nil {
		return err
	}

	return nil
}

func (storage *FileStorage) SavePublicFile(fileHeader *multipart.FileHeader, place ...string) (string, string, error) {
	return storage.saveFile(storage.PublicPath, fileHeader, place...)
}

func (storage *FileStorage) SavePrivateFile(fileHeader *multipart.FileHeader, place ...string) (string, string, error) {
	return storage.saveFile(storage.PrivatePath, fileHeader, place...)
}

func (storage *FileStorage) saveFile(basePath string, fileHeader *multipart.FileHeader, place ...string) (string, string, error) {
	f, err := fileHeader.Open()
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	ext, mediaType, err := storage.validateFile(f)
	if err != nil {
		return "", "", err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", "", err
	}

	filename, err := utils.GenerateSecureString(40)
	if err != nil {
		return "", "", err
	}
	filename += ext

	parts := append([]string{basePath}, place...)
	parts = append(parts, filename)

	fullPath := filepath.Join(parts...)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", "", err
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", "", err
	}
	defer dst.Close()

	if mediaType == TypeImage {
		if err := storage.compressImage(f, dst); err != nil {
			return "", "", err
		}
	} else {
		if _, err := io.Copy(dst, f); err != nil {
			return "", "", err
		}
	}

	return filename, mediaType, nil
}

func (storage *FileStorage) SaveAllPublicFile(ctx context.Context, ls []*multipart.FileHeader, place ...string) ([]string, []string, error) {
	return storage.saveAllFiles(ctx, storage.PublicPath, ls, place...)
}

func (storage *FileStorage) SaveAllPrivateFile(ctx context.Context, ls []*multipart.FileHeader, place ...string) ([]string, []string, error) {
	return storage.saveAllFiles(ctx, storage.PrivatePath, ls, place...)
}

func (storage *FileStorage) saveAllFiles(ctx context.Context, basePath string, ls []*multipart.FileHeader, place ...string) ([]string, []string, error) {
	if len(ls) == 0 {
		return []string{}, []string{}, nil
	}

	g, ctx := errgroup.WithContext(ctx)
	semaphore := make(chan struct{}, storage.maxWorkers)
	filenames := make([]string, len(ls))
	mediaTypes := make([]string, len(ls))

	for i, fh := range ls {
		g.Go(func() error {
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return ctx.Err()
			}

			filename, mt, err := storage.saveFile(basePath, fh, place...)
			if err != nil {
				return err
			}

			filenames[i] = filename
			mediaTypes[i] = mt
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		for _, filename := range filenames {
			if filename == "" {
				continue
			}

			parts := make([]string, 0, 1+len(place)+1)
			parts = append(parts, basePath)
			parts = append(parts, place...)
			parts = append(parts, filename)
			if err := os.Remove(filepath.Join(parts...)); err != nil {
				fmt.Println("error while cleaning up the files : ", err.Error())
			}
		}
		return nil, nil, err
	}

	return filenames, mediaTypes, nil
}

func (storage *FileStorage) deleteFiles(basePath string, files string, place ...string) error {
	parts := append([]string{basePath}, place...)
	parts = append(parts, files)

	path := filepath.Join(parts...)

	return os.Remove(path)
}

func (storage *FileStorage) DeleteAllPrivateFiles(files []string, place ...string) {
	for _, v := range files {
		storage.deleteFiles(storage.PrivatePath, v, place...)
	}
}

func (storage *FileStorage) DeletePublicFile(filename string, place ...string) error {
	return storage.deleteFiles(storage.PublicPath, filename, place...)
}
