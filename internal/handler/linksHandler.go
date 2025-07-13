package handler

import (
	"io"
	"net/http"

	"github.com/Alexey-zaliznuak/shortener/internal/model"
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

	http.Redirect(res, req, fullUrl, http.StatusTemporaryRedirect)
}

func createLink(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	link := &model.Link{FullUrl: string(body)}
	err = linksService.CreateLink(link)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Write([]byte(link.ShortUrl))
}

func defaultHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		createLink(res, req)
	}

	if req.Method == http.MethodGet {
		redirect(res, req)
	}
}

func init() {
	Mux.HandleFunc("/", defaultHandler)
}
