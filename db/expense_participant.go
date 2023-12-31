package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	planetscale "github.com/harshav17/planet_scale"
)

type expenseParticipantRepo struct {
	db *DB
}

func NewExpenseParticipantRepo(db *DB) *expenseParticipantRepo {
	return &expenseParticipantRepo{
		db: db,
	}
}

func (r *expenseParticipantRepo) Get(tx *sql.Tx, expenseID int64, userID string) (*planetscale.ExpenseParticipant, error) {
	query := `
		SELECT
			expense_id,
			user_id,
			amount_owed,
			share_percentage,
			split_method,
			note
		FROM expense_participants
		WHERE expense_id = ? AND user_id = ?
	`

	var participant planetscale.ExpenseParticipant
	row := tx.QueryRow(query, expenseID, userID)
	err := row.Scan(&participant.ExpenseID, &participant.UserID, &participant.AmountOwed, &participant.SharePercentage, &participant.SplitMethod, &participant.Note)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle no rows error specifically if needed
			return nil, fmt.Errorf("no expense participant found with expenseID %d and userID %s", expenseID, userID)
		}
		return nil, err
	}
	slog.Info("loaded expense participant", slog.Int64("id", expenseID), slog.String("user_id", userID))

	return &participant, nil
}

func (r *expenseParticipantRepo) Create(tx *sql.Tx, participant *planetscale.ExpenseParticipant) error {
	query := `
		INSERT INTO expense_participants (
			expense_id,
			user_id,
			amount_owed,
			share_percentage,
			split_method,
			note
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := tx.Exec(query, participant.ExpenseID, participant.UserID, participant.AmountOwed, participant.SharePercentage, participant.SplitMethod, participant.Note)
	if err != nil {
		return err
	}
	_, err = result.LastInsertId()
	if err != nil {
		return err
	}
	slog.Info("created expense participant", slog.Int64("id", participant.ExpenseID), slog.String("user_id", participant.UserID))

	return nil
}

func (r *expenseParticipantRepo) Delete(tx *sql.Tx, expenseID int64, userID string) error {
	query := `
		DELETE FROM expense_participants
		WHERE expense_id = ? AND user_id = ?
	`

	result, err := tx.Exec(query, expenseID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no expense participant found with expenseID %d and userID %s", expenseID, userID)
	}
	slog.Info("deleted expense participant", slog.Int64("id", expenseID), slog.String("user_id", userID))

	return nil
}

func (r *expenseParticipantRepo) Update(tx *sql.Tx, expenseID int64, userID string, update *planetscale.ExpenseParticipantUpdate) (*planetscale.ExpenseParticipant, error) {
	participant, err := r.Get(tx, expenseID, userID)
	if err != nil {
		return nil, err
	}

	if update.AmountOwed != nil {
		participant.AmountOwed = *update.AmountOwed
	}
	if update.SharePercentage != nil {
		participant.SharePercentage = *update.SharePercentage
	}
	if update.SplitMethod != nil {
		participant.SplitMethod = *update.SplitMethod
	}
	if update.Note != nil {
		participant.Note = *update.Note
	}

	query := `
		UPDATE expense_participants
		SET amount_owed = ?, share_percentage = ?, split_method = ?, note = ?
		WHERE expense_id = ? AND user_id = ?
	`

	result, err := tx.Exec(query, participant.AmountOwed, participant.SharePercentage, participant.SplitMethod, participant.Note, expenseID, userID)
	if err != nil {
		return nil, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("no expense participant found with expenseID %d and userID %s", expenseID, userID)
	}
	slog.Info("updated expense participant", slog.Int64("id", expenseID), slog.String("user_id", userID))

	return r.Get(tx, expenseID, userID)
}
