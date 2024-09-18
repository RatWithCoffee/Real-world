package session

import (
	"context"
	"net/http"
)

const SESSION_CTX_KEY = "session_ctx_key"

func GetFromCtx(r *http.Request) (*Session, bool) {
	session, ok := r.Context().Value(SESSION_CTX_KEY).(*Session)
	return session, ok
}

func SaveToCtx(session *Session) context.Context {
	return context.WithValue(context.Background(), SESSION_CTX_KEY, session)
}
