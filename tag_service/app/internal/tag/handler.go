package tag

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gitlab.konstweb.ru/ow/arch/notes/tag_service/internal/apperror"
	"gitlab.konstweb.ru/ow/arch/notes/tag_service/pkg/logging"
	"net/http"
	"strconv"
	"strings"
)

const (
	tagsURL = "/api/tags"
	tagURL  = "/api/tags/:id"
)

type Handler struct {
	Logger     logging.Logger
	TagService Service
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, tagURL, apperror.Middleware(h.GetTag))
	router.HandlerFunc(http.MethodGet, tagsURL, apperror.Middleware(h.GetTags))
	router.HandlerFunc(http.MethodPost, tagsURL, apperror.Middleware(h.CreateTag))
	router.HandlerFunc(http.MethodPatch, tagURL, apperror.Middleware(h.PartiallyUpdateTag))
	router.HandlerFunc(http.MethodDelete, tagURL, apperror.Middleware(h.DeleteTag))
}

func (h *Handler) GetTag(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	tagIDStr := params.ByName("id")
	id, err := strconv.Atoi(tagIDStr)
	if err != nil {
		return apperror.BadRequestError("id resource identifier is required and must be an integer")
	}

	tag, err := h.TagService.GetOne(r.Context(), id)
	if err != nil {
		return err
	}

	tagsBytes, err := json.Marshal(tag)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write(tagsBytes)

	return nil
}

func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	idsParam := r.URL.Query().Get("id")
	if idsParam == "" {
		return apperror.BadRequestError("id query parameter is required and must be a comma separated integers")
	}

	var tagsIds []int
	idsStr := strings.Split(idsParam, ",")
	for _, idStr := range idsStr {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return apperror.BadRequestError("id query parameter is required and must be a comma separated integers")
		}
		tagsIds = append(tagsIds, id)
	}

	tags, err := h.TagService.GetMany(r.Context(), tagsIds)
	if err != nil {
		return err
	}

	tagsBytes, err := json.Marshal(tags)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write(tagsBytes)

	return nil
}

func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	var dto CreateTagDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return apperror.BadRequestError("invalid JSON scheme")
	}

	tagID, err := h.TagService.Create(r.Context(), dto)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("%s/%d", tagsURL, tagID))
	w.WriteHeader(http.StatusCreated)

	return nil
}

func (h *Handler) PartiallyUpdateTag(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	paramID := params.ByName("id")

	tagID, err := strconv.Atoi(paramID)
	if err != nil {
		return apperror.BadRequestError("id query parameter is required and must be a comma separated integers")
	}

	var dto UpdateTagDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return apperror.BadRequestError("invalid JSON scheme")
	}

	dto.ID = tagID

	err = h.TagService.Update(r.Context(), dto)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *Handler) DeleteTag(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	paramID := params.ByName("id")

	tagID, err := strconv.Atoi(paramID)
	if err != nil {
		return apperror.BadRequestError("id query parameter is required and must be a comma separated integers")
	}

	err = h.TagService.Delete(r.Context(), tagID)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
}
