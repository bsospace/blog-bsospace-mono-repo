package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	NormalUser UserRole = "NORMAL_USER"
	WriterUser UserRole = "WRITER_USER"
	AdminUser  UserRole = "ADMIN_USER"
)

type BaseModel struct {
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email    string    `gorm:"uniqueIndex;not null" json:"email"`
	Avatar   string    `json:"avatar,omitempty"`
	Password string    `gorm:"not null" json:"-"`
	UserName string    `json:"username,omitempty"`
	Image    string    `json:"image,omitempty"`
	Bio      string    `json:"bio,omitempty"`
	Role     UserRole  `gorm:"type:varchar(20);default:NORMAL_USER" json:"role"`
	BaseModel

	Posts         []Post         `gorm:"foreignKey:AuthorID;references:ID" json:"posts,omitempty"`
	Comments      []Comment      `gorm:"foreignKey:AuthorID;references:ID" json:"comments,omitempty"`
	AIUsageLogs   []AIUsageLog   `gorm:"foreignKey:UserID;references:ID" json:"ai_usage_logs,omitempty"`
	Notifications []Notification `gorm:"foreignKey:UserID;references:ID" json:"notifications,omitempty"`
}

type Post struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Slug        string     `gorm:"uniqueIndex;not null" json:"slug"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `json:"description,omitempty"`
	Thumbnail   string     `json:"thumbnail,omitempty"`
	Example     string     `json:"example,omitempty"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	Published   bool       `gorm:"default:false" json:"published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Keywords    []string   `gorm:"type:text[]" json:"keywords,omitempty"`
	Key         string     `json:"key,omitempty"`
	Likes       int        `gorm:"default:0" json:"likes"`
	Views       int        `gorm:"default:0" json:"views"`
	ReadTime    float64    `gorm:"default:0" json:"read_time"`
	BaseModel

	AuthorID   uuid.UUID   `gorm:"not null" json:"author_id"`
	Author     User        `gorm:"foreignKey:AuthorID;references:ID" json:"author,omitempty"`
	Tags       []Tag       `gorm:"many2many:post_tags" json:"tags,omitempty"`
	Categories []Category  `gorm:"many2many:post_categories" json:"categories,omitempty"`
	Comments   []Comment   `gorm:"foreignKey:PostID;references:ID" json:"comments,omitempty"`
	Embeddings []Embedding `gorm:"foreignKey:PostID;references:ID" json:"embeddings,omitempty"`
}

type Comment struct {
	ID      uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Content string `gorm:"type:text;not null" json:"content"`
	BaseModel

	PostID   uuid.UUID `gorm:"not null" json:"post_id"`
	AuthorID uuid.UUID `gorm:"not null" json:"author_id"`
	Post     Post      `gorm:"foreignKey:PostID;references:ID" json:"post,omitempty"`
	Author   User      `gorm:"foreignKey:AuthorID;references:ID" json:"author,omitempty"`
}

type Tag struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"uniqueIndex;not null" json:"name"`
	BaseModel

	Posts []Post `gorm:"many2many:post_tags" json:"posts,omitempty"`
}

type Category struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"uniqueIndex;not null" json:"name"`
	BaseModel

	Posts []Post `gorm:"many2many:post_categories" json:"posts,omitempty"`
}

type Embedding struct {
	ID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	PostID  uuid.UUID `gorm:"not null;index" json:"post_id"`
	Content string    `gorm:"type:text;not null" json:"content"`
	Vector  []float32 `gorm:"type:float4[]" json:"vector"`
	BaseModel

	Post Post `gorm:"foreignKey:PostID;references:ID" json:"post,omitempty"`
}

type Notification struct {
	ID      uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title   string    `gorm:"not null" json:"title"`
	Content string    `gorm:"type:text;not null" json:"content"`
	Link    string    `gorm:"not null" json:"link"`
	Seen    bool      `gorm:"default:false" json:"seen"`
	SeenAt  time.Time `gorm:"autoUpdateTime" json:"seen_at,omitempty"`
	BaseModel

	UserID uuid.UUID `gorm:"not null" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

type AIUsageLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"index;not null" json:"user_id"`
	UsedAt    time.Time `gorm:"autoCreateTime" json:"used_at"`
	Action    string    `gorm:"type:varchar(50)" json:"action"`
	TokenUsed int       `json:"token_used"`
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	BaseModel

	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

type AIResponse struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"not null" json:"user_id"`
	PostID      uuid.UUID `gorm:"not null" json:"post_id"`
	EmbeddingID uuid.UUID `gorm:"not null" json:"embedding_id"`
	UsedAt      time.Time `gorm:"autoCreateTime" json:"used_at"`
	Prompt      string    `gorm:"type:text" json:"prompt"`
	Response    string    `gorm:"type:text" json:"response"`
	TokenUsed   int       `json:"token_used"`
	Success     bool      `json:"success"`
	Message     string    `json:"message,omitempty"`
	BaseModel

	User      User      `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Post      Post      `gorm:"foreignKey:PostID;references:ID" json:"post,omitempty"`
	Embedding Embedding `gorm:"foreignKey:EmbeddingID;references:ID" json:"embedding,omitempty"`
}
