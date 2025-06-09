package todo

import (
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
)

type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusDone       TodoStatus = "done"
)

func (s TodoStatus) String() string {
	return string(s)
}

func (l *TodoStatus) FromString(str string) (*TodoStatus, error) {
	switch {
	case TodoStatusPending.String() == str:
		*l = TodoStatusPending

	case TodoStatusInProgress.String() == str:
		*l = TodoStatusInProgress

	case TodoStatusDone.String() == str:
		*l = TodoStatusDone

	default:
		l = nil
		return l, apperr.ErrUnsupportedTodoStatus
	}

	return l, nil
}

type TodoItem struct {
	Id        int
	Title     string
	Body      string
	Status    TodoStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func todoItemFromDataBase(td database.Todo) (TodoItem, error) {
	status, err := new(TodoStatus).FromString(td.Status)
	if err != nil {
		return TodoItem{}, err
	}

	var delectedAt *time.Time
	if td.DeletedAt.Valid {
		delectedAt = &td.DeletedAt.Time
	}

	return TodoItem{
		Id:        int(td.ID),
		Title:     td.Title,
		Body:      td.Body,
		Status:    *status,
		CreatedAt: td.CreatedAt.Time,
		UpdatedAt: td.UpdatedAt.Time,
		DeletedAt: delectedAt,
	}, nil
}

type TodoData struct {
	Title  *string
	Body   *string
	Status *TodoStatus
}
