package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/mdobak/go-xerrors"
	"net/http"
	"realworld/internal/utils/log"
)

const sessionTableName = "session"

type SessionRepo struct {
	Db *sql.DB
}

type Session struct {
	SessionId string
	UserId    uint
}

type UserInterface interface {
	GetId() uint
	SetToken(string)
}

func (sessionRepo *SessionRepo) Check(r *http.Request) (*Session, error) {
	token := r.Header.Get("Authorization")
	id, err := sessionRepo.GetUserIdByToken(token)
	if err != nil {
		return nil, err
	}
	session := Session{SessionId: token, UserId: id}
	return &session, nil
}

func (sessionRepo *SessionRepo) Create(w http.ResponseWriter, u UserInterface) error {
	token := CreateToken()
	query := "INSERT INTO session (session_token, user_id) VALUES ($1, $2)"
	_, err := sessionRepo.Db.Exec(query, token, u.GetId())
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, u.GetId())
		return err
	}
	u.SetToken(token)
	newUserJson, _ := json.Marshal(map[string]UserInterface{"user": u})
	w.WriteHeader(201)
	log.ErrWriteResp(w.Write(newUserJson))
	return nil
}

func (sessionRepo *SessionRepo) UpdateToken(userId uint) (string, error) {
	token := CreateToken()
	query := fmt.Sprintf("UPDATE %s SET session_token='%s' WHERE user_id = %d", sessionTableName, token, userId)
	_, err := sessionRepo.Db.Exec(query)
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, userId)
		return "", err
	}
	return token, nil

}

func (sessionRepo *SessionRepo) GetUserIdByToken(token string) (uint, error) {
	query := fmt.Sprintf("SELECT user_id FROM %s WHERE session_token = $1", sessionTableName)
	row := sessionRepo.Db.QueryRow(query, token)
	var userId int
	if err := row.Scan(&userId); err != nil {
		if err == sql.ErrNoRows {
			log.DbQueryCtx(context.Background(), xerrors.New(err), query, token)
			return 0, err
		}
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, token)
		return 0, err
	}

	return uint(userId), nil
}

func (sessionRepo *SessionRepo) GetTokenByUserId(userId uint) (string, error) {
	query := fmt.Sprintf("SELECT session_token FROM %s WHERE user_id = $1", sessionTableName)
	row := sessionRepo.Db.QueryRow(query, userId)
	var token string
	if err := row.Scan(&token); err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, userId)
		return "", err
	}

	return token, nil
}

func (sessionRepo *SessionRepo) DestroyCurrent(userId uint) error {
	query := "DELETE FROM session WHERE user_id = $1"
	_, err := sessionRepo.Db.Exec(query, userId)
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, userId)
		return err
	}
	return nil
}
