package models

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/pgtype"
)

type Error struct {
	Message string `json:"message"`
}

type User struct {
	About    string `json:"about"`
	Email    string `json:"email"`
	FullName string `json:"fullname"`
	Nickname string `json:"nickname"`
}

type Forum struct {
	Title   string `json:"title"`
	User    string `json:"user"`
	Slug    string `json:"slug"`
	Posts   int    `json:"posts"`
	Threads int    `json:"threads"`
}

type Thread struct {
	Id      int    `json:"id"`
	Author  string `json:"author"`
	Created string `json:"created"`
	Forum   string `json:"forum"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Slug    string `json:"slug"`
	Votes   int    `json:"votes"`
}

type ThreadWithoutSlug struct {
	Id      int    `json:"id"`
	Author  string `json:"author"`
	Created string `json:"created"`
	Forum   string `json:"forum"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Votes   int    `json:"votes"`
}

func ThreadToWithout(thread Thread) ThreadWithoutSlug {
	return ThreadWithoutSlug{
		Id:      thread.Id,
		Author:  thread.Author,
		Created: thread.Created,
		Forum:   thread.Forum,
		Title:   thread.Title,
		Message: thread.Message,
		Votes:   thread.Votes,
	}
}

type Post struct {
	Id       int              `json:"id"`
	Author   string           `json:"author"`
	Created  string           `json:"created"`
	Forum    string           `json:"forum"`
	Message  string           `json:"message"`
	IsEdited bool             `json:"isEdited"`
	Parent   JsonNullInt      `json:"parent"`
	Thread   int              `json:"thread"`
	Path     pgtype.Int8Array `json:"-"`
}

type JsonNullInt struct {
	sql.NullInt64
}

func (v JsonNullInt) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JsonNullInt) UnmarshalJSON(data []byte) error {
	var x *int64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

var PostParentError = `insert or update on table "post" violates foreign key constraint "post_parent_fkey"`

type Vote struct {
	Nickname string `json:"nickname"`
	Voice    int    `json:"voice"`
	IdThread int    `json:"-"`
}

type QueryParameters struct {
	Limit int
	Since string
	Desc  bool
}

func IsUUID(value string) bool {
	n := len(value)

	if n > 36 || n < 32 {
		return false
	}

	_, err := uuid.Parse(value)

	return err == nil
}
