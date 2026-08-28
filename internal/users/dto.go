package users

import (
	"encoding/json"
)

type UpdateProfilesDto struct {
	DisplayName *string          `json:"display_name" binding:"omitempty,min=1,max=75"`
	Bio         *string          `json:"bio" binding:"omitempty,max=500"`
	SocialLinks *json.RawMessage `json:"social_links"`
	CosplayTags []string         `json:"cosplay_tags" binding:"omitempty,max=10,dive,min=1,max=30"`
}

type UsernamePostsDto struct {
	Username string `json:"username" binding:"required,min=4,max=30,alphanum"`
}

type UpdateUsersPrvacy struct {
	Visibility string `json:"visibility" binding:"required,oneof=PUBLIC PRIVATE"`
}
