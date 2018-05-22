package jsonapi

import (
	"time"
)

type BadModel struct {
	ID int `jsonapi:"primary"`
}

type ModelNonTagged struct {
	ID int
}

type ModelBadTypes struct {
	ID           string     `jsonapi:"primary,badtypes"`
	StringField  string     `jsonapi:"attr,string_field"`
	FloatField   float64    `jsonapi:"attr,float_field"`
	TimeField    time.Time  `jsonapi:"attr,time_field"`
	TimePtrField *time.Time `jsonapi:"attr,time_ptr_field"`
}

type WithPointer struct {
	ID       *uint64  `jsonapi:"primary,with-pointers"`
	Name     *string  `jsonapi:"attr,name"`
	IsActive *bool    `jsonapi:"attr,is-active"`
	IntVal   *int     `jsonapi:"attr,int-val"`
	FloatVal *float32 `jsonapi:"attr,float-val"`
}

type Timestamp struct {
	ID   int        `jsonapi:"primary,timestamps"`
	Time time.Time  `jsonapi:"attr,timestamp,iso8601"`
	Next *time.Time `jsonapi:"attr,next,iso8601"`
}

type NonRelatedModel struct {
	ID   int    `jsonapi:"primary,nonrelted"`
	Name string `jsonapi:"attr,name"`
}

type NoPrimaryModel struct {
	ID   int
	Name string `jsonapi:"attr,name"`
}

type User struct {
	privateField int
	ID           int    `jsonapi:"primary,users"`
	Lang         string `jsonapi:lang"`
	Name         string `jsonapi:"attr,name"`
	Pets         []*Pet `jsonapi:"relation,pets"`
}

type Pet struct {
	ID     int     `jsonapi:"primary,pets"`
	Name   string  `jsonapi:"attr,name"`
	Owners []*User `jsonapi:"relation,owners"`
}

/* HasMany Example */

type Driver struct {
	ID          int    `jsonapi:"primary,drivers"`
	Name        string `jsonapi:"attr,name,omitempty"`
	Age         int    `jsonapi:"attr,age,omitempty"`
	Cars        []*Car `jsonapi:"relation,cars"`
	FavoriteCar Car    `jsonapi:"relation,favorite-car"`
}

type Car struct {
	ID               *string `jsonapi:"primary,cars"`
	Make             *string `jsonapi:"attr,make,omitempty"`
	Model            *string `jsonapi:"attr,model,omitempty"`
	Year             *uint   `jsonapi:"attr,year,omitempty"`
	DriverID         int     `jsonapi:"attr,driver_id,omitempty"`
	somePrivateField *uint
}

type Blog struct {
	ID            int       `jsonapi:"primary,blogs"`
	ClientID      string    `jsonapi:"client-id"`
	Title         string    `jsonapi:"attr,title"`
	Posts         []*Post   `jsonapi:"relation,posts"`
	CurrentPost   *Post     `jsonapi:"relation,current_post"`
	CurrentPostID int       `jsonapi:"attr,current_post_id"`
	CreatedAt     time.Time `jsonapi:"attr,created_at,iso8601"`
	ViewCount     int       `jsonapi:"attr,view_count"`
}

type Post struct {
	ID            uint64     `jsonapi:"primary,posts"`
	BlogID        int        `jsonapi:"attr,blog_id"`
	ClientID      string     `jsonapi:"client-id"`
	Title         string     `jsonapi:"attr,title"`
	Body          string     `jsonapi:"attr,body"`
	Comments      []*Comment `jsonapi:"relation,comments"`
	LatestComment *Comment   `jsonapi:"relation,latest_comment"`
}

type Comment struct {
	ID       int    `jsonapi:"primary,comments"`
	ClientID string `jsonapi:"client-id"`
	PostID   int    `jsonapi:"attr,post_id"`
	Body     string `jsonapi:"attr,body"`
}

type Book struct {
	ID          uint64  `jsonapi:"primary,books"`
	Author      string  `jsonapi:"attr,author"`
	ISBN        string  `jsonapi:"attr,isbn"`
	Title       string  `jsonapi:"attr,title,omitempty"`
	Description *string `jsonapi:"attr,description"`
	Pages       *uint   `jsonapi:"attr,pages,omitempty"`
	PublishedAt time.Time
	Tags        []string `jsonapi:"attr,tags"`
}

type RelationOnBasic struct {
	ID            int    `jsonapi:"primary,relationonbasics"`
	BasicRelation string `jsonapi:"relation,basicrelation"`
}

type RelationBasicOnPtr struct {
	ID               int     `jsonapi:"primary,relationonbasicptr"`
	BasicPtrRelation *string `jsonapi:"relation,basicptrrelation"`
}

type Modeli18n struct {
	ID   int    `jsonapi:"primary,translateable"`
	Name string `jsonapi:"attr,name,i18n"`
	Lang string `jsonapi:"attr,langcode,langtag"`
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
