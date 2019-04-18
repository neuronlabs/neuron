package internal

import (
	"time"
)

type BadModel struct {
	ID int `neuron:"typeprimary"`
}

type ModelNonTagged struct {
	ID int
}

type UnmarshalModel struct {
	ID          string     `neuron:"type=primary"`
	PtrString   *string    `neuron:"type=attr"`
	PtrTime     *time.Time `neuron:"type=attr"`
	StringSlice []string   `neuron:"type=attr"`
}

type ModelBadTypes struct {
	ID           string     `neuron:"type=primary"`
	StringField  string     `neuron:"type=attr;name=string_field"`
	FloatField   float64    `neuron:"type=attr;name=float_field"`
	TimeField    time.Time  `neuron:"type=attr;name=time_field"`
	TimePtrField *time.Time `neuron:"type=attr;name=time_ptr_field"`
}

type WithPointer struct {
	ID       *uint64  `neuron:"type=primary"`
	Name     *string  `neuron:"type=attr;name=name"`
	IsActive *bool    `neuron:"type=attr;name=is-active"`
	IntVal   *int     `neuron:"type=attr;name=int-val"`
	FloatVal *float32 `neuron:"type=attr;name=float-val"`
}

type Timestamp struct {
	ID   int        `neuron:"type=primary"`
	Time time.Time  `neuron:"type=attr;name=timestamp;flags=iso8601"`
	Next *time.Time `neuron:"type=attr;name=next;flags=iso8601"`
}

type NonRelatedModel struct {
	ID   int    `neuron:"type=primary"`
	Name string `neuron:"type=attr;name=name"`
}

type NoPrimaryModel struct {
	ID   int
	Name string `neuron:"type=attr;name=name"`
}

type User struct {
	privateField int
	ID           int    `neuron:"type=primary"`
	Lang         string `neuron:"type=attr;name=lang;flags=langtag"`
	Name         string `neuron:"type=attr;name=name"`
	Pets         []*Pet `neuron:"type=relation;name=pets;relation=many2many,sync,Owners"`
}

type Pet struct {
	ID     int     `neuron:"type=primary"`
	Name   string  `neuron:"type=attr;name=name"`
	Owners []*User `neuron:"type=relation;name=owners;relation=many2many,sync,Pets"`
}

/* HasMany Example */

type Driver struct {
	ID            int     `neuron:"type=primary"`
	Name          string  `neuron:"type=attr;flags=omitempty"`
	Age           int     `neuron:"type=attr;flags=omitempty"`
	Cars          []*Car  `neuron:"type=relation"`
	FavoriteCar   Car     `neuron:"type=relation;name=favorite-car;foreign=FavoriteCarID"`
	FavoriteCarID *string `neuron:"type=foreign;name=favorite_car_id"`
}

// at first check if FieldWithID does exists

// the relation would be

type Car struct {
	ID               *string `neuron:"type=primary"`
	Make             *string `neuron:"type=attr;name=make;flags=omitempty"`
	Model            *string `neuron:"type=attr;name=model;flags=omitempty"`
	Year             *uint   `neuron:"type=attr;name=year;flags=omitempty"`
	DriverID         int     `neuron:"type=foreign;name=driver_id"`
	somePrivateField *uint
}

type Blog struct {
	ID            int       `neuron:"type=primary"`
	Title         string    `neuron:"type=attr;name=title"`
	Posts         []*Post   `neuron:"type=relation;name=posts"`
	CurrentPost   *Post     `neuron:"type=relation;name=current_post"`
	CurrentPostID uint64    `neuron:"type=foreign;name=current_post_id"`
	CreatedAt     time.Time `neuron:"type=attr;name=created_at;flags=iso8601"`
	ViewCount     int       `neuron:"type=attr;name=view_count;flags=omitempty"`
}

type Post struct {
	ID            uint64     `neuron:"type=primary"`
	BlogID        int        `neuron:"type=foreign;name=blog_id"`
	Title         string     `neuron:"type=attr;name=title"`
	Body          string     `neuron:"type=attr;name=body"`
	Comments      []*Comment `neuron:"type=relation;name=comments;foreign=PostID"`
	LatestComment *Comment   `neuron:"type=relation;name=latest_comment;foreign=PostID"`
}

type Comment struct {
	ID     int    `neuron:"type=primary"`
	PostID uint64 `neuron:"type=foreign;name=post_id"`
	Body   string `neuron:"type=attr;name=body"`
}

type Book struct {
	ID          uint64  `neuron:"type=primary"`
	Author      string  `neuron:"type=attr;name=author"`
	ISBN        string  `neuron:"type=attr;name=isbn"`
	Title       string  `neuron:"type=attr;name=title;flags=omitempty"`
	Description *string `neuron:"type=attr;name=description"`
	Pages       *uint   `neuron:"type=attr;name=pages;flags=omitempty"`
	PublishedAt time.Time
	Tags        []string `neuron:"type=attr;name=tags"`
}

type RelationOnBasic struct {
	ID            int    `neuron:"type=primary"`
	BasicRelation string `neuron:"type=relation;name=basicrelation"`
}

func (c *RelationOnBasic) CollectionName() string {
	return "relationonbasics"
}

type RelationBasicOnPtr struct {
	ID               int     `neuron:"type=primary"`
	BasicPtrRelation *string `neuron:"type=relation;name=basicptrrelation"`
}

func (c *RelationBasicOnPtr) CollectionName() string {
	return "relationonbasicptr"
}

type Modeli18n struct {
	ID   int    `neuron:"type=primary"`
	Name string `neuron:"type=attr;name=name;flags=i18n"`
	Lang string `neuron:"type=attr;name=langcode;flags=langtag"`
}

func (m *Modeli18n) CollectionName() string {
	return "translateable"
}

// func (b *Blog) JSONAPILinks() *Links {
// 	return &Links{
// 		"self": fmt.Sprintf("https://example.com/api/blogs/%d", b.ID),
// 		"comments": Link{
// 			Href: fmt.Sprintf("https://example.com/api/blogs/%d/comments", b.ID),
// 			Meta: Meta{
// 				"counts": map[string]uint{
// 					"likes":    4,
// 					"comments": 20,
// 				},
// 			},
// 		},
// 	}
// }

// func (b *Blog) JSONAPIRelationshipLinks(relation string) *Links {
// 	if relation == "posts" {
// 		return &Links{
// 			"related": Link{
// 				Href: fmt.Sprintf("https://example.com/api/blogs/%d/posts", b.ID),
// 				Meta: Meta{
// 					"count": len(b.Posts),
// 				},
// 			},
// 		}
// 	}
// 	if relation == "current_post" {
// 		return &Links{
// 			"self": fmt.Sprintf("https://example.com/api/posts/%s", "3"),
// 			"related": Link{
// 				Href: fmt.Sprintf("https://example.com/api/blogs/%d/current_post", b.ID),
// 			},
// 		}
// 	}
// 	return nil
// }

// func (b *Blog) JSONAPIMeta() *Meta {
// 	return &Meta{
// 		"detail": "extra details regarding the blog",
// 	}
// }

// func (b *Blog) JSONAPIRelationshipMeta(relation string) *Meta {
// 	if relation == "posts" {
// 		return &Meta{
// 			"this": map[string]interface{}{
// 				"can": map[string]interface{}{
// 					"go": []interface{}{
// 						"as",
// 						"deep",
// 						map[string]interface{}{
// 							"as": "required",
// 						},
// 					},
// 				},
// 			},
// 		}
// 	}
// 	if relation == "current_post" {
// 		return &Meta{
// 			"detail": "extra current_post detail",
// 		}
// 	}
// 	return nil
// }

// type BadComment struct {
// 	ID   uint64 `neuron:"primary,bad-comment"`
// 	Body string `neuron:"attr,body"`
// }

// func (bc *BadComment) JSONAPILinks() *Links {
// 	return &Links{
// 		"self": []string{"invalid", "should error"},
// 	}
// }

type ModelI18nSDK struct {
	ID   int    `neuron:"type=primary"`
	Lang string `neuron:"type=attr;name=language;flags=langtag"`
}

func (m *ModelI18nSDK) CollectionName() string {
	return "i18n"
}

type ModelSDK struct {
	ID   int    `neuron:"type=primary"`
	Name string `neuron:"type=attr"`
}

func (c ModelSDK) CollectionName() string {
	return "models"
}

type AuthorSDK struct {
	ID    int        `neuron:"type=primary"`
	Name  string     `neuron:"type=attr"`
	Blogs []*BlogSDK `neuron:"type=relation;foreign=AuthorID"`
}

func (c AuthorSDK) CollectionName() string {
	return "authors"
}

type BlogSDK struct {
	ID                int      `neuron:"type=primary"`
	Lang              string   `neuron:"type=attr;flags=langtag"`
	SomeAttr          string   `neuron:"type=attr"`
	AuthorID          int      `neuron:"type=foreign"`
	CurrentPost       *PostSDK `neuron:"type=relation;foreign=BlogID"`
	CurrentPostNoSync *PostSDK `neuron:"type=relation;foreign=BlogIDNoSync;relation=nosync"`
}

func (c BlogSDK) CollectionName() string {
	return "blogs"
}

type PostSDK struct {
	ID             int           `neuron:"type=primary"`
	Title          string        `neuron:"type=attr"`
	BlogID         int           `neuron:"type=foreign"`
	BlogIDNoSync   int           `neuron:"type=foreign"`
	CreatedAt      time.Time     `neuron:"type=attr"`
	Comments       []*CommentSDK `neuron:"type=relation;foreign=PostID"`
	CommentsNoSync []*CommentSDK `neuron:"type=relation;foreign=PostIDNoSync;relation=nosync"`
}

func (c PostSDK) CollectionName() string {
	return "posts"
}

type CommentSDK struct {
	ID           int      `neuron:"type=primary"`
	Body         string   `neuron:"type=attr"`
	Post         *PostSDK `neuron:"type=relation;foreign=PostID"`
	PostID       int      `neuron:"type=foreign"`
	PostIDNoSync int      `neuron:"type=foreign"`
}

func (c CommentSDK) CollectionName() string {
	return "comments"
}

type PetSDK struct {
	ID         int         `neuron:"type=primary"`
	Name       string      `neuron:"type=attr"`
	Humans     []*HumanSDK `neuron:"type=relation;relation=many2many"`
	HumansSync []*HumanSDK `neuron:"type=relation;relation=many2many,sync,Pets"`
	Legs       int         `neuron:"type=attr"`
}

func (c PetSDK) CollectionName() string {
	return "pets"
}

type HumanSDK struct {
	ID   int       `neuron:"type=primary"`
	Name string    `neuron:"type=attr"`
	Pets []*PetSDK `neuron:"type=relation;relation=many2many"`
}

func (c HumanSDK) CollectionName() string {
	return "humans"
}

type ModCliGenID struct {
	ID string `neuron:"type=primary;flags=client-id"`
}

func (m *ModCliGenID) CollectionName() string {
	return "client-generated"
}
