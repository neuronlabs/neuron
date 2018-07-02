package jsonapirepo

import (
	"github.com/kucjac/jsonapi"
	"net/http"
)

type JSONAPIRepository struct {
	URL        string
	HTTPClient *http.Client
}

func New(URL string, Client *http.Client) *JSONAPIRepository {
	return &JSONAPIRepository{URL: URL, HTTPClient: Client}
}

func (j *JSONAPIRepository) Create(scope *jsonapi.Scope) error {
	return nil
}

func (j *JSONAPIRepository) Get(scope *jsonapi.Scope) error {
	return nil
}

func (j *JSONAPIRepository) List(scope *jsonapi.Scope) error {
	return nil
}

func (j *JSONAPIRepository) Patch(scope *jsonapi.Scope) error {
	return nil
}

func (j *JSONAPIRepository) Delete(scope *jsonapi.Scope) error {
	return nil
}
