package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mtlynch/whatgotdone/backend/datastore"
	"github.com/mtlynch/whatgotdone/backend/types"
)

type mockDatastore struct {
	journalEntries []types.JournalEntry
	journalDrafts  []types.JournalEntry
	users          []string
	reactions      []types.Reaction
}

func (ds mockDatastore) Users() ([]string, error) {
	return ds.users, nil
}

func (ds mockDatastore) All(username string) ([]types.JournalEntry, error) {
	return ds.journalEntries, nil
}

func (ds mockDatastore) GetDraft(username string, date string) (types.JournalEntry, error) {
	if len(ds.journalDrafts) > 0 {
		return ds.journalDrafts[0], nil
	}
	return types.JournalEntry{}, datastore.DraftNotFoundError{
		Username: username,
		Date:     date,
	}
}

func (ds mockDatastore) Insert(username string, j types.JournalEntry) error {
	return nil
}

func (ds mockDatastore) InsertDraft(username string, j types.JournalEntry) error {
	return nil
}

func (ds mockDatastore) Close() error {
	return nil
}

type mockAuthenticator struct {
	tokensToUsers map[string]string
}

func (a mockAuthenticator) UserFromAuthToken(authToken string) (string, error) {
	for k, v := range a.tokensToUsers {
		if k == authToken {
			return v, nil
		}
	}
	return "", errors.New("mock token not found")
}

func init() {
	// The handler uses relative paths to find the template file. Switch to the
	// app's root directory so that the relative paths work.
	if err := os.Chdir("../"); err != nil {
		panic(err)
	}
}

func TestDraftHandlerWhenUserIsNotLoggedIn(t *testing.T) {
	drafts := []types.JournalEntry{
		types.JournalEntry{Date: "2019-04-19", LastModified: "2019-04-19", Markdown: "Drove to the zoo"},
	}
	ds := mockDatastore{
		journalDrafts: drafts,
	}
	router := mux.NewRouter()
	s := defaultServer{
		authenticator: mockAuthenticator{},
		datastore:     &ds,
		router:        router,
	}
	s.routes()

	req, err := http.NewRequest("GET", "/api/draft/2019-04-19", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusForbidden {
		t.Fatalf("handler returned wrong status code: got %v want %v",
			status, http.StatusForbidden)
	}
}

func TestDraftHandlerWhenUserTokenIsInvalid(t *testing.T) {
	drafts := []types.JournalEntry{
		types.JournalEntry{Date: "2019-04-19", LastModified: "2019-04-19", Markdown: "Drove to the zoo"},
	}
	ds := mockDatastore{
		journalDrafts: drafts,
	}
	router := mux.NewRouter()
	s := defaultServer{
		authenticator: mockAuthenticator{
			tokensToUsers: map[string]string{
				"mock_token_A": "dummyUser",
			},
		},
		datastore: &ds,
		router:    router,
	}
	s.routes()

	req, err := http.NewRequest("GET", "/api/draft/2019-04-19", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", fmt.Sprintf("%s=mock_token_invalid", userKitAuthCookieName))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusForbidden {
		t.Fatalf("handler returned wrong status code: got %v want %v",
			status, http.StatusForbidden)
	}
}

func TestDraftHandlerWhenDateMatches(t *testing.T) {
	drafts := []types.JournalEntry{
		types.JournalEntry{Date: "2019-04-19", LastModified: "2019-04-19", Markdown: "Drove to the zoo"},
	}
	ds := mockDatastore{
		journalDrafts: drafts,
	}
	router := mux.NewRouter()
	s := defaultServer{
		authenticator: mockAuthenticator{
			tokensToUsers: map[string]string{
				"mock_token_A": "dummyUser",
			},
		},
		datastore: &ds,
		router:    router,
	}
	s.routes()

	req, err := http.NewRequest("GET", "/api/draft/2019-04-19", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", fmt.Sprintf("%s=mock_token_A", userKitAuthCookieName))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	var response types.JournalEntry
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Response is not valid JSON: %v", w.Body.String())
	}

	if !reflect.DeepEqual(response, drafts[0]) {
		t.Fatalf("Unexpected response: got %v want %v", response, drafts[0])
	}
}

func TestDraftHandlerReturns404WhenDatastoreReturnsEntryNotFoundError(t *testing.T) {
	entries := []types.JournalEntry{}
	ds := mockDatastore{
		journalDrafts: entries,
	}
	router := mux.NewRouter()
	s := defaultServer{
		authenticator: mockAuthenticator{
			tokensToUsers: map[string]string{
				"mock_token_A": "dummyUser",
			},
		},
		datastore: &ds,
		router:    router,
	}
	s.routes()

	req, err := http.NewRequest("GET", "/api/draft/2019-04-19", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", fmt.Sprintf("%s=mock_token_A", userKitAuthCookieName))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusNotFound {
		t.Fatalf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestDraftHandlerReturnsBadRequestWhenDateIsInvalid(t *testing.T) {
	entries := []types.JournalEntry{}
	ds := mockDatastore{
		journalDrafts: entries,
	}
	router := mux.NewRouter()
	s := defaultServer{
		authenticator: mockAuthenticator{
			tokensToUsers: map[string]string{
				"mock_token_A": "dummyUser",
			},
		},
		datastore: &ds,
		router:    router,
	}
	s.routes()

	req, err := http.NewRequest("GET", "/api/draft/201904-19", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", fmt.Sprintf("%s=mock_token_A", userKitAuthCookieName))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusBadRequest {
		t.Fatalf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}
