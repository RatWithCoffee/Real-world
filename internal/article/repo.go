package article

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gosimple/slug"
	"github.com/mdobak/go-xerrors"
	"realworld/internal/utils"
	"realworld/internal/utils/log"
	"strconv"
	"strings"
	"time"
)

type Article struct {
	ID          uint     `json:"-"`
	Slug        string   `json:"slug"`
	UserID      uint     `json:"-"`
	Body        string   `json:"body"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	Tags        []string `json:"tagList"`
	Author      Author   `json:"author"`
}

type Author struct {
	Username string                  `json:"username"`
	Bio      utils.NullStringWrapper `json:"bio"`
}

type ArticleManagerInterface interface {
	CreateArticle(article Article) (*Article, error)
	AllArticles() ([]Article, error)
	ArticlesByAuthor(author string) ([]Article, error)
	ArticlesByTag(tag string) ([]Article, error)
}

type ArticlesRepo struct {
	Db *sql.DB
}

func (articleRepo *ArticlesRepo) ArticlesByTag(tag string) ([]Article, error) {
	query := "SELECT a.id, a.slug, a.body, a.title, a.description, a.created_at, a.updated_at, u.username, u.bio" +
		" FROM article AS a JOIN user_data AS u ON a.user_id = u.id " +
		"WHERE a.id IN (SELECT article_id FROM article_tags JOIN tag ON id = tag_id WHERE text = $1)"
	rows, err := articleRepo.Db.Query(query, tag)
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, tag)
		return nil, err
	}
	defer rows.Close()
	article := Article{}
	articles := make([]Article, 0)
	for rows.Next() {
		if err := rows.Scan(&article.ID, &article.Slug, &article.Body, &article.Title, &article.Description, &article.CreatedAt, &article.UpdatedAt, &article.Author.Username, &article.Author.Bio); err != nil {
			log.DbQueryCtx(context.Background(), xerrors.New(err), query, tag)
			return nil, err
		}
		tags, err := articleRepo.getTags(article.ID)
		if err != nil {
			return nil, err
		}
		article.Tags = tags
		articles = append(articles, article)
	}
	return articles, nil
}

func (articleRepo *ArticlesRepo) ArticlesByAuthor(username string) ([]Article, error) {
	query := "SELECT  a.id, a.slug, a.body, a.title, a.description, a.created_at, a.updated_at, u.username, u.bio" +
		" FROM article AS a JOIN user_data AS u ON a.user_id = u.id WHERE u.username = $1"
	rows, err := articleRepo.Db.Query(query, username)
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, username)
		return nil, err
	}
	defer rows.Close()
	article := Article{}
	articles := make([]Article, 0)
	for rows.Next() {
		if err := rows.Scan(&article.ID, &article.Slug, &article.Body, &article.Title, &article.Description, &article.CreatedAt, &article.UpdatedAt, &article.Author.Username, &article.Author.Bio); err != nil {
			log.DbQueryCtx(context.Background(), xerrors.New(err), query, username)

			return nil, err
		}
		tags, err := articleRepo.getTags(article.ID)
		if err != nil {
			return nil, err
		}
		article.Tags = tags
		articles = append(articles, article)
	}
	return articles, nil
}

// TODO: add pagination
func (articleRepo *ArticlesRepo) AllArticles() ([]Article, error) {
	query := "SELECT a.id, a.slug, a.body, a.title, a.description, a.created_at, a.updated_at, u.username, u.bio" +
		" FROM article AS a JOIN user_data AS u ON a.user_id = u.id"
	rows, err := articleRepo.Db.Query(query)
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, "")
		return nil, err
	}
	defer rows.Close()
	article := Article{}
	articles := make([]Article, 0)
	for rows.Next() {
		if err := rows.Scan(&article.ID, &article.Slug, &article.Body, &article.Title, &article.Description, &article.CreatedAt, &article.UpdatedAt, &article.Author.Username, &article.Author.Bio); err != nil {
			log.DbQueryCtx(context.Background(), xerrors.New(err), query, "")
			return nil, err
		}
		tags, err := articleRepo.getTags(article.ID)
		if err != nil {
			return nil, err
		}
		article.Tags = tags
		articles = append(articles, article)
	}
	return articles, nil
}

func (articleRepo *ArticlesRepo) getTags(articleId uint) ([]string, error) {
	query := "SELECT text FROM tag JOIN article_tags ON tag.id = article_tags.tag_id WHERE article_id = $1"
	rows, err := articleRepo.Db.Query(query, articleId)
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, articleId)
		return nil, err
	}
	var tag string
	tags := make([]string, 0)
	for rows.Next() {
		if err := rows.Scan(&tag); err != nil {
			log.DbQueryCtx(context.Background(), xerrors.New(err), query, articleId)
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (articleRepo *ArticlesRepo) CreateArticle(article Article) (*Article, error) {
	query := "INSERT INTO article (slug, user_id, body, title, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	articleSlug, err := articleRepo.generateSlug(article.Title)
	if err != nil {
		return nil, err
	}
	article.Slug = articleSlug

	currTime := time.Now().Format(time.RFC3339)
	article.UpdatedAt = currTime
	article.CreatedAt = currTime
	insertValues := []interface{}{article.Slug, article.UserID, article.Body, article.Title, article.Description, article.CreatedAt, article.UpdatedAt}
	var lastInsertedId uint
	err = articleRepo.Db.QueryRow(query, insertValues...).Scan(&lastInsertedId)
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, insertValues)
		return nil, err
	}
	article.ID = lastInsertedId
	err = articleRepo.addTags(article)
	if err != nil {
		return nil, err
	}
	author, err := articleRepo.getAuthor(article)
	if err != nil {
		return nil, err
	}
	article.Author = *author
	return &article, nil
}

func (articleRepo *ArticlesRepo) getAuthor(article Article) (*Author, error) {
	query := "SELECT username, bio FROM user_data WHERE id = $1"
	var author Author
	err := articleRepo.Db.QueryRow(query, article.UserID).Scan(&author.Username, &author.Bio)
	if err != nil {
		return nil, err
	}
	return &author, nil
}

func (articleRepo *ArticlesRepo) addTags(article Article) error {
	if len(article.Tags) == 0 {
		return nil
	}
	var tagsBuilder strings.Builder
	for _, t := range article.Tags {
		tagsBuilder.Write([]byte(fmt.Sprintf("'%s', ", t)))

	}
	tags := tagsBuilder.String()
	tags = tags[:len(tags)-2]

	query := fmt.Sprintf("SELECT id FROM tag WHERE text IN (%s)", tags)
	rows, err := articleRepo.Db.Query(query)
	defer rows.Close()
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, tags)
		return err
	}
	var valuesBuilder strings.Builder
	var tagId int
	for rows.Next() {
		if err := rows.Scan(&tagId); err != nil {
			log.DbQueryCtx(context.Background(), xerrors.New(err), query, tags)
			return err
		}
		valuesBuilder.Write([]byte(fmt.Sprintf("(%d, %d),", article.ID, tagId)))
	}
	if valuesBuilder.Len() == 0 {
		return nil
	}
	values := valuesBuilder.String()
	values = values[:len(values)-1]
	query = fmt.Sprintf("INSERT INTO article_tags (article_id, tag_id) VALUES %s", values)
	_, err = articleRepo.Db.Exec(query)
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, values)
		return err
	}

	return nil

}

func (articleRepo *ArticlesRepo) generateSlug(title string) (string, error) {
	query := "SELECT COUNT(*) FROM article WHERE title = $1"
	row := articleRepo.Db.QueryRow(query, title)
	articleNum := 0
	err := row.Scan(&articleNum)
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, title)
		return "", err
	}

	artcileSlug := slug.Make(title)
	if articleNum != 0 {
		strId := strconv.Itoa(articleNum)
		return artcileSlug + "--" + strId, nil
	}
	return artcileSlug, nil
}
