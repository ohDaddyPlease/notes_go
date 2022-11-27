package metric

import (
	"github.com/julienschmidt/httprouter"
	"gitlab.konstweb.ru/ow/arch/notes/note_service/pkg/logging"
	"net/http"
)

const (
	URL = "/api/heartbeat"
)

type Handler struct {
	Logger logging.Logger
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, URL, h.Heartbeat)
}

func (h *Handler) Heartbeat(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(204)
}
