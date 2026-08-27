package posts

import "mime/multipart"

type CreatePostsDto struct {
	Caption  *string                 `form:"caption" bind:"omitempty,max=255"`
	Location *string                 `form:"location" bind:"omitempty,max=255"`
	Media    []*multipart.FileHeader `form:"medias" bind:"required,min=1"`
}
