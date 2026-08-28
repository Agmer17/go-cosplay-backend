package posts

import "mime/multipart"

type CreatePostsDto struct {
	Caption  *string                 `form:"caption" binding:"omitempty,max=255"`
	Location *string                 `form:"location" binding:"omitempty,max=255"`
	Media    []*multipart.FileHeader `form:"medias" binding:"required,min=1"`
}
