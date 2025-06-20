package paginate

import (
	"fmt"
	"maps"
	"net/url"
	"strconv"
)

type PaginatedData[T any] struct {
	Data []T    `json:"data"`
	Next string `json:"next,omitzero"`
	Prev string `json:"prev,omitzero"`
}

func (d *PaginatedData[T]) setNext(path string, requestQueryParam url.Values, page, perPage int) {
	param := maps.Clone(requestQueryParam)
	param.Set("page", strconv.Itoa(page))
	param.Set("per_page", strconv.Itoa(perPage))
	d.Next = fmt.Sprintf("%s?%s", path, param.Encode())
}

func (d *PaginatedData[T]) setPrev(path string, requestQueryParam url.Values, page, perPage int) {
	if page <= 0 {
		return
	}
	param := maps.Clone(requestQueryParam)
	param.Set("page", strconv.Itoa(page))
	param.Set("per_page", strconv.Itoa(perPage))
	d.Prev = fmt.Sprintf("%s?%s", path, param.Encode())
}

func PaginatedDataMapper[T, R any](d *PaginatedData[T], mapper func(e T) R) *PaginatedData[R] {
	paginatedData := &PaginatedData[R]{}
	paginatedData.Next = d.Next
	paginatedData.Prev = d.Prev

	data := make([]R, len(d.Data))
	for i, e := range d.Data {
		data[i] = mapper(e)
	}
	paginatedData.Data = data

	return paginatedData
}
