package di

// IVideoRepository is the port the user service needs from the video domain.
// It is defined in di (which imports neither user nor video) so that the user
// package can depend on it without pulling in the video package — video already
// imports user, so any user -> video edge would form an import cycle.
type IVideoRepository interface {
	FindSubscribedVideos(userID string) ([]SubscribedVideo, error)
}

// SubscribedVideo is the transport DTO returned by IVideoRepository. It mirrors
// the Prisma include of the channel (with its owning user) and the likes, but
// belongs to di so neither user nor di has to import the video package.
type SubscribedVideo struct {
	ID           string `json:"id"`
	PublicId     string `json:"publicId"`
	Title        string `json:"title"`
	ThumbnailUrl string `json:"thumbnailUrl"`
	ViewsCount   int    `json:"viewsCount"`
	ChannelID    string `json:"channelId"`

	Channel *SubscribedChannel `json:"channel,omitempty"`
	Likes   []SubscribedLike   `json:"likes"`
}

// SubscribedChannel is the eager-loaded channel with its owning user.
type SubscribedChannel struct {
	ID         string  `json:"id"`
	Slug       string  `json:"slug"`
	AvatarUrl  *string `json:"avatarUrl"`
	IsVerified bool    `json:"isVerified"`

	User *SubscribedChannelUser `json:"user,omitempty"`
}

// SubscribedChannelUser is the channel owner, projected to public fields only.
type SubscribedChannelUser struct {
	ID    string  `json:"id"`
	Name  *string `json:"name"`
	Email string  `json:"email"`
}

// SubscribedLike is an eager-loaded like on a video.
type SubscribedLike struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}
