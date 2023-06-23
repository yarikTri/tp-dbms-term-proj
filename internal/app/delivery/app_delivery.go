package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/yarikTri/dbms-term-proj/internal/app"
	"github.com/yarikTri/dbms-term-proj/internal/models"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
)

type AppHandler struct {
	appUseCase app.UseCase
}

func NewAppHandler(router *mux.Router, appUseCase app.UseCase) {
	handler := &AppHandler{
		appUseCase: appUseCase,
	}

	router.HandleFunc("/api/user/{nickname}/create", handler.CreateUser).Methods(http.MethodPost)
	router.HandleFunc("/api/user/{nickname}/profile", handler.UserProfile).Methods(http.MethodGet, http.MethodPost)

	router.HandleFunc("/api/forum/create", handler.CreateForum).Methods(http.MethodPost)
	router.HandleFunc("/api/forum/{slug}/details", handler.ForumDetails).Methods(http.MethodGet)
	router.HandleFunc("/api/forum/{slug}/create", handler.CreateThread).Methods(http.MethodPost)
	router.HandleFunc("/api/forum/{slug}/threads", handler.ForumThreads).Methods(http.MethodGet)
	router.HandleFunc("/api/forum/{slug}/users", handler.ForumUsers).Methods(http.MethodGet)

	router.HandleFunc("/api/thread/{slug_or_id}/create", handler.CreatePosts).Methods(http.MethodPost)
	router.HandleFunc("/api/thread/{slug_or_id}/vote", handler.VoteThread).Methods(http.MethodPost)
	router.HandleFunc("/api/thread/{slug_or_id}/details", handler.ThreadDetails).Methods(http.MethodGet, http.MethodPost)

	router.HandleFunc("/api/thread/{slug_or_id}/posts", handler.ThreadPosts).Methods(http.MethodGet)

	router.HandleFunc("/api/post/{id}/details", handler.PostDetails).Methods(http.MethodGet, http.MethodPost)

	router.HandleFunc("/api/service/status", handler.StatusHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/service/clear", handler.ClearHandler).Methods(http.MethodPost)
}

func errorMarshal(message string) ([]byte, error) {
	e := models.Error{Message: message}
	body, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (h AppHandler) CreateUser(writer http.ResponseWriter, request *http.Request) {
	nickname := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/api/user/"), "/create")

	var user models.User
	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		return
	}
	user.Nickname = nickname

	_, err = h.appUseCase.CreateUser(user)
	if err != nil {
		if users, err := h.appUseCase.HasUser(user); err == nil {
			body, err := json.Marshal(users)
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusConflict)
			writer.Write(body)

			return
		}

		return
	}

	body, err := json.Marshal(user)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusCreated)
	writer.Write(body)
}

func (h AppHandler) UserProfile(writer http.ResponseWriter, request *http.Request) {
	nickname := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/api/user/"), "/profile")

	if request.Method == "GET" {
		user, err := h.appUseCase.CheckUserByNickname(nickname)
		if err != nil {
			body, err := errorMarshal("Can't find user\n")
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusNotFound)
			writer.Write(body)

			return
		}
		body, err := json.Marshal(user)
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write(body)

		return
	}

	var user models.User
	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		return
	}
	user.Nickname = nickname

	result, err := h.appUseCase.EditUser(user)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok && pgErr.Code == "23505" {
			body, err := errorMarshal("Conflict email\n")
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusConflict)
			writer.Write(body)

			return
		}

		body, err := errorMarshal("Can't find user\n")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)

		return
	}
	// check conflict with constraint unique on email

	body, err := json.Marshal(result)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}

func (h AppHandler) CreateForum(writer http.ResponseWriter, request *http.Request) {
	var forum models.Forum
	err := json.NewDecoder(request.Body).Decode(&forum)
	if err != nil {
		return
	}

	user, err := h.appUseCase.CheckUserByNickname(forum.User)
	if err != nil {
		body, err := errorMarshal("Can't find user")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)

		return
	}

	forum.User = user.Nickname

	f, err := h.appUseCase.CreateForum(forum)
	if pgErr, ok := err.(pgx.PgError); ok {
		switch pgErr.Code {
		case "23505":
			forumSlug, err := h.appUseCase.CheckForumBySlug(forum.Slug)
			if err != nil {
				return
			}

			body, err := json.Marshal(forumSlug)
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusConflict)
			writer.Write(body)
		case "23503":
			body, err := errorMarshal("Can't find user")
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusNotFound)
			writer.Write(body)
		}

		return
	}

	body, err := json.Marshal(f)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusCreated)
	writer.Write(body)
}

func (h AppHandler) ForumDetails(writer http.ResponseWriter, request *http.Request) {
	slug := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/api/forum/"), "/details")

	forum, err := h.appUseCase.CheckForumBySlug(slug)
	if err != nil {
		body, err := errorMarshal("Can't find forum")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)

		return
	}

	body, err := json.Marshal(forum)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}

func (h AppHandler) CreateThread(writer http.ResponseWriter, request *http.Request) {
	slug := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/api/forum/"), "/create")

	thread := models.Thread{Forum: slug}
	err := json.NewDecoder(request.Body).Decode(&thread)
	if err != nil {
		return
	}

	flag := thread.Slug == ""

	newThread, err := h.appUseCase.CreateForumThread(thread)
	if pgErr, ok := err.(pgx.PgError); ok && pgErr.Code == "23505" {
		oldThread, err := h.appUseCase.CheckThreadBySlug(thread.Slug)
		if err != nil {
			return
		}

		body, err := json.Marshal(oldThread)
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusConflict)
		writer.Write(body)

		return
	}

	if err != nil {
		body, err := errorMarshal("Can't find forum or user\n")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)

		return
	}

	forum, err := h.appUseCase.CheckForumBySlug(slug)
	if err != nil {
		body, err := errorMarshal("Can't find forum or user\n")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)

		return
	}

	if flag {
		threadWithoutSlug := models.ThreadWithoutSlug{
			Id:      newThread.Id,
			Author:  newThread.Author,
			Created: newThread.Created,
			Forum:   forum.Slug,
			Title:   newThread.Title,
			Message: newThread.Message,
			Votes:   newThread.Votes,
		}

		body, err := json.Marshal(threadWithoutSlug)
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusCreated)
		writer.Write(body)

		return
	}

	newThread.Forum = forum.Slug

	body, err := json.Marshal(newThread)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusCreated)
	writer.Write(body)
}

func (h AppHandler) CreatePosts(writer http.ResponseWriter, request *http.Request) {
	var posts []models.Post
	err := json.NewDecoder(request.Body).Decode(&posts)
	if err != nil {
		return
	}

	slugOrId := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/api/thread/"), "/create")

	var id int
	var thread models.Thread
	id, err = strconv.Atoi(slugOrId)
	if err != nil {
		id, err = h.appUseCase.CheckThreadIdBySlug(slugOrId)
		if err != nil {
			body, err := errorMarshal("Haven't this thread")
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusNotFound)
			writer.Write(body)

			return
		}
	} else {
		thread, err = h.appUseCase.CheckThreadById(id)
		if err != nil {
			body, err := errorMarshal("Haven't this thread")
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusNotFound)
			writer.Write(body)

			return
		}
	}

	if len(posts) == 0 {
		body, err := json.Marshal(posts)
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusCreated)
		writer.Write(body)

		return
	}

	author := posts[0].Author

	resultPosts, err := h.appUseCase.CreatePosts(posts, id)
	if len(resultPosts) == 0 {
		err = pgx.ErrNoRows
	}
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			if pgErr.Code == "00409" {
				body, err := errorMarshal("conflict")
				if err != nil {
					return
				}

				writer.WriteHeader(http.StatusConflict)
				writer.Write(body)

				return
			}
		}

		if err == pgx.ErrNoRows {
			_, err := h.appUseCase.CheckUserByNickname(author)
			if err == pgx.ErrNoRows {
				body, err := errorMarshal("Haven't this user")
				if err != nil {
					return
				}

				writer.WriteHeader(http.StatusNotFound)
				writer.Write(body)

				return
			}

			if thread.Title == "" {
				body, err := errorMarshal("conflict")
				if err != nil {
					return
				}

				writer.WriteHeader(http.StatusConflict)
				writer.Write(body)

				return
			}

			body, err := errorMarshal("conflict")
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusConflict)
			writer.Write(body)

			return
		}
	}

	body, err := json.Marshal(resultPosts)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusCreated)
	writer.Write(body)
}

func (h AppHandler) ThreadDetails(writer http.ResponseWriter, request *http.Request) {
	slugOrId := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/api/thread/"), "/details")

	var thread models.Thread
	if request.Method == "GET" {
		id, err := strconv.Atoi(slugOrId)
		if err != nil {
			thread, err = h.appUseCase.CheckThreadBySlug(slugOrId)
		} else {
			thread, err = h.appUseCase.CheckThreadById(id)
		}

		if err != nil {
			body, err := errorMarshal("can't find thread")
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusNotFound)
			writer.Write(body)

			return
		}

		if models.IsUUID(thread.Slug) {
			result := models.ThreadToWithout(thread)

			body, err := json.Marshal(result)
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusOK)
			writer.Write(body)

			return
		}

		body, err := json.Marshal(thread)
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write(body)

		return
	}

	err := json.NewDecoder(request.Body).Decode(&thread)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		thread.Slug = slugOrId
	} else {
		thread.Id = id
	}

	newThread, err := h.appUseCase.EditThread(thread)
	if err != nil {
		body, err := errorMarshal("can't find thread")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)

		return
	}

	if models.IsUUID(newThread.Slug) {
		result := models.ThreadToWithout(newThread)

		body, err := json.Marshal(result)
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write(body)

		return
	}

	body, err := json.Marshal(newThread)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}

func (h AppHandler) VoteThread(writer http.ResponseWriter, request *http.Request) {
	slugOrId := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/api/thread/"), "/vote")

	var vote models.Vote
	err := json.NewDecoder(request.Body).Decode(&vote)
	if err != nil {
		return
	}

	var thread models.Thread
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		thread, err = h.appUseCase.CheckThreadBySlug(slugOrId)
	}

	if err != nil {
		body, err := errorMarshal("can't find thread")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)

		return
	}

	if thread.Id != 0 {
		id = thread.Id
	}

	vote.IdThread = id

	_, err = h.appUseCase.AddVote(vote)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok && pgErr.Code == "23505" {
			_, err := h.appUseCase.UpdateVote(vote)
			if err != nil {
				body, err := errorMarshal("can't find thread or this")
				if err != nil {
					return
				}

				writer.WriteHeader(http.StatusNotFound)
				writer.Write(body)

				return
			}
		} else {
			body, err := errorMarshal("can't find thread this")
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusNotFound)
			writer.Write(body)

			return
		}
	}

	thread, err = h.appUseCase.CheckThreadById(id)

	if models.IsUUID(thread.Slug) {
		result := models.ThreadToWithout(thread)

		body, err := json.Marshal(result)
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write(body)

		return
	}

	body, err := json.Marshal(thread)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}

func (h AppHandler) StatusHandler(writer http.ResponseWriter, request *http.Request) {
	info, err := h.appUseCase.GetServiceStatus()
	if err != nil {
		body, err := errorMarshal("vse ploho")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)
	}

	body, err := json.Marshal(info)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}

func (h AppHandler) ClearHandler(writer http.ResponseWriter, request *http.Request) {
	err := h.appUseCase.ClearDatabase()
	if err != nil {
		body, err := errorMarshal("ochen ploho")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)
	}

	writer.WriteHeader(http.StatusOK)
}

func (h AppHandler) ForumUsers(writer http.ResponseWriter, request *http.Request) {
	var parameters models.QueryParameters
	limit, err := strconv.Atoi(request.URL.Query().Get("limit"))
	if err != nil {
		limit = 100
	}
	parameters.Limit = limit

	since := request.URL.Query().Get("since")
	parameters.Since = since

	desc, err := strconv.ParseBool(request.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}
	parameters.Desc = desc

	slug := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/api/forum/"), "/users")

	users, err := h.appUseCase.CheckUsersByForum(slug, parameters)
	if err != nil {
		body, err := errorMarshal("can't find something")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)

		return
	}

	if users == nil {
		_, err := h.appUseCase.CheckForumBySlug(slug)
		if err != nil {
			body, err := errorMarshal("can't find something")
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusNotFound)
			writer.Write(body)

			return
		}
		body, err := json.Marshal([]int{})
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write(body)

		return
	}

	body, err := json.Marshal(users)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}

func (h AppHandler) ForumThreads(writer http.ResponseWriter, request *http.Request) {
	var parameters models.QueryParameters
	limit, err := strconv.Atoi(request.URL.Query().Get("limit"))
	if err != nil {
		limit = 0
	}
	parameters.Limit = limit

	since := request.URL.Query().Get("since")
	parameters.Since = since

	desc, err := strconv.ParseBool(request.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}
	parameters.Desc = desc

	slug := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/api/forum/"), "/threads")

	threads, err := h.appUseCase.CheckThreadsByForum(slug, parameters)
	if err == pgx.ErrNoRows || len(threads) == 0 {
		_, err := h.appUseCase.CheckThreadByForum(slug)
		if err == nil {
			body, err := json.Marshal([]int{})
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusOK)
			writer.Write(body)

			return
		}

		body, err := errorMarshal("can't find something bad")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)

		return
	}

	var result []interface{}
	for _, thr := range threads {
		if models.IsUUID(thr.Slug) {
			tw := models.ThreadToWithout(thr)

			result = append(result, tw)
		} else {
			result = append(result, thr)
		}
	}

	body, err := json.Marshal(result)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}

func (h AppHandler) PostDetails(writer http.ResponseWriter, request *http.Request) {
	id, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/api/post/"), "/details"))
	if err != nil {
		return
	}

	if request.Method == "GET" {
		related := strings.Split(request.URL.Query().Get("related"), ",")

		data, err := h.appUseCase.CheckPostById(id, related)
		if err != nil {
			body, err := errorMarshal("can't find something")
			if err != nil {
				return
			}

			writer.WriteHeader(http.StatusNotFound)
			writer.Write(body)

			return
		}

		body, err := json.Marshal(data)
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write(body)

		return
	}

	var post models.Post
	err = json.NewDecoder(request.Body).Decode(&post)
	if err != nil {
		return
	}

	post, err = h.appUseCase.EditPost(id, post.Message)
	if err != nil {
		body, err := errorMarshal("can't find something")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)

		return
	}

	body, err := json.Marshal(post)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}

func (h AppHandler) ThreadPosts(writer http.ResponseWriter, request *http.Request) {
	limit, err := strconv.Atoi(request.URL.Query().Get("limit"))
	if err != nil {
		limit = 0
	}

	since, err := strconv.Atoi(request.URL.Query().Get("since"))

	desc, err := strconv.ParseBool(request.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	sort := request.URL.Query().Get("sort")
	if sort == "" {
		sort = "flat"
	}

	slugOrId := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/api/thread/"), "/posts")
	var thread models.Thread
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		thread.Slug = slugOrId
	} else {
		thread.Id = id
	}

	posts, err := h.appUseCase.CheckPostsByThread(thread, limit, since, sort, desc)
	if err != nil {
		body, err := errorMarshal("can't find something this")
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusNotFound)
		writer.Write(body)

		return
	}

	if posts == nil {
		if thread.Id == 0 {
			_, err := h.appUseCase.CheckThreadBySlug(thread.Slug)
			if err == pgx.ErrNoRows {
				body, err := errorMarshal("can't find something")
				if err != nil {
					return
				}

				writer.WriteHeader(http.StatusNotFound)
				writer.Write(body)

				return
			}
		} else {
			_, err := h.appUseCase.CheckThreadById(thread.Id)
			if err == pgx.ErrNoRows {
				body, err := errorMarshal("can't find something")
				if err != nil {
					return
				}

				writer.WriteHeader(http.StatusNotFound)
				writer.Write(body)

				return
			}
		}

		body, err := json.Marshal([]int{})
		if err != nil {
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write(body)

		return
	}

	body, err := json.Marshal(posts)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(body)
}
