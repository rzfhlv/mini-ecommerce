package delivery

import (
	"net/http"

	"mini-ecommerce/internal/catalog/usecase"
)

type CategoryHandler struct{ listUC usecase.ListCategoriesUseCase }

func NewCategoryHandler(listUC usecase.ListCategoriesUseCase) *CategoryHandler { return &CategoryHandler{listUC: listUC} }

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.listUC.Execute(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Message: "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: categories})
}
