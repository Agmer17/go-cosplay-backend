package users

import "encoding/json"

type UpdateProfilesDto struct {
	DisplayName *string          `json:"display_name"`
	Bio         *string          `json:"bio"`
	SocialLinks *json.RawMessage `json:"social_links"`
	CosplayTags []string         `json:"cosplay_tags"`
}
