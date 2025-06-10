package todo

import (
	"context"
	"errors"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Repository interface {
	GetTodos(ctx context.Context, userId, offset, limit int) ([]TodoItem, error)
	GetTodo(ctx context.Context, userId, todoId int) (TodoItem, error)

	CreateTodo(ctx context.Context, userId int, data TodoData) (TodoItem, error)

	UpdateTodo(ctx context.Context, userId, todoId int, data TodoData) (TodoItem, error)

	DeleteTodo(ctx context.Context, userId, todoId int) error
}

func NewRepository(db *database.Service, redis *redis.Client) Repository {
	return &repositoryImpl{db: db, redis: redis}
}

// ---------------------------------------------------------------------------------

type repositoryImpl struct {
	db    *database.Service
	redis *redis.Client
}

func (repo repositoryImpl) GetTodos(ctx context.Context, userId, offset, limit int) ([]TodoItem, error) {
	zlog := zerolog.Ctx(ctx)

	data, err := repo.db.Queries.TodoGetTodosForUser(
		ctx,
		database.TodoGetTodosForUserParams{
			UserID: int32(userId),
			Offset: int64(offset),
			Limit:  int64(limit),
		},
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = apperr.ErrNoResult
		} else {
			zlog.Err(err).Msg("can not get todo")
		}
		return []TodoItem{}, err
	}

	todoItems := make([]TodoItem, len(data))

	for i, v := range data {
		todoItem, err := todoItemFromDataBase(v)
		if err != nil {
			zlog.Err(err).Msg("can not convert database.Todo to TodoItem")
			return []TodoItem{}, err
		}
		todoItems[i] = todoItem
	}

	return todoItems, nil
}

func (repo repositoryImpl) GetTodo(ctx context.Context, userId, todoId int) (TodoItem, error) {
	zlog := zerolog.Ctx(ctx).With().Int("todo_id", todoId).Logger()

	res, err := repo.db.Queries.TodoGetTodoLinkedToUser(
		ctx,
		database.TodoGetTodoLinkedToUserParams{
			ID:     int32(todoId),
			UserID: int32(userId),
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = apperr.ErrNoResult
		} else {
			zlog.Err(err).Msg("can not get todo")
		}
		return TodoItem{}, err
	}

	createdTodo, err := todoItemFromDataBase(res)
	if err != nil {
		zlog.Err(err).Msg("can not convert database.Todo to TodoItem")
		return TodoItem{}, err
	}

	return createdTodo, nil
}

func (repo repositoryImpl) CreateTodo(ctx context.Context, userId int, data TodoData) (TodoItem, error) {
	zlog := zerolog.Ctx(ctx)

	nilToEmptyString := func(strPtr *string) string {
		if strPtr == nil {
			return ""
		}
		return *strPtr
	}

	var status TodoStatus
	if data.Status == nil {
		status = TodoStatusPending
	} else {
		status = *data.Status
	}

	res, err := repo.db.Queries.TodoCreateTodo(
		ctx,
		database.TodoCreateTodoParams{
			Title:  nilToEmptyString(data.Title),
			Body:   nilToEmptyString(data.Body),
			Status: status.String(),
			UserID: int32(userId),
		},
	)
	if err != nil {
		zlog.Err(err).Msg("can not create todo")
		return TodoItem{}, err
	}

	createdTodo, err := todoItemFromDataBase(res)
	if err != nil {
		zlog.Err(err).Msg("can not convert database.Todo to TodoItem")
		return TodoItem{}, err
	}

	return createdTodo, nil
}

func (repo repositoryImpl) UpdateTodo(ctx context.Context, userId, todoId int, data TodoData) (TodoItem, error) {
	zlog := zerolog.Ctx(ctx).With().Int("todo_id", todoId).Logger()

	stringToPgTextType := func(strPtr *string) pgtype.Text {
		var txt pgtype.Text
		if strPtr != nil {
			txt.String = *strPtr
			txt.Valid = true
		}
		return txt
	}

	var status pgtype.Text
	if data.Status != nil {
		status.Valid = true
		status.String = data.Status.String()
	}

	res, err := repo.db.Queries.TodoUpdateTodo(
		ctx, database.TodoUpdateTodoParams{
			ID:     int32(todoId),
			UserID: int32(userId),
			Title:  stringToPgTextType(data.Title),
			Body:   stringToPgTextType(data.Body),
			Status: status,
		},
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = apperr.ErrNoResult
		} else {
			zlog.Err(err).Msg("can not create todo")
		}
		return TodoItem{}, err
	}

	createdTodo, err := todoItemFromDataBase(res)
	if err != nil {
		zlog.Err(err).Msg("can not convert database.Todo to TodoItem")
		return TodoItem{}, err
	}

	return createdTodo, nil
}

func (repo repositoryImpl) DeleteTodo(ctx context.Context, userId, todoId int) error {
	zlog := zerolog.Ctx(ctx).With().Int("todo_id", todoId).Logger()

	err := repo.db.Queries.TodoSoftDeleteTodoLinkedToUser(
		ctx,
		database.TodoSoftDeleteTodoLinkedToUserParams{
			ID:     int32(todoId),
			UserID: int32(userId),
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		err = apperr.ErrNoResult
	} else {
		zlog.Err(err).Msg("can not delete todo")
	}

	return err
}
