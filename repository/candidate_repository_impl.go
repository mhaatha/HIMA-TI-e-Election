package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/HIMA-TI-e-Election/model/domain"
)

// Variables to handle null values
var (
	vision  sql.NullString
	mission sql.NullString
)

func NewCandidateRepository() CandidateRepository {
	return &CandidateRepositoryImpl{}
}

type CandidateRepositoryImpl struct{}

func (repository *CandidateRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, candidate domain.Candidate) (domain.Candidate, error) {
	SQL := `
	INSERT INTO candidates (number, president, vice, vision, mission, photo_key, president_study_program, vice_study_program, president_nim, vice_nim)
	VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7, $8, $9, $10)
	RETURNING id, created_at, updated_at
	`

	err := tx.QueryRow(
		ctx,
		SQL,
		candidate.Number,
		candidate.President,
		candidate.Vice,
		candidate.Vision,
		candidate.Mission,
		candidate.PhotoKey,
		candidate.PresidentStudyProgram,
		candidate.ViceStudyProgram,
		candidate.PresidentNIM,
		candidate.ViceNIM,
	).Scan(
		&candidate.Id,
		&candidate.CreatedAt,
		&candidate.UpdatedAt,
	)
	if err != nil {
		return domain.Candidate{}, err
	}

	return candidate, nil
}

func (repository *CandidateRepositoryImpl) GetAll(ctx context.Context, tx pgx.Tx) ([]domain.Candidate, error) {
	SQL := `
		SELECT id, number, president, vice, vision, mission, photo_key, president_study_program, vice_study_program, president_nim, vice_nim, created_at, updated_at
		FROM candidates
	`

	rows, err := tx.Query(ctx, SQL)
	if err != nil {
		return []domain.Candidate{}, err
	}
	defer rows.Close()

	var candidates []domain.Candidate
	for rows.Next() {
		var candidate domain.Candidate

		err := rows.Scan(
			&candidate.Id,
			&candidate.Number,
			&candidate.President,
			&candidate.Vice,
			&vision,
			&mission,
			&candidate.PhotoKey,
			&candidate.PresidentStudyProgram,
			&candidate.ViceStudyProgram,
			&candidate.PresidentNIM,
			&candidate.ViceNIM,
			&candidate.CreatedAt,
			&candidate.UpdatedAt,
		)

		// Handle null fields
		if vision.Valid {
			candidate.Vision = vision.String
		}
		if mission.Valid {
			candidate.Mission = mission.String
		}

		if err != nil {
			return []domain.Candidate{}, err
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

func (repository *CandidateRepositoryImpl) GetByPeriod(ctx context.Context, tx pgx.Tx, period int) ([]domain.Candidate, error) {
	SQL := `
	SELECT id, number, president, vice, vision, mission, photo_key, president_study_program, vice_study_program, president_nim, vice_nim, created_at, updated_at
	FROM candidates
	WHERE EXTRACT(YEAR FROM created_at) = $1
	`

	rows, err := tx.Query(ctx, SQL, period)
	if err != nil {
		return []domain.Candidate{}, err
	}
	defer rows.Close()

	var candidates []domain.Candidate
	for rows.Next() {
		var candidate domain.Candidate

		err := rows.Scan(
			&candidate.Id,
			&candidate.Number,
			&candidate.President,
			&candidate.Vice,
			&vision,
			&mission,
			&candidate.PhotoKey,
			&candidate.PresidentStudyProgram,
			&candidate.ViceStudyProgram,
			&candidate.PresidentNIM,
			&candidate.ViceNIM,
			&candidate.CreatedAt,
			&candidate.UpdatedAt,
		)

		// Handle null fields
		if vision.Valid {
			candidate.Vision = vision.String
		}
		if mission.Valid {
			candidate.Mission = mission.String
		}

		if err != nil {
			return []domain.Candidate{}, err
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

func (repository *CandidateRepositoryImpl) GetById(ctx context.Context, tx pgx.Tx, candidateId int) (domain.Candidate, error) {
	SQL := `
	SELECT id, number, president, vice, vision, mission, photo_key, president_study_program, vice_study_program, president_nim, vice_nim, created_at, updated_at
	FROM candidates
	WHERE id = $1
	`

	var candidate domain.Candidate
	err := tx.QueryRow(ctx, SQL, candidateId).Scan(
		&candidate.Id,
		&candidate.Number,
		&candidate.President,
		&candidate.Vice,
		&vision,
		&mission,
		&candidate.PhotoKey,
		&candidate.PresidentStudyProgram,
		&candidate.ViceStudyProgram,
		&candidate.PresidentNIM,
		&candidate.ViceNIM,
		&candidate.CreatedAt,
		&candidate.UpdatedAt,
	)

	// Handle null fields
	if vision.Valid {
		candidate.Vision = vision.String
	}
	if mission.Valid {
		candidate.Mission = mission.String
	}

	if err != nil {
		return domain.Candidate{}, err
	}

	return candidate, nil
}

func (repository *CandidateRepositoryImpl) UpdateById(ctx context.Context, tx pgx.Tx, candidateId int, candidate domain.Candidate) (domain.Candidate, error) {
	SQL := `
	UPDATE candidates
	SET number = $1, president = $2, vice = NULLIF($3, ''), vision = NULLIF($4, ''), mission = $5, photo_key = $6, president_study_program = $7, vice_study_program = $8, president_nim = $9, vice_nim = $10, updated_at = $11
	WHERE id = $12
	`

	updatedAt := time.Now()

	_, err := tx.Exec(
		ctx,
		SQL,
		candidate.Number,
		candidate.President,
		candidate.Vice,
		candidate.Vision,
		candidate.Mission,
		candidate.PhotoKey,
		candidate.PresidentStudyProgram,
		candidate.ViceStudyProgram,
		candidate.PresidentNIM,
		candidate.ViceNIM,
		updatedAt,
		candidateId,
	)

	if err != nil {
		return domain.Candidate{}, err
	}

	return domain.Candidate{
		Id:                    candidateId,
		Number:                candidate.Number,
		President:             candidate.President,
		Vice:                  candidate.Vice,
		Vision:                candidate.Vision,
		Mission:               candidate.Mission,
		PhotoKey:              candidate.PhotoKey,
		PresidentStudyProgram: candidate.PresidentStudyProgram,
		ViceStudyProgram:      candidate.ViceStudyProgram,
		PresidentNIM:          candidate.PresidentNIM,
		ViceNIM:               candidate.ViceNIM,
		CreatedAt:             candidate.CreatedAt,
		UpdatedAt:             updatedAt,
	}, nil
}

func (repository *CandidateRepositoryImpl) DeleteById(ctx context.Context, tx pgx.Tx, candidateId int) error {
	SQL := `
	DELETE FROM candidates
	WHERE id = $1
	`

	_, err := tx.Exec(ctx, SQL, candidateId)
	return err
}
