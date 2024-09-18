package user

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/mdobak/go-xerrors"
	"realworld/internal/utils/log"
	"time"
)

const userTableName = "user_data"

type UserRepo struct {
	Db *sql.DB
}

type UserStorageInterface interface {
	Create(User) (*User, error)
	GetUserById(id uint) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(id uint, fieldsToUpdate []string, newValues []interface{}) error
}

func (userRepo *UserRepo) Create(user User) (*User, error) {
	query := fmt.Sprintf("INSERT INTO %s (email, created_at, updated_at, username, password, salt) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", userTableName)
	currTime := time.Now().Format(time.RFC3339)
	user.UpdatedAt = currTime
	user.CreatedAt = currTime

	encrPassword, salt := GetPasswordSalt([]byte(user.Password))
	insertValues := []interface{}{user.Email, user.CreatedAt, user.UpdatedAt, user.Username, encrPassword, salt}
	var lastInsertedId int
	err := userRepo.Db.QueryRow(query, insertValues...).Scan(&lastInsertedId)
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, user)
		return nil, err
	}
	user.Id = uint(lastInsertedId)

	return &user, nil
}

func (userRepo *UserRepo) GetUserByEmail(email string) (*User, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE email = $1", userTableName)
	row := userRepo.Db.QueryRow(query, email)

	user := &User{}
	if err := row.Scan(&user.Id, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.Username, &user.Bio, &user.Password, &user.Salt); err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, email)
		return nil, err
	}

	return user, nil
}

func (userRepo *UserRepo) GetUserById(id uint) (*User, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", userTableName)
	row := userRepo.Db.QueryRow(query, id)
	user := &User{}
	if err := row.Scan(&user.Id, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.Username, &user.Bio, &user.Password, &user.Salt); err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, id)
		return nil, err
	}

	return user, nil
}

func (userRepo *UserRepo) UpdateUser(id uint, fieldsToUpdate []string, newValues []interface{}) error {
	fieldValuePairs := ""
	for i, name := range fieldsToUpdate {
		fieldValuePairs += fmt.Sprintf("%s = $%d, ", name, i+1)
	}
	updatedAt := time.Now().Format(time.RFC3339)
	fieldValuePairs += fmt.Sprintf("updated_at = $%d", len(newValues)+1)
	newValues = append(newValues, updatedAt)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", userTableName, fieldValuePairs, len(newValues)+1)
	newValues = append(newValues, id)

	_, err := userRepo.Db.Exec(query, newValues...)
	if err != nil {
		log.DbQueryCtx(context.Background(), xerrors.New(err), query, newValues)
		return err
	}
	return nil
}
