package app

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"realworld/internal/article"
	"realworld/internal/session"
	"realworld/internal/user"
	"realworld/internal/utils"
	"realworld/internal/utils/log"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"pass"`
		DbName   string `yaml:"db_name"`
	} `yaml:"database"`
}

func GetApp() http.Handler {
	logger := log.GetLogger()
	slog.SetDefault(logger)

	cfg := parseConfig()
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DbName)

	db, err := sql.Open("postgres", psqlconn)
	PanicIfErr(err)
	err = db.Ping()
	PanicIfErr(err)

	colsNames, err := utils.GetDbTablesCols(db)
	PanicIfErr(err)

	userRepo := user.UserRepo{Db: db}
	sessionRepo := session.SessionRepo{Db: db}
	userHandler := user.UserHandler{UserStorage: &userRepo, SessionManager: sessionRepo, ListOfCols: colsNames}

	articleRepo := article.ArticlesRepo{Db: db}
	articleHandler := article.ArticleHandler{ArticleRepo: &articleRepo}

	r := mux.NewRouter()
	r.HandleFunc("/api/users", userHandler.Registration).Methods(http.MethodPost)
	r.HandleFunc("/api/users/login", userHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/api/articles", articleHandler.HandleGetArticles).Methods(http.MethodGet)

	api := r.PathPrefix("/api/").Subrouter()
	api.HandleFunc("/user", userHandler.CurrUser).Methods(http.MethodGet)
	api.HandleFunc("/user/logout", userHandler.Logout).Methods(http.MethodPost)
	api.HandleFunc("/user", userHandler.UpdateUser).Methods(http.MethodPut)
	api.HandleFunc("/articles", articleHandler.CreateArticle).Methods(http.MethodPost)
	http.Handle("/", r)
	api.Use(userHandler.AuthMiddleware)
	return r
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func parseConfig() Config {
	rootPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	filePath := filepath.Join(rootPath, "configs", "config.yaml")
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
