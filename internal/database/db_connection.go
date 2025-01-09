package database

import (
	"database/sql"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/sohWenMing/aggregator/internal/config"
)

type State struct {
	Db     *Queries
	Cfg    *config.Config
	Client *http.Client
}

func CreateDBConnection() (state *State, err error) {
	config, err := config.Read()
	if err != nil {
		return nil, err
	}
	dbURL := config.DbUrl
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	queries := New(db)

	newState := State{
		Db:     queries,
		Cfg:    config,
		Client: &http.Client{},
	}
	return &newState, nil

}
