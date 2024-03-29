package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	planetscale "github.com/harshav17/planet_scale"
	db_mock "github.com/harshav17/planet_scale/mock/db"
	service_mock "github.com/harshav17/planet_scale/mock/service"
)

func TestHandleExpense_All(t *testing.T) {
	t.Parallel()

	server := MustOpenServer(t)
	defer MustCloseServer(t, server.Server)

	t.Run("GET /groups/1/expenses", func(t *testing.T) {
		t.Run("successful find", func(t *testing.T) {
			userID := "test-user-id"
			groupID := int64(1)
			server.repos.Expense = &db_mock.ExpenseRepo{
				FindFn: func(tx *sql.Tx, filter planetscale.ExpenseFilter) ([]*planetscale.Expense, error) {
					return []*planetscale.Expense{
						{
							GroupID:     &groupID,
							PaidBy:      userID,
							Amount:      100,
							Description: "test expense",
							Timestamp:   time.Now(),
						},
					}, nil
				},
			}
			server.repos.GroupMember = &db_mock.GroupMemberRepo{
				GetFn: func(tx *sql.Tx, groupID int64, userID string) (*planetscale.GroupMember, error) {
					return &planetscale.GroupMember{
						GroupID: 1,
						UserID:  userID,
					}, nil
				},
			}
			server.repos.ExpenseParticipant = &db_mock.ExpenseParticipantRepo{
				FindFn: func(tx *sql.Tx, filter planetscale.ExpenseParticipantFilter) ([]*planetscale.ExpenseParticipant, error) {
					return []*planetscale.ExpenseParticipant{
						{
							ExpenseID:       1,
							UserID:          userID,
							AmountOwed:      100,
							SharePercentage: 100,
							Note:            "test expense",
						},
					}, nil
				},
			}

			token := server.buildJWTForTesting(t, userID)
			req, err := http.NewRequest("GET", "/groups/1/expenses", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.router.ServeHTTP)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, status)
			}

			var got findExpensesResponse
			err = json.Unmarshal(rr.Body.Bytes(), &got)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Expenses) != 1 && got.N != 1 {
				t.Errorf("expected 1 expense, got %d", len(got.Expenses))
			}
			if *got.Expenses[0].GroupID != groupID {
				t.Errorf("expected group id 1, got %d", *got.Expenses[0].GroupID)
			}
		})

		t.Run("user not a member of group", func(t *testing.T) {
			userID := "test-user-id"
			groupID := int64(1)
			server.repos.Expense = &db_mock.ExpenseRepo{
				FindFn: func(tx *sql.Tx, filter planetscale.ExpenseFilter) ([]*planetscale.Expense, error) {
					return []*planetscale.Expense{
						{
							GroupID:     &groupID,
							PaidBy:      userID,
							Amount:      100,
							Description: "test expense",
							Timestamp:   time.Now(),
						},
					}, nil
				},
			}
			server.repos.GroupMember = &db_mock.GroupMemberRepo{
				GetFn: func(tx *sql.Tx, groupID int64, userID string) (*planetscale.GroupMember, error) {
					return nil, planetscale.Errorf(planetscale.ENOTFOUND, "no group member found with ID %d", groupID)
				},
			}

			token := server.buildJWTForTesting(t, userID)
			req, err := http.NewRequest("GET", "/groups/1/expenses", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.router.ServeHTTP)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("expected status code %d, got %d", http.StatusNotFound, status)
			}
		})
	})

	t.Run("POST /expenses", func(t *testing.T) {
		t.Run("successful post", func(t *testing.T) {
			userID := "test-user-id"
			groupID := int64(1)
			server.services.Expense = &service_mock.ExpenseService{
				CreateExpenseFn: func(ctx context.Context, expense *planetscale.Expense) error {
					expense.ExpenseID = 1
					return nil
				},
			}

			expense := planetscale.Expense{
				GroupID:     &groupID,
				PaidBy:      userID,
				Amount:      100,
				Description: "test expense",
				Timestamp:   time.Now(),
				SplitTypeID: 1,
			}

			body, err := json.Marshal(expense)
			if err != nil {
				t.Fatal(err)
			}

			token := server.buildJWTForTesting(t, userID)
			req, err := http.NewRequest("POST", "/expenses", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.router.ServeHTTP)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusCreated {
				t.Errorf("expected status code %d, got %d", http.StatusCreated, status)
			}

			var got planetscale.Expense
			err = json.Unmarshal(rr.Body.Bytes(), &got)
			if err != nil {
				t.Fatal(err)
			}
			if got.ExpenseID != 1 {
				t.Errorf("expected expense id 1, got %d", got.ExpenseID)
			} else if got.CreatedBy != userID {
				t.Errorf("expected created by %s, got %s", userID, got.CreatedBy)
			} else if got.UpdatedBy != userID {
				t.Errorf("expected updated by %s, got %s", userID, got.UpdatedBy)
			} else if got.PaidBy != userID {
				t.Errorf("expected paid by %s, got %s", userID, got.PaidBy)
			}
		})
	})

	t.Run("GET /expenses/{id}", func(t *testing.T) {
		t.Run("successful get", func(t *testing.T) {
			userID := "test-user-id"
			groupID := int64(1)
			server.repos.Expense = &db_mock.ExpenseRepo{
				GetFn: func(tx *sql.Tx, expenseID int64) (*planetscale.Expense, error) {
					return &planetscale.Expense{
						GroupID:     &groupID,
						PaidBy:      userID,
						Amount:      100,
						Description: "test expense",
						Timestamp:   time.Now(),
					}, nil
				},
			}
			server.repos.GroupMember = &db_mock.GroupMemberRepo{
				GetFn: func(tx *sql.Tx, groupID int64, userID string) (*planetscale.GroupMember, error) {
					return &planetscale.GroupMember{
						GroupID: 1,
						UserID:  userID,
					}, nil
				},
			}
			server.repos.ExpenseParticipant = &db_mock.ExpenseParticipantRepo{
				FindFn: func(tx *sql.Tx, filter planetscale.ExpenseParticipantFilter) ([]*planetscale.ExpenseParticipant, error) {
					return []*planetscale.ExpenseParticipant{
						{
							ExpenseID:       1,
							UserID:          userID,
							AmountOwed:      100,
							SharePercentage: 100,
							Note:            "test expense",
						},
					}, nil
				},
			}

			token := server.buildJWTForTesting(t, userID)
			req, err := http.NewRequest("GET", "/expenses/1", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.router.ServeHTTP)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, status)
			}

			var got planetscale.Expense
			err = json.Unmarshal(rr.Body.Bytes(), &got)
			if err != nil {
				t.Fatal(err)
			}
			if *got.GroupID != groupID {
				t.Errorf("expected group id 1, got %d", *got.GroupID)
			}
		})

		t.Run("user not a member of group", func(t *testing.T) {
			userID := "test-user-id"
			groupID := int64(1)
			server.repos.Expense = &db_mock.ExpenseRepo{
				GetFn: func(tx *sql.Tx, expenseID int64) (*planetscale.Expense, error) {
					return &planetscale.Expense{
						GroupID:     &groupID,
						PaidBy:      userID,
						Amount:      100,
						Description: "test expense",
						Timestamp:   time.Now(),
					}, nil
				},
			}
			server.repos.GroupMember = &db_mock.GroupMemberRepo{
				GetFn: func(tx *sql.Tx, groupID int64, userID string) (*planetscale.GroupMember, error) {
					return nil, planetscale.Errorf(planetscale.ENOTFOUND, "no group member found with ID %d", groupID)
				},
			}

			token := server.buildJWTForTesting(t, "test_user_id")
			req, err := http.NewRequest("GET", "/expenses/1", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.router.ServeHTTP)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("expected status code %d, got %d", http.StatusNotFound, status)
			}
		})
	})

	t.Run("PATCH /expenses/{id}", func(t *testing.T) {
		t.Run("successful update", func(t *testing.T) {
			userID := "test-user-id"
			newAmount := float64(200)
			groupID := int64(1)
			server.repos.Expense = &db_mock.ExpenseRepo{
				UpdateFn: func(tx *sql.Tx, expenseID int64, update *planetscale.ExpenseUpdate) (*planetscale.Expense, error) {
					return &planetscale.Expense{
						GroupID:     &groupID,
						PaidBy:      userID,
						Amount:      newAmount,
						Description: "test expense",
						Timestamp:   time.Now(),
						CreatedBy:   userID,
						UpdatedBy:   userID,
					}, nil
				},
				GetFn: func(tx *sql.Tx, expenseID int64) (*planetscale.Expense, error) {
					return &planetscale.Expense{
						GroupID:     &groupID,
						PaidBy:      userID,
						Amount:      100,
						Description: "test expense",
						Timestamp:   time.Now(),
						CreatedBy:   userID,
						UpdatedBy:   userID,
					}, nil
				},
			}
			server.repos.GroupMember = &db_mock.GroupMemberRepo{
				GetFn: func(tx *sql.Tx, groupID int64, userID string) (*planetscale.GroupMember, error) {
					return &planetscale.GroupMember{
						GroupID: 1,
						UserID:  userID,
					}, nil
				},
			}
			findCallCount := 0
			server.repos.ExpenseParticipant = &db_mock.ExpenseParticipantRepo{
				FindFn: func(tx *sql.Tx, filter planetscale.ExpenseParticipantFilter) ([]*planetscale.ExpenseParticipant, error) {
					findCallCount++
					if findCallCount == 1 {
						return []*planetscale.ExpenseParticipant{
							{
								ExpenseID:       1,
								UserID:          userID,
								AmountOwed:      100,
								SharePercentage: 100,
								Note:            "test expense",
							},
						}, nil
					} else {
						return []*planetscale.ExpenseParticipant{
							{
								ExpenseID:       1,
								UserID:          userID,
								AmountOwed:      200,
								SharePercentage: 100,
								Note:            "test expense",
							},
							{
								ExpenseID:       1,
								UserID:          "test-user-id-2",
								AmountOwed:      200,
								SharePercentage: 100,
								Note:            "test expense",
							},
						}, nil
					}
				},
				DeleteFn: func(tx *sql.Tx, expenseID int64, userID string) error {
					return nil
				},
				UpsertFn: func(tx *sql.Tx, expense *planetscale.ExpenseParticipant) error {
					return nil
				},
			}

			expenseUpdate := planetscale.ExpenseUpdate{
				Amount: &newAmount,
				Participants: []*planetscale.ExpenseParticipant{
					{
						ExpenseID: 1,
						UserID:    userID,
						Note:      "test expense",
					},
					{
						ExpenseID: 1,
						UserID:    "test-user-id-2",
						Note:      "test expense",
					},
				},
			}

			body, err := json.Marshal(expenseUpdate)
			if err != nil {
				t.Fatal(err)
			}

			token := server.buildJWTForTesting(t, userID)
			req, err := http.NewRequest("PATCH", "/expenses/1", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.router.ServeHTTP)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, status)
			}

			var got planetscale.Expense
			err = json.Unmarshal(rr.Body.Bytes(), &got)
			if err != nil {
				t.Fatal(err)
			}
			if *got.GroupID != groupID {
				t.Fatalf("expected group id 1, got %d", *got.GroupID)
			} else if got.Amount != newAmount {
				t.Fatalf("expected amount %f, got %f", newAmount, got.Amount)
			} else if len(got.Participants) != 2 {
				t.Fatalf("expected 2 participants, got %d", len(got.Participants))
			}
		})

		t.Run("user not a member of group", func(t *testing.T) {
			userID := "test-user-id"
			newAmount := float64(200)
			groupID := int64(1)
			server.repos.Expense = &db_mock.ExpenseRepo{
				UpdateFn: func(tx *sql.Tx, expenseID int64, update *planetscale.ExpenseUpdate) (*planetscale.Expense, error) {
					return &planetscale.Expense{
						GroupID:     &groupID,
						PaidBy:      userID,
						Amount:      newAmount,
						Description: "test expense",
						Timestamp:   time.Now(),
						CreatedBy:   userID,
						UpdatedBy:   userID,
					}, nil
				},
				GetFn: func(tx *sql.Tx, expenseID int64) (*planetscale.Expense, error) {
					return &planetscale.Expense{
						GroupID:     &groupID,
						PaidBy:      userID,
						Amount:      100,
						Description: "test expense",
						Timestamp:   time.Now(),
						CreatedBy:   userID,
						UpdatedBy:   userID,
					}, nil
				},
			}
			server.repos.GroupMember = &db_mock.GroupMemberRepo{
				GetFn: func(tx *sql.Tx, groupID int64, userID string) (*planetscale.GroupMember, error) {
					return nil, planetscale.Errorf(planetscale.ENOTFOUND, "no group member found with ID %d", groupID)
				},
			}

			expenseUpdate := planetscale.ExpenseUpdate{
				Amount: &newAmount,
			}
			body, err := json.Marshal(expenseUpdate)
			if err != nil {
				t.Fatal(err)
			}

			token := server.buildJWTForTesting(t, userID)
			req, err := http.NewRequest("PATCH", "/expenses/1", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.router.ServeHTTP)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("expected status code %d, got %d", http.StatusNotFound, status)
			}
		})
	})

	t.Run("DELETE /expenses/{id}", func(t *testing.T) {
		t.Run("successful delete", func(t *testing.T) {
			userID := "test-user-id"
			groupID := int64(1)
			server.repos.Expense = &db_mock.ExpenseRepo{
				DeleteFn: func(tx *sql.Tx, expenseID int64) error {
					return nil
				},
				GetFn: func(tx *sql.Tx, expenseID int64) (*planetscale.Expense, error) {
					return &planetscale.Expense{
						GroupID:     &groupID,
						PaidBy:      userID,
						Amount:      100,
						Description: "test expense",
						Timestamp:   time.Now(),
					}, nil
				},
			}
			server.repos.GroupMember = &db_mock.GroupMemberRepo{
				GetFn: func(tx *sql.Tx, groupID int64, userID string) (*planetscale.GroupMember, error) {
					return &planetscale.GroupMember{
						GroupID: 1,
						UserID:  userID,
					}, nil
				},
			}

			token := server.buildJWTForTesting(t, userID)
			req, err := http.NewRequest("DELETE", "/expenses/1", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.router.ServeHTTP)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusNoContent {
				t.Errorf("expected status code %d, got %d", http.StatusNoContent, status)
			}
		})

		t.Run("user not a member of group", func(t *testing.T) {
			userID := "test-user-id"
			groupID := int64(1)
			server.repos.Expense = &db_mock.ExpenseRepo{
				DeleteFn: func(tx *sql.Tx, expenseID int64) error {
					return nil
				},
				GetFn: func(tx *sql.Tx, expenseID int64) (*planetscale.Expense, error) {
					return &planetscale.Expense{
						GroupID:     &groupID,
						PaidBy:      userID,
						Amount:      100,
						Description: "test expense",
						Timestamp:   time.Now(),
					}, nil
				},
			}
			server.repos.GroupMember = &db_mock.GroupMemberRepo{
				GetFn: func(tx *sql.Tx, groupID int64, userID string) (*planetscale.GroupMember, error) {
					return nil, planetscale.Errorf(planetscale.ENOTFOUND, "no group member found with ID %d", groupID)
				},
			}

			token := server.buildJWTForTesting(t, userID)
			req, err := http.NewRequest("DELETE", "/expenses/1", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.router.ServeHTTP)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("expected status code %d, got %d", http.StatusNotFound, status)
			}
		})
	})
}
