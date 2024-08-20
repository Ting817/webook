package sqlx_store

import (
	"net/http"

	ginSession "github.com/gin-contrib/sessions"
	"github.com/gorilla/sessions"
)

type Store struct {
}

func (s Store) Get(r *http.Request, name string) (*sessions.Session, error) {
	// TODO implement me
	panic("implement me")
}

func (s Store) New(r *http.Request, name string) (*sessions.Session, error) {
	// TODO implement me
	panic("implement me")
}

func (s Store) Save(r *http.Request, w http.ResponseWriter, st *sessions.Session) error {
	// TODO implement me
	panic("implement me")
}

func (s Store) Options(options ginSession.Options) {
	// TODO implement me
	panic("implement me")
}
