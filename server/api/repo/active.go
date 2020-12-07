package repo

import (
	"net/http"
	"strconv"

	"github.com/bleenco/abstruse/pkg/lib"
	"github.com/bleenco/abstruse/server/api/render"
	"github.com/bleenco/abstruse/server/core"
	"github.com/go-chi/chi"
)

// HandleActive returns an http.HandlerFunc that writes JSON encoded
// result about saving active status to the http response body.
func HandleActive(repos core.RepositoryStore) http.HandlerFunc {
	type form struct {
		Active bool `json:"active"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var f form
		var err error
		defer r.Body.Close()

		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			render.InternalServerError(w, err.Error())
			return
		}

		if err = lib.DecodeJSON(r.Body, &f); err != nil {
			render.InternalServerError(w, err.Error())
			return
		}

		if err = repos.SetActive(uint(id), f.Active); err != nil {
			render.NotFoundError(w, err.Error())
			return
		}

		render.JSON(w, http.StatusOK, render.Empty{})
	}
}