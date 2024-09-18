package user

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"realworld/internal/session"
	"realworld/internal/utils"
	"realworld/internal/utils/log"
)

type UserHandler struct {
	UserStorage    UserStorageInterface
	SessionManager session.SessionRepo
	ListOfCols     map[string]utils.ListOfColumns
}

type UserJson struct {
	User User `json:"user"`
}

type UpdateUserBody struct {
	User map[string]interface{} `json:"user"`
}

func (userHandler *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
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

	var updateUserBody UpdateUserBody
	err = json.Unmarshal(body, &updateUserBody)
	if err != nil {
		log.UnmarshalBodyErr(err)
		http.Error(w, "error", http.StatusBadRequest)
		return
	}
	valuesMap := updateUserBody.User
	fieldsToUpdate := make([]string, len(valuesMap))
	newValues := make([]interface{}, len(valuesMap))
	i := 0
	for k, v := range valuesMap {
		if _, ok := userHandler.ListOfCols[userTableName][k]; !ok {
			errMsg := fmt.Sprintf("no such field for user: [%s]", k)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		fieldsToUpdate[i] = k
		newValues[i] = v
		i++
	}
	if len(fieldsToUpdate) == 0 {
		errMsg := fmt.Sprintf("no such field for user:")
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	err = userHandler.UserStorage.UpdateUser(userSession.UserId, fieldsToUpdate, newValues)
	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}
	user, err := userHandler.UserStorage.GetUserById(userSession.UserId)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	token, err := userHandler.SessionManager.UpdateToken(user.Id)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	user.Token = token
	resp, _ := json.Marshal(UserJson{User: User{Email: user.Email, CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt, Username: user.Username, Bio: user.Bio, Token: user.Token}})
	log.ErrWriteResp(w.Write(resp))
}

func (userHandler *UserHandler) CurrUser(w http.ResponseWriter, r *http.Request) {
	userSession, ok := session.GetFromCtx(r)
	if !ok {
		http.Error(w, "error", http.StatusUnauthorized)
		return
	}
	user, err := userHandler.UserStorage.GetUserById(userSession.UserId)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	respBody, _ := json.Marshal(UserJson{User: User{Email: user.Email, CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt, Username: user.Username, Bio: user.Bio}})
	log.ErrWriteResp(w.Write(respBody))
}
