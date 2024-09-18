package article

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"realworld/internal/session"
	"realworld/internal/utils/log"
)

type ArticleHandler struct {
	ArticleRepo ArticleManagerInterface
}

type ArticleReq struct {
	Article Article `json:"article"`
}

type ArticleResp struct {
	Article Article `json:"article"`
}

type Articles struct {
	Articles      []Article `json:"articles"`
	ArticlesCount int       `json:"articlesCount"`
}

func (articleHander *ArticleHandler) HandleGetArticles(w http.ResponseWriter, r *http.Request) {
	parsedUrl, _ := url.Parse(r.URL.String())
	params, _ := url.ParseQuery(parsedUrl.RawQuery)
	authorArr := params["author"]
	tagArr := params["tag"]
	if len(authorArr) != 0 {
		author := authorArr[0]
		articleHander.articlesByAuthor(w, r, author)
	} else if len(tagArr) != 0 {
		tag := tagArr[0]
		articleHander.articlesByTag(w, r, tag)
	} else {
		articleHander.allArticles(w, r)

	}
}

func (articleHander *ArticleHandler) articlesByTag(w http.ResponseWriter, r *http.Request, tag string) {
	articles, err := articleHander.ArticleRepo.ArticlesByTag(tag)
	if err != nil {
		http.Error(w, "error getting articles", http.StatusInternalServerError)
	}

	respBody, _ := json.Marshal(Articles{Articles: articles, ArticlesCount: len(articles)})
	log.ErrWriteResp(w.Write(respBody))
}

func (articleHander *ArticleHandler) articlesByAuthor(w http.ResponseWriter, r *http.Request, author string) {
	articles, err := articleHander.ArticleRepo.ArticlesByAuthor(author)
	if err != nil {
		http.Error(w, "error getting articles", http.StatusInternalServerError)
	}

	respBody, _ := json.Marshal(Articles{Articles: articles, ArticlesCount: len(articles)})
	log.ErrWriteResp(w.Write(respBody))

}

func (articleHandler *ArticleHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	userSession, ok := session.GetFromCtx(r)
	if !ok {
		http.Error(w, "error", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.ErrReadBody(err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	var articleReq ArticleReq
	err = json.Unmarshal(body, &articleReq)
	if err != nil {
		log.UnmarshalBodyErr(err)
		http.Error(w, "err", http.StatusBadRequest)
		return
	}
	article := articleReq.Article
	article.UserID = userSession.UserId
	createdArticle, err := articleHandler.ArticleRepo.CreateArticle(article)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	respBody, _ := json.Marshal(ArticleResp{*createdArticle})
	log.ErrWriteResp(w.Write(respBody))
}

func (articleHandler *ArticleHandler) allArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := articleHandler.ArticleRepo.AllArticles()
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
	}
	respBody, _ := json.Marshal(Articles{Articles: articles, ArticlesCount: len(articles)})
	log.ErrWriteResp(w.Write(respBody))
}
