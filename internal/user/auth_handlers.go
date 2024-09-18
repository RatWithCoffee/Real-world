package user

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"realworld/internal/session"
	log2 "realworld/internal/utils/log"
)

func (userHandler *UserHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userSession, err := userHandler.SessionManager.Check(r)
		if err != nil {
			http.Error(w, "Forbidden", http.StatusUnauthorized)
		} else {
			r = r.WithContext(session.SaveToCtx(userSession))
			next.ServeHTTP(w, r)
		}
	})
}

func (userHandler *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userSession, ok := session.GetFromCtx(r)
	if !ok {
		http.Error(w, "error", http.StatusUnauthorized)
		return
	}
	err := userHandler.SessionManager.DestroyCurrent(userSession.UserId)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
}

func (userHandler *UserHandler) Registration(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log2.ErrReadBody(err)
		http.Error(w, "can't read body", http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	var registrationBody UserJson
	err = json.Unmarshal(body, &registrationBody)
	if err != nil {
		log.Printf("[%s] Error unmarshaling body: %v", "RegistrationHandler", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	user := registrationBody.User
	newUser, err := userHandler.UserStorage.Create(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = userHandler.SessionManager.Create(w, newUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (userHandler *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log2.ErrReadBody(err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	loginBody := UserJson{}
	err = json.Unmarshal(body, &loginBody)
	if err != nil {
		log2.UnmarshalBodyErr(err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	respUser := loginBody.User

	user, err := userHandler.UserStorage.GetUserByEmail(respUser.Email)
	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	encrPassword := []byte(user.Password)
	salt := user.Salt
	isValid := IsValidPassword([]byte(respUser.Password), salt, encrPassword)
	if isValid {
		sendUserInfo(userHandler, w, user)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

}

func sendUserInfo(userHandler *UserHandler, w http.ResponseWriter, u *User) {
	user, err := userHandler.UserStorage.GetUserByEmail(u.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	token, err := userHandler.SessionManager.GetTokenByUserId(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	respBody, err := json.Marshal(UserJson{User: User{Email: user.Email, CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt, Username: user.Username, Token: token}})
	if err != nil {
		log.Printf("[%s]: err - %v, value - %v", "Login", err, user)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	log2.ErrWriteResp(w.Write(respBody))

}
