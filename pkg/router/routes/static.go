package routes

import (
	iofs "io/fs"
	"net/http"
)

func Static(staticFS iofs.FS) http.HandlerFunc {
	return http.StripPrefix("/static", http.FileServer(http.FS(staticFS))).ServeHTTP
}
