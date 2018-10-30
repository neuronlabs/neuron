package gormrepo

import (
	"time"
)

type Blog struct {
	ID          int    `jsonapi:"type=primary"`
	Title       string `jsonapi:"type=attr"`
	CurrentPost *Post  `jsonapi:"type=relation;foreign=BlogID" gorm:"foreignkey:BlogID"`
	Author      *User  `jsonapi:"type=relation" gorm:"foreignkey:AuthorID"`
	AuthorID    int    `jsonapi:"type=foreign"`
}

type User struct {
	ID        int       `jsonapi:"type=primary"`
	Name      string    `jsonapi:"type=attr"`
	CreatedAt time.Time `jsonapi:"type=attr"`
	Blogs     []*Blog   `jsonapi:"type=relation;foreign=AuthorID" gorm:"foreignkey:AuthorID"`
	Houses    []*House  `jsonapi:"type=relation;relation=many2many,Users" gorm:"many2many:user_houses"`
}

type Post struct {
	ID int `gorm:"primary_index" jsonapi:"type=primary"`

	BlogID    int        `jsonapi:"type=foreign"`
	Lang      string     `gorm:"primary_index" jsonapi:"type=attr;flags=langtag"`
	Title     string     `jsonapi:"type=attr"`
	Comments  []*Comment `jsonapi:"type=relation"`
	CreatedAt time.Time  `jsonapi:"type=attr"`
	UpdatedAt time.Time  `jsonapi:"type=attr"`
	DeletedAt time.Time  `jsonapi:"type=attr"`
}

type Comment struct {
	ID        uint64    `jsonapi:"type=primary"`
	Body      string    `jsonapi:"type=attr"`
	Post      *Post     `jsonapi:"type=relation"`
	PostID    int       `jsonapi:"type=foreign"`
	CreatedAt time.Time `jsonapi:"type=attr"`
	UpdatedAt time.Time `jsonapi:"type=attr"`
}

type House struct {
	ID    int     `jsonapi:"type=primary"`
	Users []*User `jsonapi:"type=relation;relation=many2many,Houses" gorm:"many2many:user_houses"`
}

type Simple struct {
	ID    int    `jsonapi:"type=primary"`
	Attr1 string `jsonapi:"type=attr"`
	Attr2 int    `jsonapi:"type=attr"`
}

type Human struct {
	ID            int         `jsonapi:"type=primary"`
	Nose          *BodyPart   `jsonapi:"type=relation;foreign=HumanID" gorm:"foreignkey:HumanID"`
	NoseNonSynced *BodyPart   `jsonapi:"type=relation;foreign=HumanNonSyncID;relation=nosync" gorm:"foreignkey:HumanNonSyncID"`
	Ears          []*BodyPart `jsonapi:"type=relation;foreign=HumanID" gorm:"foreignkey:HumanID"`
	EarsNonSync   []*BodyPart `jsonapi:"type=relation;foreign=HumanNonSyncID;relation=nosync" gorm:"foreignkey:HumanNonSyncID"`
}

type BodyPart struct {
	ID             int `jsonapi:"type=primary"`
	HumanID        int `jsonapi:"type=foreign"`
	HumanNonSyncID int `jsonapi:"type=foreign"`
}

type M2MFirst struct {
	ID          int          `jsonapi:"type=primary"`
	Seconds     []*M2MSecond `jsonapi:"type=relation;relation=many2many" gorm:"many2many:first_seconds"`
	SecondsSync []*M2MSecond `jsonapi:"type=relation;relation=many2many,sync,Firsts"`
}

type M2MSecond struct {
	ID     int         `jsonapi:"type=primary"`
	Firsts []*M2MFirst `jsonapi:"type=relation;relation=many2many" gorm:"many2many:first_seconds"`
}
