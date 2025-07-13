package handler

import (
	"net/http"

	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
)

var Mux = http.NewServeMux()
var linksService = service.NewLinksService(database.Client)

func redirect(res http.ResponseWriter, req *http.Request) {
	shortUrl := req.URL.Path[1:]
	fullUrl, err := linksService.GetFullUrlFromShort(shortUrl)

	if err != nil || fullUrl == "" {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(res, req, fullUrl, http.StatusMovedPermanently)
}

func init() {
	Mux.HandleFunc(`/`, redirect)
}
