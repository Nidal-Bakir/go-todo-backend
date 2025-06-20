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
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/paginate"
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
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}

		userAndSession, ok := auth.UserAndSessionFromContext(ctx)
		utils.Assert(ok, "we should find the user in the context tree, but we did not. something is wrong.")

		todoData, err := extractTodoData(r)
		if err != nil {
			writeError(ctx, w, return400IfAppErrOr500(err), err)
			return
		}

		res, err := todoRepo.CreateTodo(ctx, int(userAndSession.UserID), todoData)
		if err != nil {
			writeError(ctx, w, return400IfAppErrOr500(err), err)
			return
		}

		writeJson(ctx, w, http.StatusCreated, publicTodoItemFromRepoModel(res))
	}
}

func extractTodoData(r *http.Request) (todo.TodoData, error) {
	data := todo.TodoData{}

	var title string
	title = r.FormValue("title")
	if len(title) != 0 {
		data.Title = &title
	}

	var body string
	body = r.FormValue("body")
	if len(body) != 0 {
		data.Body = &body
	}

	status := new(todo.TodoStatus)
	statusStr := r.FormValue("status")
	if len(statusStr) != 0 {
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
			writeError(ctx, w, http.StatusBadRequest, errors.New("can not parse the todo id from the url"))
			return
		}

		err = r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}

		userAndSession, ok := auth.UserAndSessionFromContext(ctx)
		utils.Assert(ok, "we should find the user in the context tree, but we did not. something is wrong.")

		todoData, err := extractTodoData(r)
		if err != nil {
			writeError(ctx, w, return400IfApp404IfNoResultErrOr500(err), err)
			return
		}

		res, err := todoRepo.UpdateTodo(ctx, int(userAndSession.UserID), todoId, todoData)
		if err != nil {
			writeError(ctx, w, return400IfAppErrOr500(err), err)
			return
		}

		writeJson(ctx, w, http.StatusCreated, publicTodoItemFromRepoModel(res))
	}
}

func todoIndex(todoRepo todo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, err)
			return
		}

		userAndSession, ok := auth.UserAndSessionFromContext(ctx)
		utils.Assert(ok, "we should find the user in the context tree, but we did not. something is wrong.")

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
			writeError(ctx, w, return400IfAppErrOr500(err), err)
			return
		}

		// convert items to public form
		publicPaginatedDate := paginate.PaginatedDataMapper(paginatedDate, publicTodoItemFromRepoModel)

		writeJson(ctx, w, http.StatusOK, publicPaginatedDate)
	}
}

func todoShow(todoRepo todo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		todoId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, errors.New("can not parse the todo id from the url"))
			return
		}

		userAndSession, ok := auth.UserAndSessionFromContext(ctx)
		utils.Assert(ok, "we should find the user in the context tree, but we did not. something is wrong.")

		res, err := todoRepo.GetTodo(ctx, int(userAndSession.UserID), todoId)
		if err != nil {
			writeError(ctx, w, return400IfApp404IfNoResultErrOr500(err), err)
			return
		}

		writeJson(ctx, w, http.StatusCreated, publicTodoItemFromRepoModel(res))
	}
}

func deleteTodo(todoRepo todo.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		todoId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeError(ctx, w, http.StatusBadRequest, errors.New("can not parse the todo id from the url"))
			return
		}

		userAndSession, ok := auth.UserAndSessionFromContext(ctx)
		utils.Assert(ok, "we should find the user in the context tree, but we did not. something is wrong.")

		err = todoRepo.DeleteTodo(ctx, int(userAndSession.UserID), todoId)
		if err != nil {
			writeError(ctx, w, return400IfApp404IfNoResultErrOr500(err), err)
			return
		}
		writeOperationDoneSuccessfullyJson(ctx, w)
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
