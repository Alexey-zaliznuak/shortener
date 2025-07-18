package model

var LinksStorage = make(map[string]*Link)

type Link struct {
	FullURL  string
	Shortcut string
}
