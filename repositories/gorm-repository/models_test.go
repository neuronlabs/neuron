package gormrepo

import (
	"time"
)

type Blog struct {
	ID            int    `jsonapi:"primary,blogs"`
	Title         string `jsonapi:"attr,title"`
	CurrentPost   *Post  `jsonapi:"relation,current_post" gorm:"foreignkey:CurrentPostID"`
	CurrentPostID int
	Author        *User `jsonapi:"relation,author" gorm:"foreignkey:AuthorID"`
	AuthorID      int
}

type User struct {
	ID        int       `jsonapi:"primary,users"`
	Name      string    `jsonapi:"attr,name"`
	CreatedAt time.Time `jsonapi:"attr,created_at"`
	Blogs     []*Blog   `jsonapi:"relation,blogs" gorm:"foreignkey:AuthorID"`
	Houses    []*House  `jsonapi:"relation,houses" gorm:"many2many:user_houses"`
}

type Post struct {
	ID        int        `gorm:"primary_index" jsonapi:"primary,posts"`
	Lang      string     `gorm:"primary_index" jsonapi:"attr,language,langtag"`
	Title     string     `jsonapi:"attr,title"`
	Comments  []*Comment `jsonapi:"relation,comments"`
	BlogID    int
	CreatedAt time.Time `jsonapi:"attr,created_at"`
	UpdatedAt time.Time `jsonapi:"attr,updated_at"`
	DeletedAt time.Time `jsonapi:"attr,deleted_at"`
}

type Comment struct {
	ID        uint64 `jsonapi:"primary,comments"`
	Body      string `jsonapi:"attr,body"`
	PostID    int
	CreatedAt time.Time `jsonapi:"attr,created_at"`
	UpdatedAt time.Time `jsonapi:"attr,updated_at"`
}

type House struct {
	ID    int     `jsonapi:"primary,houses"`
	Users []*User `jsonapi:"relation,users" gorm:"many2many:user_houses"`
}
