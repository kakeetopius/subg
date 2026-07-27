package subdl

import "github.com/kakeetopius/subg/internal/sessions"

type SubDL struct {
	session sessions.Session
}

type SDSubtitle struct {
	Name         string
	SubID        string
	Author       string
	DownloadLink string
}

type SubdlSearchResults struct {
	Results []SubdlSearchResult `json:"results"`
}

type SubdlSearchResult struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Year int    `json:"year"`
	Link string `json:"link"`
}

type SubDlSearchParams struct {
	Query string `url:"query"`
}
