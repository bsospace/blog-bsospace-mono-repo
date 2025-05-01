package models

import (
	"time"
)

type UserRole string

const (
	NormalUser UserRole = "NORMAL_USER"
	WriterUser UserRole = "WRITER_USER"
	AdminUser  UserRole = "ADMIN_USER"
)

type User struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email     string `gorm:"uniqueIndex;not null"`
	Password  *string
	Image     *string
	Name      *string
	Bio       *string
	Role      UserRole  `gorm:"type:varchar(20);default:NORMAL_USER"`
	Posts     []Post    `gorm:"foreignKey:AuthorID"`
	Comments  []Comment `gorm:"foreignKey:AuthorID"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	DeletedAt *time.Time
}

type Post struct {
	ID          string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Title       string `gorm:"not null"`
	Description string
	Thumbnail   *string
	Example     *string
	Content     string `gorm:"type:text;not null"`
	Published   bool   `gorm:"default:false"`
	PublishedAt *time.Time
	Keywords    []string `gorm:"type:text[]"`
	Key         *string  // สำหรับ Embedding RAG
	AuthorID    string
	Author      User        `gorm:"foreignKey:AuthorID"`
	Likes       int         `gorm:"default:0"`
	Views       int         `gorm:"default:0"`
	ReadTime    float64     `gorm:"default:0"`
	Tags        []Tag       `gorm:"many2many:post_tags"`
	Categories  []Category  `gorm:"many2many:post_categories"`
	Comments    []Comment   `gorm:"foreignKey:PostID"`
	Embeddings  []Embedding `gorm:"foreignKey:PostID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type Comment struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Content   string `gorm:"type:text;not null"`
	PostID    string
	Post      Post `gorm:"foreignKey:PostID"`
	AuthorID  string
	Author    User `gorm:"foreignKey:AuthorID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Tag struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"uniqueIndex;not null"`
	Posts     []Post `gorm:"many2many:post_tags"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Category struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"uniqueIndex;not null"`
	Posts     []Post `gorm:"many2many:post_categories"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Embedding struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PostID    string    `gorm:"index;not null"`
	Post      Post      `gorm:"foreignKey:PostID"`
	Content   string    `gorm:"type:text;not null"` // raw text ชิ้นนี้ (chunk)
	Vector    []float32 `gorm:"type:float4[]"`      // embedding vector จาก LLM
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
