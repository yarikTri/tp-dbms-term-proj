package usecase

import (
	"github.com/yarikTri/dbms-term-proj/internal/app"
	"github.com/yarikTri/dbms-term-proj/internal/models"

	"github.com/google/uuid"
)

type appUseCase struct {
	appRepository app.Repository
}

func NewAppUseCase(ar app.Repository) app.UseCase {
	return &appUseCase{
		appRepository: ar,
	}
}

func (a appUseCase) CreateUser(user models.User) (models.User, error) {
	err := a.appRepository.InsertUser(user)

	return user, err
}

func (a appUseCase) CheckUserByEmail(email string) (models.User, error) {
	user, err := a.appRepository.SelectUserByEmail(email)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (a appUseCase) CheckUserByNickname(nickname string) (models.User, error) {
	user, err := a.appRepository.SelectUserByNickname(nickname)

	return user, err
}

func (a appUseCase) HasUser(user models.User) ([]models.User, error) {
	users, err := a.appRepository.SelectUsersByNickAndEmail(user.Nickname, user.Email)

	return users, err
}

func (a appUseCase) EditUser(newUser models.User) (models.User, error) {
	u, err := a.appRepository.UpdateUser(newUser)

	return u, err
}

func (a appUseCase) CreateForum(forum models.Forum) (models.Forum, error) {
	f, err := a.appRepository.InsertForum(forum)
	if err != nil {
	}

	return f, err
}

func (a appUseCase) CheckForumBySlug(slug string) (models.Forum, error) {
	forum, err := a.appRepository.SelectForumBySlug(slug)
	if err != nil {
	}

	return forum, err
}

func (a appUseCase) CreateForumThread(thread models.Thread) (models.Thread, error) {
	if thread.Slug == "" {
		u, err := uuid.NewRandom()
		if err != nil {
			panic("AAAAAAAAAAAAAAAAA")
		}
		thread.Slug = u.String()
	}
	thr, err := a.appRepository.InsertThread(thread)

	return thr, err
}

func (a appUseCase) CheckThreadBySlug(slug string) (models.Thread, error) {
	thread, err := a.appRepository.SelectThreadBySlug(slug)

	return thread, err
}

func (a appUseCase) CheckThreadById(id int) (models.Thread, error) {
	thread, err := a.appRepository.SelectThreadById(id)

	return thread, err
}

func (a appUseCase) CreatePosts(posts []models.Post, id int) ([]models.Post, error) {
	result, err := a.appRepository.InsertPosts(posts, id)

	return result, err
}

func (a appUseCase) EditThread(thread models.Thread) (models.Thread, error) {
	newThread, err := a.appRepository.UpdateThread(thread)

	return newThread, err
}

func (a appUseCase) AddVote(vote models.Vote) (models.Vote, error) {
	newVote, err := a.appRepository.InsertVote(vote)

	return newVote, err
}

func (a appUseCase) UpdateVote(vote models.Vote) (models.Vote, error) {
	newVote, err := a.appRepository.UpdateVote(vote)

	return newVote, err
}

func (a appUseCase) GetServiceStatus() (map[string]int, error) {
	return a.appRepository.GetServiceStatus()
}

func (a appUseCase) ClearDatabase() error {
	return a.appRepository.ClearDatabase()
}

func (a appUseCase) CheckUsersByForum(slugForum string, parameters models.QueryParameters) ([]models.User, error) {
	users, err := a.appRepository.SelectUsersByForum(slugForum, parameters)

	return users, err
}

func (a appUseCase) CheckThreadsByForum(slugForum string, parameters models.QueryParameters) ([]models.Thread, error) {
	threads, err := a.appRepository.SelectThreadsByForum(slugForum, parameters)

	return threads, err
}

func (a appUseCase) CheckPostById(id int, related []string) (map[string]interface{}, error) {
	post, err := a.appRepository.SelectPostById(id)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"post": post,
	}

	for _, item := range related {
		switch item {
		case "forum":
			forum, err := a.appRepository.SelectForumBySlug(post.Forum)
			if err != nil {
				return nil, err
			}

			data["forum"] = forum
			break
		case "user":
			user, err := a.appRepository.SelectUserByNickname(post.Author)
			if err != nil {
				return nil, err
			}

			data["author"] = user
			break
		case "thread":
			thread, err := a.appRepository.SelectThreadById(post.Thread)
			if err != nil {
				return nil, err
			}

			if models.IsUUID(thread.Slug) {
				result := models.ThreadToWithout(thread)

				data["thread"] = result
			} else {
				data["thread"] = thread
			}

			break
		}
	}

	return data, nil
}

func (a appUseCase) EditPost(id int, message string) (models.Post, error) {
	post, err := a.appRepository.UpdatePost(id, message)

	return post, err
}

func (a appUseCase) CheckPostsByThread(thread models.Thread, limit, since int, sort string, desc bool) ([]models.Post, error) {
	posts, err := a.appRepository.SelectPostsByThread(thread, limit, since, sort, desc)

	return posts, err
}

func (a appUseCase) CheckThreadByForum(forum string) (models.Thread, error) {
	thread, err := a.appRepository.SelectThreadByForum(forum)

	return thread, err
}

func (a appUseCase) CheckThreadIdBySlug(slug string) (int, error) {
	id, err := a.appRepository.SelectThreadIdBySlug(slug)

	return id, err
}
