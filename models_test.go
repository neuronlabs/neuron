package jsonapi

import (
	"time"
)

type BadModel struct {
	ID int `jsonapi:"typeprimary"`
}

type ModelNonTagged struct {
	ID int
}

type ModelBadTypes struct {
	ID           string     `jsonapi:"type=primary"`
	StringField  string     `jsonapi:"type=attr;name=string_field"`
	FloatField   float64    `jsonapi:"type=attr;name=float_field"`
	TimeField    time.Time  `jsonapi:"type=attr;name=time_field"`
	TimePtrField *time.Time `jsonapi:"type=attr;name=time_ptr_field"`
}

type WithPointer struct {
	ID       *uint64  `jsonapi:"type=primary"`
	Name     *string  `jsonapi:"type=attr;name=name"`
	IsActive *bool    `jsonapi:"type=attr;name=is-active"`
	IntVal   *int     `jsonapi:"type=attr;name=int-val"`
	FloatVal *float32 `jsonapi:"type=attr;name=float-val"`
}

type Timestamp struct {
	ID   int        `jsonapi:"type=primary"`
	Time time.Time  `jsonapi:"type=attr;name=timestamp;flags=iso8601"`
	Next *time.Time `jsonapi:"type=attr;name=next;flags=iso8601"`
}

type NonRelatedModel struct {
	ID   int    `jsonapi:"type=primary"`
	Name string `jsonapi:"type=attr;name=name"`
}

type NoPrimaryModel struct {
	ID   int
	Name string `jsonapi:"type=attr;name=name"`
}

type User struct {
	privateField int
	ID           int    `jsonapi:"type=primary"`
	Lang         string `jsonapi:"type=attr;name=lang;flags=langtag"`
	Name         string `jsonapi:"type=attr;name=name"`
	Pets         []*Pet `jsonapi:"type=relation;name=pets;relation=many2many,sync,Owners"`
}

type Pet struct {
	ID     int     `jsonapi:"type=primary"`
	Name   string  `jsonapi:"type=attr;name=name"`
	Owners []*User `jsonapi:"type=relation;name=owners;relation=many2many,sync,Pets"`
}

/* HasMany Example */

type Driver struct {
	ID            int     `jsonapi:"type=primary"`
	Name          string  `jsonapi:"type=attr;flags=omitempty"`
	Age           int     `jsonapi:"type=attr;flags=omitempty"`
	Cars          []*Car  `jsonapi:"type=relation"`
	FavoriteCar   Car     `jsonapi:"type=relation;name=favorite-car;foreign=FavoriteCarID"`
	FavoriteCarID *string `jsonapi:"type=foreign;name=favorite_car_id"`
}

// at first check if FieldWithID does exists

// the relation would be

type Car struct {
	ID               *string `jsonapi:"type=primary"`
	Make             *string `jsonapi:"type=attr;name=make;flags=omitempty"`
	Model            *string `jsonapi:"type=attr;name=model;flags=omitempty"`
	Year             *uint   `jsonapi:"type=attr;name=year;flags=omitempty"`
	DriverID         int     `jsonapi:"type=foreign;name=driver_id"`
	somePrivateField *uint
}

type Blog struct {
	ID            int       `jsonapi:"type=primary"`
	Title         string    `jsonapi:"type=attr;name=title"`
	Posts         []*Post   `jsonapi:"type=relation;name=posts"`
	CurrentPost   *Post     `jsonapi:"type=relation;name=current_post"`
	CurrentPostID uint64    `jsonapi:"type=foreign;name=current_post_id"`
	CreatedAt     time.Time `jsonapi:"type=attr;name=created_at;flags=iso8601"`
	ViewCount     int       `jsonapi:"type=attr;name=view_count"`
}

type Post struct {
	ID            uint64     `jsonapi:"type=primary"`
	BlogID        int        `jsonapi:"type=foreign;name=blog_id"`
	Title         string     `jsonapi:"type=attr;name=title"`
	Body          string     `jsonapi:"type=attr;name=body"`
	Comments      []*Comment `jsonapi:"type=relation;name=comments;foreign=PostID"`
	LatestComment *Comment   `jsonapi:"type=relation;name=latest_comment;foreign=PostID"`
}

type Comment struct {
	ID     int    `jsonapi:"type=primary"`
	PostID uint64 `jsonapi:"type=foreign;name=post_id"`
	Body   string `jsonapi:"type=attr;name=body"`
}

type Book struct {
	ID          uint64  `jsonapi:"type=primary"`
	Author      string  `jsonapi:"type=attr;name=author"`
	ISBN        string  `jsonapi:"type=attr;name=isbn"`
	Title       string  `jsonapi:"type=attr;name=title;flags=omitempty"`
	Description *string `jsonapi:"type=attr;name=description"`
	Pages       *uint   `jsonapi:"type=attr;name=pages;flags=omitempty"`
	PublishedAt time.Time
	Tags        []string `jsonapi:"type=attr;name=tags"`
}

type RelationOnBasic struct {
	ID            int    `jsonapi:"type=primary"`
	BasicRelation string `jsonapi:"type=relation;name=basicrelation"`
}

func (c *RelationOnBasic) CollectionName() string {
	return "relationonbasics"
}

type RelationBasicOnPtr struct {
	ID               int     `jsonapi:"type=primary"`
	BasicPtrRelation *string `jsonapi:"type=relation;name=basicptrrelation"`
}

func (c *RelationBasicOnPtr) CollectionName() string {
	return "relationonbasicptr"
}

type Modeli18n struct {
	ID   int    `jsonapi:"type=primary"`
	Name string `jsonapi:"type=attr;name=name;flags=i18n"`
	Lang string `jsonapi:"type=attr;name=langcode;flags=langtag"`
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
// 	ID   uint64 `jsonapi:"primary,bad-comment"`
// 	Body string `jsonapi:"attr,body"`
// }

// func (bc *BadComment) JSONAPILinks() *Links {
// 	return &Links{
// 		"self": []string{"invalid", "should error"},
// 	}
// }
