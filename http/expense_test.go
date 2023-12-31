package http

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	planetscale "github.com/harshav17/planet_scale"
	db_mock "github.com/harshav17/planet_scale/mock/db"
)

func TestHandleExpense_All(t *testing.T) {
	t.Parallel()

	server := MustOpenServer(t)
	defer MustCloseServer(t, server.Server)

	t.Run("GET /groups/1/expenses", func(t *testing.T) {
		t.Run("successful find", func(t *testing.T) {
			server.repos.Expense = &db_mock.ExpenseRepo{
				FindFn: func(tx *sql.Tx, filter planetscale.ExpenseFilter) ([]*planetscale.Expense, error) {
					return []*planetscale.Expense{
						{
							GroupID:     1,
							PaidBy:      "test-user-id",
							Amount:      100,
							Description: "test expense",
							Timestamp:   time.Now(),
							CreatedBy:   "test-user-id",
							UpdatedBy:   "test-user-id",
						},
					}, nil
				},
			}

			req, err := http.NewRequest("GET", "/groups/1/expenses", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")

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
			if got.Expenses[0].GroupID != 1 {
				t.Errorf("expected group id 1, got %d", got.Expenses[0].GroupID)
			}
		})
	})

	t.Run("POST /expenses", func(t *testing.T) {
		t.Run("successful post", func(t *testing.T) {
			server.repos.Expense = &db_mock.ExpenseRepo{
				CreateFn: func(tx *sql.Tx, expense *planetscale.Expense) error {
					return nil
				},
			}

			expense := planetscale.Expense{
				GroupID:     1,
				PaidBy:      "test-user-id",
				Amount:      100,
				Description: "test expense",
				Timestamp:   time.Now(),
				CreatedBy:   "test-user-id",
				UpdatedBy:   "test-user-id",
			}

			body, err := json.Marshal(expense)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest("POST", "/expenses", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.router.ServeHTTP)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusCreated {
				t.Errorf("expected status code %d, got %d", http.StatusCreated, status)
			}
		})
	})

	t.Run("GET /expenses/{id}", func(t *testing.T) {
		t.Run("successful get", func(t *testing.T) {
			server.repos.Expense = &db_mock.ExpenseRepo{
				GetFn: func(tx *sql.Tx, expenseID int64) (*planetscale.Expense, error) {
					return &planetscale.Expense{
						GroupID:     1,
						PaidBy:      "test-user-id",
						Amount:      100,
						Description: "test expense",
						Timestamp:   time.Now(),
						CreatedBy:   "test-user-id",
						UpdatedBy:   "test-user-id",
					}, nil
				},
			}

			req, err := http.NewRequest("GET", "/expenses/1", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")

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
			if got.GroupID != 1 {
				t.Errorf("expected group id 1, got %d", got.GroupID)
			}
		})
	})

	t.Run("PATCH /expenses/{id}", func(t *testing.T) {
		t.Run("successful update", func(t *testing.T) {
			server.repos.Expense = &db_mock.ExpenseRepo{
				UpdateFn: func(tx *sql.Tx, expenseID int64, update *planetscale.ExpenseUpdate) (*planetscale.Expense, error) {
					return &planetscale.Expense{
						GroupID:     1,
						PaidBy:      "test-user-id",
						Amount:      100,
						Description: "test expense",
						Timestamp:   time.Now(),
						CreatedBy:   "test-user-id",
						UpdatedBy:   "test-user-id",
					}, nil
				},
			}

			newAmount := float64(200)
			expenseUpdate := planetscale.ExpenseUpdate{
				Amount: &newAmount,
			}

			body, err := json.Marshal(expenseUpdate)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest("PATCH", "/expenses/1", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")

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
			if got.GroupID != 1 {
				t.Errorf("expected group id 1, got %d", got.GroupID)
			}
		})
	})

	t.Run("DELETE /expenses/{id}", func(t *testing.T) {
		t.Run("successful delete", func(t *testing.T) {
			server.repos.Expense = &db_mock.ExpenseRepo{
				DeleteFn: func(tx *sql.Tx, expenseID int64) error {
					return nil
				},
			}

			req, err := http.NewRequest("DELETE", "/expenses/1", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.router.ServeHTTP)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusNoContent {
				t.Errorf("expected status code %d, got %d", http.StatusNoContent, status)
			}
		})
	})
}
