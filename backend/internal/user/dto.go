package user

type UpdateUserRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
	Avatar    string `json:"avatar,omitempty"`
	Location  string `json:"location,omitempty"`
	Website   string `json:"website,omitempty"`
	GitHub    string `json:"github,omitempty"`
	Twitter   string `json:"twitter,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	Facebook  string `json:"facebook,omitempty"`
	YouTube   string `json:"youtube,omitempty"`
	Discord   string `json:"discord,omitempty"`
	Telegram  string `json:"telegram,omitempty"`
}

type UserProfileResponse struct {
	Username    string           `json:"username"`
	FirstName   string           `json:"first_name"`
	LastName    string           `json:"last_name"`
	Avatar      string           `json:"avatar,omitempty"`
	Bio         string           `json:"bio,omitempty"`
	Role        string           `json:"role"`
	Location    string           `json:"location,omitempty"`
	Website     string           `json:"website,omitempty"`
	JoinedAt    string           `json:"joined_at,omitempty"`
	Followers   int64            `json:"followers"`
	Following   int64            `json:"following"`
	SocialMedia SocialMediaLinks `json:"social_media,omitempty"`
	CanEdit     bool             `json:"can_edit"`
}

type SocialMediaLinks struct {
	GitHub    string `json:"github,omitempty"`
	Twitter   string `json:"twitter,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	Facebook  string `json:"facebook,omitempty"`
	YouTube   string `json:"youtube,omitempty"`
	Discord   string `json:"discord,omitempty"`
	Telegram  string `json:"telegram,omitempty"`
}

type UserPostsResponse struct {
	User  UserProfileResponse `json:"user"`
	Posts []PostSummary       `json:"posts"`
	Total int64               `json:"total"`
}

type PostSummary struct {
	Slug        string  `json:"slug"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Thumbnail   string  `json:"thumbnail,omitempty"`
	PublishedAt string  `json:"published_at,omitempty"`
	Views       int     `json:"views"`
	Likes       int     `json:"likes"`
	ReadTime    float64 `json:"read_time"`
	Tags        []Tag   `json:"tags,omitempty"`
}

type Tag struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
