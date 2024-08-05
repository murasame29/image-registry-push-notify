package routes

import (
	"fmt"
	"net/http"

	"github.com/murasame29/image-registry-push-notify/sample-app/internal/handler"
)

func NewRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "status ok")
	})

	mux.HandleFunc("/kustomization", handler.HandleRequest)

	return mux
}
