package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	repo "github.com/yarikTri/dbms-term-proj/internal/app"
	"github.com/yarikTri/dbms-term-proj/internal/models"

	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx"
)

type postgresAppRepository struct {
	Conn *pgx.ConnPool
}

func NewPostgresAppRepository(conn *pgx.ConnPool) repo.Repository {
	return &postgresAppRepository{
		Conn: conn,
	}
}

func (p *postgresAppRepository) InsertUser(user models.User) error {
	_, err := p.Conn.Exec(`INSERT INTO users(nickname, fullname, about, email) VALUES ($1, $2, $3, $4)`, user.Nickname, user.FullName, user.About, user.Email)

	return err
}

func (p *postgresAppRepository) SelectUserByNickname(nickname string) (models.User, error) {
	row := p.Conn.QueryRow(`SELECT nickname, fullname, about, email FROM users WHERE nickname=$1 LIMIT 1;`, nickname)

	var user models.User
	err := row.Scan(&user.Nickname, &user.FullName, &user.About, &user.Email)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (p *postgresAppRepository) SelectUserByEmail(email string) (models.User, error) {
	row := p.Conn.QueryRow(`SELECT email, nickname, fullname, about FROM users WHERE email=$1 LIMIT 1;`, email)

	var user models.User
	err := row.Scan(&user.Email, &user.Nickname, &user.FullName, &user.About)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (p *postgresAppRepository) SelectUsersByNickAndEmail(nickname, email string) ([]models.User, error) {
	rows, err := p.Conn.Query(`SELECT * FROM users WHERE email=$1 OR nickname=$2 LIMIT 2;`, email, nickname)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.Nickname, &user.FullName, &user.About, &user.Email)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (p *postgresAppRepository) UpdateUser(user models.User) (models.User, error) {
	var newUser models.User
	err := p.Conn.QueryRow(
		`UPDATE users SET email=COALESCE(NULLIF($1, ''), email), 
							  about=COALESCE(NULLIF($2, ''), about), 
							  fullname=COALESCE(NULLIF($3, ''), fullname) WHERE nickname=$4 RETURNING *`,
		user.Email,
		user.About,
		user.FullName,
		user.Nickname,
	).Scan(&newUser.Nickname, &newUser.FullName, &newUser.About, &newUser.Email)

	return newUser, err
}

func (p *postgresAppRepository) InsertForum(forum models.Forum) (models.Forum, error) {
	var newForum models.Forum
	err := p.Conn.QueryRow(
		`INSERT INTO forum(slug, title, "user") VALUES ($1, $2, $3) RETURNING *`,
		forum.Slug,
		forum.Title,
		forum.User,
	).Scan(&newForum.Slug, &newForum.Title, &newForum.User, &newForum.Posts, &newForum.Threads)

	return newForum, err
}

func (p *postgresAppRepository) SelectForumBySlug(slug string) (models.Forum, error) {
	var forum models.Forum
	err := p.Conn.QueryRow(
		`SELECT * FROM forum WHERE slug=$1 LIMIT 1;`,
		slug).Scan(
		&forum.Slug,
		&forum.Title,
		&forum.User,
		&forum.Posts,
		&forum.Threads,
	)

	return forum, err
}

func (p *postgresAppRepository) InsertThread(thread models.Thread) (models.Thread, error) {
	query := `INSERT INTO thread(slug, author, created, message, title, forum) 
			  VALUES ($1, $2, $3, $4, $5, $6) RETURNING *`

	var row *pgx.Row
	if thread.Created != "" {
		row = p.Conn.QueryRow(
			query,
			thread.Slug,
			thread.Author,
			thread.Created,
			thread.Message,
			thread.Title,
			thread.Forum,
		)
	} else {
		row = p.Conn.QueryRow(
			query,
			thread.Slug,
			thread.Author,
			time.Time{},
			thread.Message,
			thread.Title,
			thread.Forum,
		)
	}

	var thr models.Thread
	var created time.Time
	err := row.Scan(&thr.Id, &thr.Author, &created, &thr.Forum, &thr.Message, &thr.Slug, &thr.Title, &thr.Votes)

	thr.Created = strfmt.DateTime(created.UTC()).String()

	return thr, err
}

func (p *postgresAppRepository) SelectThreadBySlug(slug string) (models.Thread, error) {
	row := p.Conn.QueryRow(`SELECT * FROM thread WHERE slug=$1 LIMIT 1;`, slug)

	var thread models.Thread
	var created time.Time
	err := row.Scan(&thread.Id, &thread.Author, &created, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)

	thread.Created = strfmt.DateTime(created.UTC()).String()

	return thread, err
}

func (p *postgresAppRepository) SelectThreadById(id int) (models.Thread, error) {
	row := p.Conn.QueryRow(`SELECT * FROM thread WHERE id=$1 LIMIT 1;`, id)

	var thread models.Thread
	var created time.Time
	err := row.Scan(&thread.Id, &thread.Author, &created, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)

	thread.Created = strfmt.DateTime(created.UTC()).String()

	return thread, err
}

func (p *postgresAppRepository) selectForumSlugById(id int) (string, error) {
	query := `SELECT forum FROM thread WHERE id=$1`

	var slug string
	err := p.Conn.QueryRow(query, id).Scan(&slug)
	return slug, err
}

func (p *postgresAppRepository) InsertPosts(posts []models.Post, thread int) ([]models.Post, error) {
	resultPosts := make([]models.Post, 0, 0)

	if len(posts) == 0 {
		return resultPosts, nil
	}

	forum, err := p.selectForumSlugById(thread)
	if err != nil {
		return nil, err
	}

	insert := `INSERT INTO post(author, created, forum, message, parent, thread) VALUES `
	var values []interface{}
	timeCreated := time.Now()
	for i, post := range posts {
		value := fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d),",
			i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6,
		)

		insert += value

		values = append(values, post.Author, timeCreated, forum, post.Message, post.Parent, thread)
	}

	insert = strings.TrimSuffix(insert, ",")
	insert += ` RETURNING *`

	rows, err := p.Conn.Query(insert, values...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// var resultPosts []models.Post
	for rows.Next() {
		var currentPost models.Post
		var created time.Time

		err := rows.Scan(
			&currentPost.Id,
			&currentPost.Author,
			&created,
			&currentPost.Forum,
			&currentPost.Message,
			&currentPost.IsEdited,
			&currentPost.Parent,
			&currentPost.Thread,
			&currentPost.Path,
		)
		if err != nil {
			return nil, err
		}

		currentPost.Created = strfmt.DateTime(created.UTC()).String()
		if !currentPost.Parent.Valid {
			currentPost.Parent.Int64 = 0
			currentPost.Parent.Valid = true
		}
		resultPosts = append(resultPosts, currentPost)
	}

	return resultPosts, nil
}

func (p *postgresAppRepository) UpdateThread(thread models.Thread) (models.Thread, error) {
	query := `UPDATE thread SET title=COALESCE(NULLIF($1, ''), title), message=COALESCE(NULLIF($2, ''), message) WHERE %s RETURNING *`

	var row *pgx.Row
	if thread.Slug == "" {
		query = fmt.Sprintf(query, `id=$3`)
		row = p.Conn.QueryRow(query, thread.Title, thread.Message, thread.Id)
	} else {
		query = fmt.Sprintf(query, `slug=$3`)
		row = p.Conn.QueryRow(query, thread.Title, thread.Message, thread.Slug)
	}

	var newThread models.Thread
	var created time.Time
	err := row.Scan(
		&newThread.Id,
		&newThread.Author,
		&created,
		&newThread.Forum,
		&newThread.Message,
		&newThread.Slug,
		&newThread.Title,
		&newThread.Votes,
	)

	if err != nil {
		return models.Thread{}, err
	}

	newThread.Created = strfmt.DateTime(created.UTC()).String()

	return newThread, nil
}

func (p *postgresAppRepository) InsertVote(vote models.Vote) (models.Vote, error) {
	_, err := p.Conn.Exec(
		`INSERT INTO votes(nickname, voice, thread_id) VALUES ($1, $2, $3)`,
		vote.Nickname,
		vote.Voice,
		vote.IdThread,
	)

	return vote, err
}

func (p *postgresAppRepository) UpdateVote(vote models.Vote) (models.Vote, error) {
	_, err := p.Conn.Exec(
		`UPDATE votes SET voice=$1 WHERE thread_id=$2 AND nickname=$3`,
		vote.Voice,
		vote.IdThread,
		vote.Nickname,
	)

	return vote, err
}

func (p *postgresAppRepository) GetServiceStatus() (map[string]int, error) {
	info, err := p.Conn.Query(
		`SELECT * FROM (SELECT COUNT(*) FROM forum) as forumCount,
		(SELECT COUNT(*) FROM post) as postCount,
		(SELECT COUNT(*) FROM thread) as threadCount, 
		(SELECT COUNT(*) FROM users) as usersCount;`,
	)

	if err != nil {
		return nil, err
	}

	defer info.Close()

	if info.Next() {
		forumCount, postCount, threadCount, usersCount := 0, 0, 0, 0
		err := info.Scan(&forumCount, &postCount, &threadCount, &usersCount)
		if err != nil {
			return nil, err
		}

		return map[string]int{
			"forum":  forumCount,
			"post":   postCount,
			"thread": threadCount,
			"user":   usersCount,
		}, nil
	}

	return nil, errors.New("have not information")
}

func (p *postgresAppRepository) ClearDatabase() error {
	_, err := p.Conn.Exec(`TRUNCATE users, thread, forum, post, votes, users_forum;`)

	return err
}

func (p *postgresAppRepository) SelectUsersByForum(slugForum string, parameters models.QueryParameters) ([]models.User, error) {
	var query string
	if parameters.Desc {
		if parameters.Since != "" {
			//	query = fmt.Sprintf(`SELECT users.about, users.Email, users.FullName, users.Nickname FROM users
			//inner join users_forum uf on users.Nickname = uf.nickname
			//WHERE uf.slug =$1 AND uf.nickname < '%s'
			//ORDER BY users.Nickname DESC LIMIT NULLIF($2, 0)`, parameters.Since)
			query = fmt.Sprintf(
				`SELECT about, email, fullname, nickname 
				FROM users_forum WHERE slug=$1 AND nickname < '%s' 
				ORDER BY nickname DESC LIMIT NULLIF($2, 0)`,
				parameters.Since,
			)
		} else {
			//	query = `SELECT users.about, users.Email, users.FullName, users.Nickname FROM users
			//inner join users_forum uf on users.Nickname = uf.nickname
			//WHERE uf.slug =$1
			//ORDER BY users.Nickname DESC LIMIT NULLIF($2, 0)`
			query = `SELECT about, email, fullname, nickname 
				FROM users_forum WHERE slug=$1 
				ORDER BY nickname DESC LIMIT NULLIF($2, 0)`
		}
	} else {
		//query = fmt.Sprintf(`SELECT users.about, users.Email, users.FullName, users.Nickname FROM users
		//inner join users_forum uf on users.Nickname = uf.nickname
		//WHERE uf.slug =$1 AND uf.nickname > '%s'
		//ORDER BY users.Nickname LIMIT NULLIF($2, 0)`, parameters.Since)
		query = fmt.Sprintf(
			`SELECT about, email, fullname, nickname
			FROM users_forum WHERE slug=$1 AND nickname > '%s'
			ORDER BY nickname LIMIT NULLIF($2, 0)`,
			parameters.Since,
		)
	}
	var data []models.User
	row, err := p.Conn.Query(query, slugForum, parameters.Limit)

	if err != nil {
		return data, nil
	}

	defer row.Close()

	for row.Next() {

		var u models.User

		err = row.Scan(&u.About, &u.Email, &u.FullName, &u.Nickname)

		if err != nil {
			return data, err
		}

		data = append(data, u)
	}

	return data, err
}

func (p *postgresAppRepository) SelectThreadsByForum(slugForum string, parameters models.QueryParameters) ([]models.Thread, error) {
	var rows *pgx.Rows
	var err error
	if parameters.Since != "" {
		if parameters.Desc {
			rows, err = p.Conn.Query(
				`SELECT * FROM thread WHERE forum=$1 AND created <= $2 
				ORDER BY created DESC LIMIT NULLIF($3, 0)`,
				slugForum, parameters.Since, parameters.Limit)
		} else {
			rows, err = p.Conn.Query(
				`SELECT * FROM thread WHERE forum=$1 AND created >= $2 
				ORDER BY created ASC LIMIT NULLIF($3, 0)`,
				slugForum, parameters.Since, parameters.Limit)
		}
	} else {
		if parameters.Desc {
			rows, err = p.Conn.Query(
				`SELECT * FROM thread WHERE forum=$1
				ORDER BY created DESC LIMIT NULLIF($2, 0)`,
				slugForum, parameters.Limit)
		} else {
			rows, err = p.Conn.Query(
				`SELECT * FROM thread WHERE forum=$1
				ORDER BY created ASC LIMIT NULLIF($2, 0)`,
				slugForum, parameters.Limit)
		}
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var threads []models.Thread
	for rows.Next() {
		var thread models.Thread
		var created time.Time

		err := rows.Scan(
			&thread.Id,
			&thread.Author,
			&created,
			&thread.Forum,
			&thread.Message,
			&thread.Slug,
			&thread.Title,
			&thread.Votes,
		)
		if err != nil {
			return nil, err
		}

		thread.Created = strfmt.DateTime(created.UTC()).String()

		threads = append(threads, thread)
	}

	return threads, err
}

func (p *postgresAppRepository) SelectPostById(id int) (models.Post, error) {
	var post models.Post
	var created time.Time

	err := p.Conn.QueryRow(
		`SELECT * FROM post WHERE id=$1 LIMIT 1;`,
		id).Scan(
		&post.Id,
		&post.Author,
		&created,
		&post.Forum,
		&post.Message,
		&post.IsEdited,
		&post.Parent,
		&post.Thread,
		&post.Path,
	)
	if err != nil {
		return models.Post{}, err
	}

	post.Created = strfmt.DateTime(created.UTC()).String()

	return post, nil
}

func (p *postgresAppRepository) UpdatePost(id int, message string) (models.Post, error) {
	var post models.Post
	var created time.Time
	err := p.Conn.QueryRow(
		`UPDATE post SET message=COALESCE(NULLIF($1, ''), message),
							 isEdited = CASE WHEN $1 = '' OR message = $1 THEN isEdited ELSE true END
							 WHERE id=$2 RETURNING *`,
		message,
		id,
	).Scan(
		&post.Id,
		&post.Author,
		&created,
		&post.Forum,
		&post.Message,
		&post.IsEdited,
		&post.Parent,
		&post.Thread,
		&post.Path,
	)

	post.Created = strfmt.DateTime(created.UTC()).String()

	return post, err
}

func (p *postgresAppRepository) selectThreadIdBySlug(slug string) (int, error) {
	row := p.Conn.QueryRow(`SELECT id FROM thread WHERE slug=$1 LIMIT 1;`, slug)

	var id int
	err := row.Scan(&id)

	return id, err
}

func (p *postgresAppRepository) SelectPostsByThread(thread models.Thread, limit, since int, sort string, desc bool) ([]models.Post, error) {
	var threadId int
	if thread.Id == 0 {
		thr, err := p.SelectThreadIdBySlug(thread.Slug)
		if err != nil {
			return nil, err
		}

		threadId = thr
	} else {
		threadId = thread.Id
	}

	switch sort {
	case "flat":
		posts, err := p.selectPostsByThreadFlat(threadId, limit, since, desc)

		return posts, err
	case "tree":
		posts, err := p.selectPostsByThreadTree(threadId, limit, since, desc)

		return posts, err
	case "parent_tree":
		posts, err := p.selectPostsByThreadParentTree(threadId, limit, since, desc)

		return posts, err
	default:
		return nil, errors.New("u gay")
	}
}

func (p *postgresAppRepository) selectPostsByThreadFlat(id, limit, since int, desc bool) ([]models.Post, error) {
	var rows *pgx.Rows
	var err error
	if since == 0 {
		if desc {
			rows, err = p.Conn.Query(`SELECT * FROM post WHERE thread=$1 ORDER BY id DESC LIMIT NULLIF($2, 0)`, id, limit)
		} else {
			rows, err = p.Conn.Query(`SELECT * FROM post WHERE thread=$1 ORDER BY id ASC LIMIT NULLIF($2, 0)`, id, limit)
		}
	} else {
		if desc {
			rows, err = p.Conn.Query(`SELECT * FROM post WHERE thread=$1 AND id < $2 ORDER BY id DESC LIMIT NULLIF($3, 0)`, id, since, limit)
		} else {
			rows, err = p.Conn.Query(`SELECT * FROM post WHERE thread=$1 AND id > $2 ORDER BY id ASC LIMIT NULLIF($3, 0)`, id, since, limit)
		}
	}
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var created time.Time

		err = rows.Scan(
			&post.Id,
			&post.Author,
			&created,
			&post.Forum,
			&post.Message,
			&post.IsEdited,
			&post.Parent,
			&post.Thread,
			&post.Path,
		)
		if err != nil {
			return nil, err
		}

		post.Created = strfmt.DateTime(created.UTC()).String()

		posts = append(posts, post)
	}

	return posts, err
}

func (p *postgresAppRepository) selectPostsByThreadTree(id, limit, since int, desc bool) ([]models.Post, error) {
	var rows *pgx.Rows
	var err error

	if since == 0 {
		if desc {
			rows, err = p.Conn.Query(
				`SELECT * FROM post
				WHERE thread=$1 ORDER BY path DESC, id  DESC LIMIT $2;`,
				id, limit,
			)
		} else {
			rows, err = p.Conn.Query(
				`SELECT * FROM post
				WHERE thread=$1 ORDER BY path ASC, id  ASC LIMIT $2;`,
				id, limit,
			)
		}
	} else {
		if desc {
			rows, err = p.Conn.Query(
				`SELECT * FROM post
				WHERE thread=$1 AND PATH < (SELECT path FROM post WHERE id = $2)
				ORDER BY path DESC, id  DESC LIMIT $3;`,
				id, since, limit,
			)
		} else {
			rows, err = p.Conn.Query(
				`SELECT * FROM post
				WHERE thread=$1 AND PATH > (SELECT path FROM post WHERE id = $2)
				ORDER BY path ASC, id  ASC LIMIT $3;`,
				id, since, limit,
			)
		}
	}
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var created time.Time

		err = rows.Scan(
			&post.Id,
			&post.Author,
			&created,
			&post.Forum,
			&post.Message,
			&post.IsEdited,
			&post.Parent,
			&post.Thread,
			&post.Path,
		)
		if err != nil {
			return nil, err
		}

		post.Created = strfmt.DateTime(created.UTC()).String()

		posts = append(posts, post)
	}

	return posts, nil
}

func (p *postgresAppRepository) selectPostsByThreadParentTree(id, limit, since int, desc bool) ([]models.Post, error) {
	var rows *pgx.Rows
	var err error

	if since == 0 {
		if desc {
			rows, err = p.Conn.Query(
				`SELECT * FROM post
				WHERE path[1] IN (SELECT id FROM post WHERE thread = $1 AND parent IS NULL ORDER BY id DESC LIMIT $2)
				ORDER BY path[1] DESC, path, id;`,
				id, limit,
			)
		} else {
			rows, err = p.Conn.Query(
				`SELECT * FROM post
				WHERE path[1] IN (SELECT id FROM post WHERE thread = $1 AND parent IS NULL ORDER BY id LIMIT $2)
				ORDER BY path, id;`,
				id, limit,
			)
		}
	} else {
		if desc {
			rows, err = p.Conn.Query(
				`SELECT * FROM post
				WHERE path[1] IN (SELECT id FROM post WHERE thread = $1 AND parent IS NULL AND PATH[1] <
				(SELECT path[1] FROM post WHERE id = $2) ORDER BY id DESC LIMIT $3) ORDER BY path[1] DESC, path, id;`,
				id, since, limit,
			)
		} else {
			rows, err = p.Conn.Query(`SELECT * FROM post
				WHERE path[1] IN (SELECT id FROM post WHERE thread = $1 AND parent IS NULL AND PATH[1] >
				(SELECT path[1] FROM post WHERE id = $2) ORDER BY id ASC LIMIT $3) ORDER BY path, id;`,
				id, since, limit,
			)
		}
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var created time.Time

		err = rows.Scan(
			&post.Id,
			&post.Author,
			&created,
			&post.Forum,
			&post.Message,
			&post.IsEdited,
			&post.Parent,
			&post.Thread,
			&post.Path,
		)
		if err != nil {
			return nil, err
		}

		post.Created = strfmt.DateTime(created.UTC()).String()

		posts = append(posts, post)
	}

	return posts, nil
}

func (p *postgresAppRepository) SelectThreadByForum(forum string) (models.Thread, error) {
	row := p.Conn.QueryRow(`SELECT * FROM thread WHERE forum=$1 LIMIT 1;`, forum)

	var thread models.Thread
	var created time.Time
	err := row.Scan(&thread.Id, &thread.Author, &created, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)

	thread.Created = strfmt.DateTime(created.UTC()).String()

	return thread, err
}

func (p *postgresAppRepository) SelectThreadIdBySlug(slug string) (int, error) {
	query := `SELECT id FROM thread WHERE slug=$1 LIMIT 1`

	var id int
	err := p.Conn.QueryRow(query, slug).Scan(&id)
	return id, err
}
