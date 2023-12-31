package http

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	planetscale "github.com/harshav17/planet_scale"
	db_mock "github.com/harshav17/planet_scale/mock/db"
)

func TestHandleSettlements_All(t *testing.T) {
	server := MustOpenServer(t)
	defer MustCloseServer(t, server.Server)

	t.Run("GET /groups/1/settlements", func(t *testing.T) {
		t.Run("settlement successful find", func(t *testing.T) {
			server.repos.Settlement = &db_mock.SettlementRepo{
				FindFn: func(tx *sql.Tx, filter planetscale.SettlementFilter) ([]*planetscale.Settlement, error) {
					return []*planetscale.Settlement{
						{
							GroupID: 1,
						},
					}, nil
				},
			}

			req, err := http.NewRequest("GET", "/groups/1/settlements", nil)
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

			var got findSettlementsResponse
			err = json.Unmarshal(rr.Body.Bytes(), &got)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Settlements) != 1 && got.N != 1 {
				t.Errorf("expected 1 settlement, got %d", len(got.Settlements))
			}
			if got.Settlements[0].GroupID != 1 {
				t.Errorf("expected group id 1, got %d", got.Settlements[0].GroupID)
			}
		})
	})

	t.Run("POST /settlements", func(t *testing.T) {
		t.Run("successful create", func(t *testing.T) {
			server.repos.Settlement = &db_mock.SettlementRepo{
				CreateFn: func(tx *sql.Tx, s *planetscale.Settlement) error {
					return nil
				},
			}

			settlement := planetscale.Settlement{
				GroupID: 1,
			}
			body, err := json.Marshal(settlement)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest("POST", "/settlements", bytes.NewReader(body))
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

	t.Run("GET /settlements/{id}", func(t *testing.T) {
		t.Run("successful get", func(t *testing.T) {
			server.repos.Settlement = &db_mock.SettlementRepo{
				GetFn: func(tx *sql.Tx, settlementID int64) (*planetscale.Settlement, error) {
					return &planetscale.Settlement{
						SettlementID: settlementID,
						GroupID:      1,
					}, nil
				},
			}

			req, err := http.NewRequest("GET", "/settlements/1", nil)
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

			var got planetscale.Settlement
			err = json.Unmarshal(rr.Body.Bytes(), &got)
			if err != nil {
				t.Fatal(err)
			}
			if got.SettlementID != 1 {
				t.Errorf("expected settlement id 1, got %d", got.SettlementID)
			}
			if got.GroupID != 1 {
				t.Errorf("expected group id 1, got %d", got.GroupID)
			}
		})
	})

	t.Run("PATCH /settlements/{id}", func(t *testing.T) {
		t.Run("successful update", func(t *testing.T) {
			server.repos.Settlement = &db_mock.SettlementRepo{
				UpdateFn: func(tx *sql.Tx, settlementID int64, settlement *planetscale.SettlementUpdate) (*planetscale.Settlement, error) {
					return nil, nil
				},
			}

			settlement := planetscale.Settlement{
				SettlementID: 1,
				GroupID:      1,
			}
			body, err := json.Marshal(settlement)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest("PATCH", "/settlements/1", bytes.NewReader(body))
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
		})
	})

	t.Run("DELETE /settlements/{id}", func(t *testing.T) {
		t.Run("successful delete", func(t *testing.T) {
			server.repos.Settlement = &db_mock.SettlementRepo{
				DeleteFn: func(tx *sql.Tx, settlementID int64) error {
					return nil
				},
			}

			req, err := http.NewRequest("DELETE", "/settlements/1", nil)
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
