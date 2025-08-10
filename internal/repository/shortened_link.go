package repository

import (
	"github.com/Alexey-zaliznuak/shortener/internal/model"
)

type LinkRepository struct{}

func (r *LinkRepository) Create(link *model.Link) {
	model.LinksStorage[link.Shortcut] = link
}

func (r *LinkRepository) GetByShortcut(shortcut string) (*model.Link, bool) {
	l, ok := model.LinksStorage[shortcut]
	return l, ok
}
