package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/todo"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/paginate"
)

// this limits are also check on the db level,
// if you need to change the limits you need to change them here and on the db level.
// see the todo table migration file(s)
const (
	todoBodyLengthLimit   int = 10000
	todoTitleLengthLimit  int = 150
	todoStatusLengthLimit int = 50
)

func todoRouter(_ context.Context, s *Server) http.Handler {
	todoRepo := todo.NewRepository(s.db, s.rdb)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /todo", todoIndex(todoRepo))
	mux.HandleFunc("GET /todo/{id}", todoShow(todoRepo))

	mux.HandleFunc("POST /todo", createTodo(todoRepo))

	mux.HandleFunc("PATCH /todo/{id}", updateTodo(todoRepo))

	mux.HandleFunc("DELETE /todo/{id}", deleteTodo(todoRepo))

	return mux
}

func createTodo(todoRepo todo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		userAndSession := auth.MustUserAndSessionFromContext(ctx)

		todoData, err := extractTodoData(r)
		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		res, err := todoRepo.CreateTodo(ctx, int(userAndSession.UserID), todoData)
		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		writeResponse(ctx, w, r, http.StatusCreated, publicTodoItemFromRepoModel(res))
	}
}

func extractTodoData(r *http.Request) (todo.TodoData, error) {
	data := todo.TodoData{}

	var title string
	title = r.FormValue("title")
	titleLen := len(title)
	if titleLen > todoTitleLengthLimit {
		return todo.TodoData{}, errors.New("too large todo title")
	}
	if titleLen != 0 {
		data.Title = &title
	}

	var body string
	body = r.FormValue("body")
	bodyLen := len(body)
	if bodyLen > todoBodyLengthLimit {
		return todo.TodoData{}, errors.New("too large todo body")
	}
	if bodyLen != 0 {
		data.Body = &body
	}

	status := new(todo.TodoStatus)
	statusStr := r.FormValue("status")
	statusStrLen := len(statusStr)
	if statusStrLen > todoStatusLengthLimit {
		return todo.TodoData{}, errors.New("too large todo status")
	}
	if statusStrLen != 0 {
		status, err := status.FromString(statusStr)
		if err != nil {
			return todo.TodoData{}, err
		}
		data.Status = status
	}

	return data, nil
}

func updateTodo(todoRepo todo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		todoId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, errors.New("can not parse the todo id from the url"))
			return
		}

		err = r.ParseForm()
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		userAndSession := auth.MustUserAndSessionFromContext(ctx)

		todoData, err := extractTodoData(r)
		if err != nil {
			writeError(ctx, w, r, return400IfApp404IfNoResultErrOr500(err), err)
			return
		}

		res, err := todoRepo.UpdateTodo(ctx, int(userAndSession.UserID), todoId, todoData)
		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		writeResponse(ctx, w, r, http.StatusCreated, publicTodoItemFromRepoModel(res))
	}
}

func todoIndex(todoRepo todo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, err)
			return
		}

		userAndSession := auth.MustUserAndSessionFromContext(ctx)

		paginatedDate, err := paginate.NewSimplePaginatedAction(
			func(offset, limit int) ([]todo.TodoItem, error) {
				return todoRepo.GetTodos(
					ctx,
					int(userAndSession.UserID),
					offset,
					limit,
				)
			},
		).Exec(r)
		if err != nil && !errors.Is(err, apperr.ErrNoResult) {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		// convert items to public form
		publicPaginatedDate := paginate.PaginatedDataMapper(paginatedDate, publicTodoItemFromRepoModel)

		writeResponse(ctx, w, r, http.StatusOK, publicPaginatedDate)
	}
}

func todoShow(todoRepo todo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		todoId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, errors.New("can not parse the todo id from the url"))
			return
		}

		userAndSession := auth.MustUserAndSessionFromContext(ctx)

		res, err := todoRepo.GetTodo(ctx, int(userAndSession.UserID), todoId)
		if err != nil {
			writeError(ctx, w, r, return400IfApp404IfNoResultErrOr500(err), err)
			return
		}

		writeResponse(ctx, w, r, http.StatusCreated, publicTodoItemFromRepoModel(res))
	}
}

func deleteTodo(todoRepo todo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		todoId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeError(ctx, w, r, http.StatusBadRequest, errors.New("can not parse the todo id from the url"))
			return
		}

		userAndSession := auth.MustUserAndSessionFromContext(ctx)

		err = todoRepo.DeleteTodo(ctx, int(userAndSession.UserID), todoId)
		if err != nil {
			writeError(ctx, w, r, return400IfApp404IfNoResultErrOr500(err), err)
			return
		}
		apiWriteOperationDoneSuccessfullyJson(ctx, w, r)
	}
}

type publicTodoItem struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func publicTodoItemFromRepoModel(i todo.TodoItem) publicTodoItem {
	return publicTodoItem{
		Id:        i.Id,
		Title:     i.Title,
		Body:      i.Body,
		Status:    i.Status.String(),
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}
