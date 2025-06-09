package todo

import (
	"context"

	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Repository interface {
	GetTodos(ctx context.Context, userId, perPage, page int) ([]TodoItem, error)
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

func (repo repositoryImpl) GetTodos(ctx context.Context, userId, perPage, page int) ([]TodoItem, error) {
	panic("not implemented")
}

func (repo repositoryImpl) GetTodo(ctx context.Context, userId, todoId int) (TodoItem, error) {
	zlog := zerolog.Ctx(ctx)

	res, err := repo.db.Queries.TodoGetTodoLinkedToUser(
		ctx,
		database.TodoGetTodoLinkedToUserParams{
			ID:     int32(todoId),
			UserID: int32(userId),
		},
	)
	if err != nil {
		zlog.Err(err).Msg("can not get todo")
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
	zlog := zerolog.Ctx(ctx)

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

func (repo repositoryImpl) DeleteTodo(ctx context.Context, userId, todoId int) error {
	return repo.db.Queries.TodoSoftDeleteTodoLinkedToUser(
		ctx,
		database.TodoSoftDeleteTodoLinkedToUserParams{
			ID:     int32(todoId),
			UserID: int32(userId),
		},
	)
}
