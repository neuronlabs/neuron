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
	Blogs []*BlogSDK `jsonapi:"type=relation"`
}

func (c AuthorSDK) CollectionName() string {
	return "authors"
}

type BlogSDK struct {
	ID          int      `jsonapi:"type=primary"`
	Lang        string   `jsonapi:"type=attr;flags=langtag"`
	CurrentPost *PostSDK `jsonapi:"type=relation"`
}

func (c BlogSDK) CollectionName() string {
	return "blogs"
}

type PostSDK struct {
	ID        int           `jsonapi:"type=primary"`
	Title     string        `jsonapi:"type=attr"`
	BlogID    int           `jsonapi:"type=attr;flags=hidden"`
	CreatedAt time.Time     `jsonapi:"type=attr"`
	Comments  []*CommentSDK `jsonapi:"type=relation"`
}

func (c PostSDK) CollectionName() string {
	return "posts"
}

type CommentSDK struct {
	ID   int      `jsonapi:"type=primary"`
	Body string   `jsonapi:"type=attr"`
	Post *PostSDK `jsonapi:"type=relation;flags=hidden"`
}

func (c CommentSDK) CollectionName() string {
	return "comments"
}

type PetSDK struct {
	ID     int         `jsonapi:"type=primary"`
	Name   string      `jsonapi:"type=attr"`
	Humans []*HumanSDK `jsonapi:"type=relation"`
	Legs   int         `jsonapi:"type=attr"`
}

func (c PetSDK) CollectionName() string {
	return "pets"
}

type HumanSDK struct {
	ID   int       `jsonapi:"type=primary"`
	Name string    `jsonapi:"type=attr"`
	Pets []*PetSDK `jsonapi:"type=relation"`
}

func (c HumanSDK) CollectionName() string {
	return "humans"
}
