package paginate

import (
	"net/http"
	"strconv"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
)

type SimplePaginatedActionFn[T any] = func(offset, limit int) ([]T, error)

type SimplePaginatedAction[T any] struct {
	Action SimplePaginatedActionFn[T]
}

func NewSimplePaginatedAction[T any](action SimplePaginatedActionFn[T]) *SimplePaginatedAction[T] {
	return &SimplePaginatedAction[T]{Action: action}
}

func (a *SimplePaginatedAction[T]) Exec(r *http.Request) (*PaginatedData[T], error) {
	paginatedData := &PaginatedData[T]{}

	page, perPage := a.validatePaginationParam(r)

	data, err := a.Action(page*perPage, perPage+1)
	if err != nil {
		return paginatedData, err
	}

	if len(data) == perPage+1 {
		paginatedData.Data = data[:perPage]
		paginatedData.setNext(r.URL.Path, r.URL.Query(), page+1, perPage)
	} else {
		paginatedData.Data = data
	}
	paginatedData.setPrev(r.URL.Path, r.URL.Query(), page-1, perPage)

	return paginatedData, nil
}

func (a *SimplePaginatedAction[T]) validatePaginationParam(r *http.Request) (page, perPage int) {
	convToInt := func(strNum string) int {
		n, err := strconv.Atoi(strNum)
		if err != nil {
			return 0
		}
		return n
	}
	page = convToInt(r.FormValue("page"))
	perPage = utils.Clamp(convToInt(r.FormValue("per_page")), 1, 40)
	return page, perPage
}
