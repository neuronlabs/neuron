package jsonapi

import (
	"time"
)

type ModelI18nSDK struct {
	ID   int    `jsonapi:"primary,i18n"`
	Lang string `jsonapi:"attr,language,langtag"`
}

type ModelSDK struct {
	ID   int    `jsonapi:"primary,models"`
	Name string `jsonapi:"attr,name"`
}

type AuthorSDK struct {
	ID    int        `jsonapi:"primary,authors"`
	Name  string     `jsonapi:"attr,name"`
	Blogs []*BlogSDK `jsonapi:"relation,blogs"`
}

type BlogSDK struct {
	ID          int      `jsonapi:"primary,blogs"`
	Lang        string   `jsonapi:"attr,language,langtag"`
	CurrentPost *PostSDK `jsonapi:"relation,current_post"`
}

type PostSDK struct {
	ID        int           `jsonapi:"primary,posts"`
	Title     string        `jsonapi:"attr,title"`
	BlogID    int           `jsonapi:"attr,blog_id,hidden"`
	CreatedAt time.Time     `jsonapi:"attr,created_at"`
	Comments  []*CommentSDK `jsonapi:"relation,comments"`
}

type CommentSDK struct {
	ID   int      `jsonapi:"primary,comments"`
	Body string   `jsonapi:"attr,body"`
	Post *PostSDK `jsonapi:"relation,post,hidden"`
}

type PetSDK struct {
	ID     int         `jsonapi:"primary,pets"`
	Name   string      `jsonapi:"attr,name"`
	Humans []*HumanSDK `jsonapi:"relation,humans"`
	Legs   int         `jsonapi:"attr,legs"`
}

type HumanSDK struct {
	ID   int       `jsonapi:"primary,humans"`
	Name string    `jsonapi:"attr,name"`
	Pets []*PetSDK `jsonapi:"relation,pets"`
}
