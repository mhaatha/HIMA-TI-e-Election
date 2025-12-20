package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
)

// Variables to handle null values
var (
	nim          sql.NullString
	studyProgram sql.NullString
	password     sql.NullString
)

func NewUserRepository() UserRepository {
	return &UserRepositoryImpl{}
}

type UserRepositoryImpl struct{}

func (repository *UserRepositoryImpl) Save(ctx context.Context, tx *sql.Tx, user domain.User) (domain.User, error) {
	var SQL string

	if user.Role != "" {
		SQL = `
		INSERT INTO users (nim, full_name, study_program, password, role, phone_number)
		VALUES (NULLIF($1, ''), $2, NULLIF($3, ''), NULLIF($4, ''), $5, $6)
		RETURNING id, created_at, updated_at
		`

		err := tx.QueryRowContext(ctx, SQL, user.NIM, user.FullName, user.StudyProgram, user.Password, user.Role, user.PhoneNumber).Scan(
			&user.Id,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return domain.User{}, err
		}
	} else {
		SQL = `
		INSERT INTO users (nim, full_name, study_program, password, phone_number)
		VALUES (NULLIF($1, ''), $2, NULLIF($3, ''), NULLIF($4, ''), $5)
		RETURNING id, role, created_at, updated_at
		`

		err := tx.QueryRowContext(ctx, SQL, user.NIM, user.FullName, user.StudyProgram, user.Password, user.PhoneNumber).Scan(
			&user.Id,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return domain.User{}, err
		}
	}

	return user, nil
}

func (repository *UserRepositoryImpl) UpdatePartial(ctx context.Context, tx *sql.Tx, id int, updates map[string]interface{}) (domain.User, error) {
	if len(updates) == 0 {
		return domain.User{}, fmt.Errorf("updates is empty")
	}

	setQuery, args := helper.BuildUpdatePartialQuery(updates)
	args = append(args, id)

	SQL := fmt.Sprintf(`
	UPDATE users
	SET %s WHERE id = $%d
	RETURNING id, nim, full_name, study_program, role, created_at, updated_at
	`, setQuery, len(args))

	var user domain.User
	err := tx.QueryRowContext(ctx, SQL, args...).Scan(
		&user.Id,
		&nim,
		&user.FullName,
		&studyProgram,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// Handle null fields
	if nim.Valid {
		user.NIM = nim.String
	}
	if studyProgram.Valid {
		user.StudyProgram = studyProgram.String
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (repository *UserRepositoryImpl) GetByNIM(ctx context.Context, tx *sql.Tx, nim string) (domain.User, error) {
	SQL := `
	SELECT id, nim, full_name, study_program, password, role, phone_number, created_at, updated_at
	FROM users
	WHERE nim = $1
	`

	var user domain.User
	err := tx.QueryRowContext(ctx, SQL, nim).Scan(
		&user.Id,
		&user.NIM,
		&user.FullName,
		&studyProgram,
		&password,
		&user.Role,
		&user.PhoneNumber,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// Handle null fields
	if studyProgram.Valid {
		user.StudyProgram = studyProgram.String
	}
	if password.Valid {
		user.Password = password.String
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (repository *UserRepositoryImpl) GetUserBySession(ctx context.Context, tx *sql.Tx, sessionId string) (domain.User, error) {
	SQL := `
	SELECT
		u.id,
		u.nim,
		u.full_name,
		u.study_program,
		u.role,
		u.phone_number,
		u.created_at,
		u.updated_at
	FROM
		sessions s
	INNER JOIN
		users u ON s.user_id = u.id
	WHERE
		s.session_id = $1
	`

	var user domain.User
	err := tx.QueryRowContext(ctx, SQL, sessionId).Scan(
		&user.Id,
		&nim,
		&user.FullName,
		&studyProgram,
		&user.Role,
		&user.PhoneNumber,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// Handle null fields
	if nim.Valid {
		user.NIM = nim.String
	}
	if studyProgram.Valid {
		user.StudyProgram = studyProgram.String
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (repository *UserRepositoryImpl) GetAdmins(ctx context.Context, tx *sql.Tx) ([]domain.User, error) {
	SQL := `
	SELECT id, nim, full_name, study_program, password, role, phone_number, created_at, updated_at
	FROM users
	WHERE nim is null and role = $1
	`

	var users []domain.User
	rows, err := tx.QueryContext(ctx, SQL, "admin")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user domain.User

		err := rows.Scan(
			&user.Id,
			&nim,
			&user.FullName,
			&studyProgram,
			&password,
			&user.Role,
			&user.PhoneNumber,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		// Handle null fields
		if nim.Valid {
			user.NIM = nim.String
		}
		if studyProgram.Valid {
			user.StudyProgram = studyProgram.String
		}
		if password.Valid {
			user.Password = password.String
		}

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (repository *UserRepositoryImpl) GetById(ctx context.Context, tx *sql.Tx, userId int) (domain.User, error) {
	SQL := `
	SELECT id, nim, full_name, study_program, role, phone_number, created_at, updated_at
	FROM users
	WHERE id = $1
	`

	var user domain.User
	err := tx.QueryRowContext(ctx, SQL, userId).Scan(
		&user.Id,
		&nim,
		&user.FullName,
		&studyProgram,
		&user.Role,
		&user.PhoneNumber,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// Handle null fields
	if nim.Valid {
		user.NIM = nim.String
	}
	if studyProgram.Valid {
		user.StudyProgram = studyProgram.String
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (repository *UserRepositoryImpl) GetAll(ctx context.Context, tx *sql.Tx) ([]domain.User, error) {
	SQL := `
	SELECT id, nim, full_name, study_program, role, phone_number, created_at, updated_at
	FROM users
	`

	var users []domain.User
	rows, err := tx.QueryContext(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user domain.User

		err := rows.Scan(
			&user.Id,
			&nim,
			&user.FullName,
			&studyProgram,
			&user.Role,
			&user.PhoneNumber,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		// Handle null fields
		if nim.Valid {
			user.NIM = nim.String
		}
		if studyProgram.Valid {
			user.StudyProgram = studyProgram.String
		}

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (repository *UserRepositoryImpl) GetByIdWithPassword(ctx context.Context, tx *sql.Tx, userId int) (domain.User, error) {
	SQL := `
	SELECT id, nim, full_name, study_program, password, role, phone_number, created_at, updated_at
	FROM users
	WHERE id = $1
	`

	var user domain.User
	err := tx.QueryRowContext(ctx, SQL, userId).Scan(
		&user.Id,
		&nim,
		&user.FullName,
		&studyProgram,
		&password,
		&user.Role,
		&user.PhoneNumber,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// Handle null fields
	if nim.Valid {
		user.NIM = nim.String
	}
	if studyProgram.Valid {
		user.StudyProgram = studyProgram.String
	}
	if password.Valid {
		user.Password = password.String
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (repository *UserRepositoryImpl) UpdateById(ctx context.Context, tx *sql.Tx, userId int, user domain.User) (domain.User, error) {
	SQL := `
	UPDATE users
	SET nim = NULLIF($1, ''), full_name = $2, study_program = NULLIF($3, ''), password = NULLIF($4, ''), role = $5, phone_number = $6, updated_at = $7
	WHERE id = $8
	`

	updatedAt := time.Now()

	_, err := tx.ExecContext(
		ctx,
		SQL,
		user.NIM,
		user.FullName,
		user.StudyProgram,
		user.Password,
		user.Role,
		user.PhoneNumber,
		updatedAt,
		userId,
	)

	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:           user.Id,
		NIM:          user.NIM,
		FullName:     user.FullName,
		StudyProgram: user.StudyProgram,
		Password:     user.Password,
		Role:         user.Role,
		PhoneNumber:  user.PhoneNumber,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    updatedAt,
	}, nil
}

func (repository *UserRepositoryImpl) DeleteById(ctx context.Context, tx *sql.Tx, userId int) error {
	SQL := `
	DELETE FROM users
	WHERE id = $1
	`

	_, err := tx.ExecContext(ctx, SQL, userId)
	if err != nil {
		return err
	}

	return nil
}

func (repository *UserRepositoryImpl) GetByPhoneNumber(ctx context.Context, tx *sql.Tx, phoneNumber string) (domain.User, error) {
	SQL := `
	SELECT id, nim, full_name, study_program, password, role, phone_number, created_at, updated_at
	FROM users
	WHERE phone_number = $1
	`

	var user domain.User
	err := tx.QueryRowContext(ctx, SQL, phoneNumber).Scan(
		&user.Id,
		&nim,
		&user.FullName,
		&studyProgram,
		&password,
		&user.Role,
		&user.PhoneNumber,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// Handle null fields
	if nim.Valid {
		user.NIM = nim.String
	}
	if studyProgram.Valid {
		user.StudyProgram = studyProgram.String
	}
	if password.Valid {
		user.Password = password.String
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (repository *UserRepositoryImpl) SaveBulk(ctx context.Context, tx *sql.Tx, users []domain.User) ([]domain.User, error) {
	var (
		queryBuilder  strings.Builder
		args          []interface{}
		usersResponse []domain.User
	)

	queryBuilder.WriteString("INSERT INTO users (nim, full_name, study_program, phone_number) VALUES ")

	for i, user := range users {
		start := i*4 + 1

		queryBuilder.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d)", start, start+1, start+2, start+3))

		if i < len(users)-1 {
			queryBuilder.WriteString(", ")
		}
		if i == len(users)-1 {
			queryBuilder.WriteString(" RETURNING id, role, created_at, updated_at")
		}

		args = append(args, user.NIM, user.FullName, user.StudyProgram, user.PhoneNumber)
	}

	query := queryBuilder.String()

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	i := 0

	for rows.Next() {
		var user domain.User

		err := rows.Scan(
			&user.Id,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		user.NIM = users[i].NIM
		user.FullName = users[i].FullName
		user.StudyProgram = users[i].StudyProgram
		user.PhoneNumber = users[i].PhoneNumber

		usersResponse = append(usersResponse, user)

		i++
	}

	return usersResponse, nil
}

func (repository *UserRepositoryImpl) GetUsersWithNullPassword(ctx context.Context, tx *sql.Tx) ([]domain.User, error) {
	SQL := `
	SELECT id, nim, full_name, study_program, role, phone_number, created_at, updated_at
	FROM users
	WHERE password IS NULL
	`

	var users []domain.User
	rows, err := tx.QueryContext(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user domain.User

		err := rows.Scan(
			&user.Id,
			&nim,
			&user.FullName,
			&studyProgram,
			&user.Role,
			&user.PhoneNumber,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		// Handle null fields
		if nim.Valid {
			user.NIM = nim.String
		}
		if studyProgram.Valid {
			user.StudyProgram = studyProgram.String
		}

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (repository *UserRepositoryImpl) UpdateBulk(ctx context.Context, tx *sql.Tx, users []domain.User) error {
	var (
		queryBuilder strings.Builder
		args         []interface{}
	)

	queryBuilder.WriteString(`
	UPDATE users AS u
    SET password = v.password
    FROM (VALUES
	`)

	for i, user := range users {
		start := i*2 + 1

		queryBuilder.WriteString(fmt.Sprintf("($%d::integer, $%d)", start, start+1))

		if i < len(users)-1 {
			queryBuilder.WriteString(", ")
		}

		args = append(args, user.Id, user.Password)
	}

	queryBuilder.WriteString(`
    ) AS v(id, password)
    WHERE u.id = v.id
	`)

	query := queryBuilder.String()

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}
