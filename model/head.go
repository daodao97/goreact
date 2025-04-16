package model

type Head struct {
	Title string
	Meta  []Meta
	Link  []Link
}

type Meta struct {
	Name     string
	Property string
	Content  string
}

type Link struct {
	Href string
	Rel  string
}
