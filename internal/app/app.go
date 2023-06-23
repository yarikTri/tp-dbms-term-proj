package app

import (
	"github.com/yarikTri/dbms-term-proj/internal/models"
)

type Repository interface {
	InsertUser(user models.User) error
	SelectUserByNickname(nickname string) (models.User, error)
	SelectUserByEmail(email string) (models.User, error)
	UpdateUser(user models.User) (models.User, error)
	SelectUsersByNickAndEmail(nickname, email string) ([]models.User, error)

	InsertForum(forum models.Forum) (models.Forum, error)
	SelectForumBySlug(slug string) (models.Forum, error)
	InsertThread(thread models.Thread) (models.Thread, error)
	SelectThreadBySlug(slug string) (models.Thread, error)
	SelectThreadById(id int) (models.Thread, error)
	InsertPosts(posts []models.Post, thread int) ([]models.Post, error)
	UpdateThread(thread models.Thread) (models.Thread, error)
	InsertVote(vote models.Vote) (models.Vote, error)
	UpdateVote(vote models.Vote) (models.Vote, error)
	GetServiceStatus() (map[string]int, error)
	ClearDatabase() error
	SelectUsersByForum(slugForum string, parameters models.QueryParameters) ([]models.User, error)
	SelectThreadsByForum(slugForum string, parameters models.QueryParameters) ([]models.Thread, error)
	SelectPostById(id int) (models.Post, error)
	UpdatePost(id int, message string) (models.Post, error)
	SelectPostsByThread(thread models.Thread, limit, since int, sort string, desc bool) ([]models.Post, error)
	SelectThreadByForum(forum string) (models.Thread, error)

	SelectThreadIdBySlug(slug string) (int, error)
}

type UseCase interface {
	CreateUser(user models.User) (models.User, error)
	CheckUserByEmail(email string) (models.User, error)
	CheckUserByNickname(nickname string) (models.User, error)
	HasUser(user models.User) ([]models.User, error)
	EditUser(newUser models.User) (models.User, error)

	CreateForum(forum models.Forum) (models.Forum, error)
	CheckForumBySlug(slug string) (models.Forum, error)
	CreateForumThread(thread models.Thread) (models.Thread, error)
	CheckThreadBySlug(slug string) (models.Thread, error)
	CheckThreadById(id int) (models.Thread, error)
	CreatePosts(posts []models.Post, id int) ([]models.Post, error)
	EditThread(thread models.Thread) (models.Thread, error)
	AddVote(vote models.Vote) (models.Vote, error)
	UpdateVote(vote models.Vote) (models.Vote, error)
	GetServiceStatus() (map[string]int, error)
	ClearDatabase() error
	CheckUsersByForum(slugForum string, parameters models.QueryParameters) ([]models.User, error)
	CheckThreadsByForum(slugForum string, parameters models.QueryParameters) ([]models.Thread, error)
	CheckPostById(id int, related []string) (map[string]interface{}, error)
	EditPost(id int, message string) (models.Post, error)
	CheckPostsByThread(thread models.Thread, limit, since int, sort string, desc bool) ([]models.Post, error)
	CheckThreadByForum(forum string) (models.Thread, error)

	CheckThreadIdBySlug(slug string) (int, error)
}
