package jsonapi

import (
	"time"
)

type ModelI18nSDK struct {
	ID   int    `jsonapi:"type=primary"`
	Lang string `jsonapi:"type=attr;name=language;flags=langtag"`
}

func (m *ModelI18nSDK) CollectionName() string {
	return "i18n"
}

type ModelSDK struct {
	ID   int    `jsonapi:"type=primary"`
	Name string `jsonapi:"type=attr"`
}

func (c ModelSDK) CollectionName() string {
	return "models"
}

type AuthorSDK struct {
	ID    int        `jsonapi:"type=primary"`
	Name  string     `jsonapi:"type=attr"`
	Blogs []*BlogSDK `jsonapi:"type=relation;foreign=AuthorID"`
}

func (c AuthorSDK) CollectionName() string {
	return "authors"
}

type BlogSDK struct {
	ID                int      `jsonapi:"type=primary"`
	Lang              string   `jsonapi:"type=attr;flags=langtag"`
	SomeAttr          string   `jsonapi:"type=attr"`
	AuthorID          int      `jsonapi:"type=foreign"`
	CurrentPost       *PostSDK `jsonapi:"type=relation;foreign=BlogID"`
	CurrentPostNoSync *PostSDK `jsonapi:"type=relation;relation=nosync"`
}

func (c BlogSDK) CollectionName() string {
	return "blogs"
}

type PostSDK struct {
	ID             int           `jsonapi:"type=primary"`
	Title          string        `jsonapi:"type=attr"`
	BlogID         int           `jsonapi:"type=foreign"`
	CreatedAt      time.Time     `jsonapi:"type=attr"`
	Comments       []*CommentSDK `jsonapi:"type=relation;foreign=PostID"`
	CommentsNoSync []*CommentSDK `jsonapi:"type=relation;relation=nosync"`
}

func (c PostSDK) CollectionName() string {
	return "posts"
}

type CommentSDK struct {
	ID     int      `jsonapi:"type=primary"`
	Body   string   `jsonapi:"type=attr"`
	Post   *PostSDK `jsonapi:"type=relation;foreign=PostID"`
	PostID int      `jsonapi:"type=foreign"`
}

func (c CommentSDK) CollectionName() string {
	return "comments"
}

type PetSDK struct {
	ID         int         `jsonapi:"type=primary"`
	Name       string      `jsonapi:"type=attr"`
	Humans     []*HumanSDK `jsonapi:"type=relation;relation=many2many,common"`
	HumansSync []*HumanSDK `jsonapi:"type=relation;relation=many2many,sync,PetsSync"`
	Legs       int         `jsonapi:"type=attr"`
}

func (c PetSDK) CollectionName() string {
	return "pets"
}

type HumanSDK struct {
	ID       int       `jsonapi:"type=primary"`
	Name     string    `jsonapi:"type=attr"`
	Pets     []*PetSDK `jsonapi:"type=relation;relation=many2many,common"`
	PetsSync []*PetSDK `jsonapi:"type=relation;relation=many2many,sync,HumansSync"`
}

func (c HumanSDK) CollectionName() string {
	return "humans"
}

type ModCliGenID struct {
	ID string `jsonapi:"type=primary;flags=client-id"`
}

func (m *ModCliGenID) CollectionName() string {
	return "client-generated"
}
