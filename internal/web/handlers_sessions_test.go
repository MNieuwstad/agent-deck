package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/asheshgoplani/agent-deck/internal/session"
)

// fakeMutator is a test double for SessionMutator that delegates to function fields.
// If a function field is nil, the method returns an error indicating it is unconfigured.
type fakeMutator struct {
	createSessionFn  func(title, tool, projectPath, groupPath string) (string, error)
	startSessionFn   func(id string) error
	stopSessionFn    func(id string) error
	restartSessionFn func(id string) error
	deleteSessionFn  func(id string) error
	forkSessionFn    func(id string) (string, error)
	createGroupFn    func(name, parentPath string) (string, error)
	renameGroupFn    func(groupPath, newName string) error
	deleteGroupFn    func(groupPath string) error
}

func (f *fakeMutator) CreateSession(title, tool, projectPath, groupPath string) (string, error) {
	if f.createSessionFn == nil {
		return "", fmt.Errorf("createSession not configured")
	}
	return f.createSessionFn(title, tool, projectPath, groupPath)
}

func (f *fakeMutator) StartSession(id string) error {
	if f.startSessionFn == nil {
		return fmt.Errorf("startSession not configured")
	}
	return f.startSessionFn(id)
}

func (f *fakeMutator) StopSession(id string) error {
	if f.stopSessionFn == nil {
		return fmt.Errorf("stopSession not configured")
	}
	return f.stopSessionFn(id)
}

func (f *fakeMutator) RestartSession(id string) error {
	if f.restartSessionFn == nil {
		return fmt.Errorf("restartSession not configured")
	}
	return f.restartSessionFn(id)
}

func (f *fakeMutator) DeleteSession(id string) error {
	if f.deleteSessionFn == nil {
		return fmt.Errorf("deleteSession not configured")
	}
	return f.deleteSessionFn(id)
}

func (f *fakeMutator) ForkSession(id string) (string, error) {
	if f.forkSessionFn == nil {
		return "", fmt.Errorf("forkSession not configured")
	}
	return f.forkSessionFn(id)
}

func (f *fakeMutator) CreateGroup(name, parentPath string) (string, error) {
	if f.createGroupFn == nil {
		return "", fmt.Errorf("createGroup not configured")
	}
	return f.createGroupFn(name, parentPath)
}

func (f *fakeMutator) RenameGroup(groupPath, newName string) error {
	if f.renameGroupFn == nil {
		return fmt.Errorf("renameGroup not configured")
	}
	return f.renameGroupFn(groupPath, newName)
}

func (f *fakeMutator) DeleteGroup(groupPath string) error {
	if f.deleteGroupFn == nil {
		return fmt.Errorf("deleteGroup not configured")
	}
	return f.deleteGroupFn(groupPath)
}

func TestSessionsCollectionGET(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr: "127.0.0.1:0",
		Profile:    "test",
	})
	srv.menuData = &fakeMenuDataLoader{
		snapshot: &MenuSnapshot{
			Profile: "test",
			Items: []MenuItem{
				{
					Type: MenuItemTypeGroup,
					Group: &MenuGroup{
						Name: "work",
						Path: "work",
					},
				},
				{
					Type: MenuItemTypeSession,
					Session: &MenuSession{
						ID:     "sess-1",
						Title:  "alpha",
						Status: session.StatusRunning,
					},
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"sessions"`) {
		t.Errorf("expected 'sessions' key in response, got: %s", body)
	}
	if !strings.Contains(body, `"groups"`) {
		t.Errorf("expected 'groups' key in response, got: %s", body)
	}
	if !strings.Contains(body, `"sess-1"`) {
		t.Errorf("expected session id in response, got: %s", body)
	}
}

func TestSessionsCollectionPOSTCreatesSession(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	srv.mutator = &fakeMutator{
		createSessionFn: func(title, tool, projectPath, groupPath string) (string, error) {
			return "new-id", nil
		},
	}

	body := strings.NewReader(`{"title":"Test","tool":"claude","projectPath":"/tmp"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "new-id") {
		t.Errorf("expected session id in response, got: %s", rr.Body.String())
	}
}

func TestSessionsCollectionPOSTNilMutatorReturns503(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	// mutator is nil

	body := strings.NewReader(`{"title":"Test","tool":"claude","projectPath":"/tmp"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d: %s", http.StatusServiceUnavailable, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), ErrCodeNotImplemented) {
		t.Errorf("expected NOT_IMPLEMENTED error, got: %s", rr.Body.String())
	}
}

func TestSessionsCollectionPOSTMutationsDisabled(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: false,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}

	body := strings.NewReader(`{"title":"Test","tool":"claude","projectPath":"/tmp"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), ErrCodeForbidden) {
		t.Errorf("expected MUTATIONS_DISABLED error, got: %s", rr.Body.String())
	}
}

func TestSessionCreateMissingTitle(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	srv.mutator = &fakeMutator{}

	body := strings.NewReader(`{"tool":"claude","projectPath":"/tmp"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), ErrCodeBadRequest) {
		t.Errorf("expected INVALID_REQUEST error, got: %s", rr.Body.String())
	}
}

func TestSessionCreateMissingPath(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	srv.mutator = &fakeMutator{}

	body := strings.NewReader(`{"title":"Test","tool":"claude"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), ErrCodeBadRequest) {
		t.Errorf("expected INVALID_REQUEST error, got: %s", rr.Body.String())
	}
}

func TestSessionStopOK(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	srv.mutator = &fakeMutator{
		stopSessionFn: func(id string) error { return nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/test-id/stop", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestSessionDeleteOK(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	srv.mutator = &fakeMutator{
		deleteSessionFn: func(id string) error { return nil },
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/test-id", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestSessionStartOK(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	srv.mutator = &fakeMutator{
		startSessionFn: func(id string) error { return nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/test-id/start", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestSessionRestartOK(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	srv.mutator = &fakeMutator{
		restartSessionFn: func(id string) error { return nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/test-id/restart", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestSessionForkOK(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	srv.mutator = &fakeMutator{
		forkSessionFn: func(id string) (string, error) { return "forked-id", nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/test-id/fork", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "forked-id") {
		t.Errorf("expected forked session id in response, got: %s", rr.Body.String())
	}
}

func TestSessionsUnauthorized(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr: "127.0.0.1:0",
		Token:      "secret-token",
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), ErrCodeUnauthorized) {
		t.Errorf("expected UNAUTHORIZED error, got: %s", rr.Body.String())
	}
}

func TestMutationNilMutatorReturns503(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	// mutator is nil

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/test-id/stop", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d: %s", http.StatusServiceUnavailable, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), ErrCodeNotImplemented) {
		t.Errorf("expected NOT_IMPLEMENTED error, got: %s", rr.Body.String())
	}
}

// TestCreateSession_ForwardsGroupPath verifies that the groupPath field in the
// POST /api/sessions body is correctly passed through the handler to the
// SessionMutator. This is the server-side counterpart to the PR-1 JS change
// that wires the GROUP selector into the CreateSessionDialog form.
func TestCreateSession_ForwardsGroupPath(t *testing.T) {
	var capturedGroupPath string
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	srv.mutator = &fakeMutator{
		createSessionFn: func(title, tool, projectPath, groupPath string) (string, error) {
			capturedGroupPath = groupPath
			return "sess-grp-1", nil
		},
	}

	body := strings.NewReader(`{"title":"t","tool":"claude","projectPath":"/tmp","groupPath":"my-group"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	if capturedGroupPath != "my-group" {
		t.Errorf("expected groupPath %q, got %q", "my-group", capturedGroupPath)
	}
}

// TestCreateSession_OmittedGroupPathIsEmpty verifies that when groupPath is
// absent from the request body the mutator receives an empty string (not a
// 400 error), preserving the pre-PR-1 behaviour for sessions with no group.
func TestCreateSession_OmittedGroupPathIsEmpty(t *testing.T) {
	var capturedGroupPath = "UNSET" // sentinel; must be overwritten
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	srv.mutator = &fakeMutator{
		createSessionFn: func(title, tool, projectPath, groupPath string) (string, error) {
			capturedGroupPath = groupPath
			return "sess-nogrp-1", nil
		},
	}

	body := strings.NewReader(`{"title":"t","tool":"claude","projectPath":"/tmp"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	if capturedGroupPath != "" {
		t.Errorf("expected empty groupPath when omitted, got %q", capturedGroupPath)
	}
}

// TestCreateSession_AcceptsAllSevenBuiltInTools verifies that the HTTP handler
// accepts every tool name in the PR-1 expanded tool list without returning an
// error. The server does not validate tool names (custom tools are also valid),
// so each should yield 201 with the tool string forwarded verbatim.
func TestCreateSession_AcceptsAllSevenBuiltInTools(t *testing.T) {
	tools := []string{"claude", "codex", "gemini", "opencode", "copilot", "pi", "shell"}
	for _, tool := range tools {
		tool := tool // capture
		t.Run(tool, func(t *testing.T) {
			t.Parallel()
			var capturedTool string
			srv := NewServer(Config{
				ListenAddr:   "127.0.0.1:0",
				WebMutations: true,
			})
			srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
			srv.mutator = &fakeMutator{
				createSessionFn: func(title, tt, projectPath, groupPath string) (string, error) {
					capturedTool = tt
					return "sess-" + tt, nil
				},
			}

			body := strings.NewReader(fmt.Sprintf(
				`{"title":"t","tool":%q,"projectPath":"/tmp"}`, tool))
			req := httptest.NewRequest(http.MethodPost, "/api/sessions", body)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rr, req)

			if rr.Code != http.StatusCreated {
				t.Fatalf("tool %q: expected 201, got %d: %s", tool, rr.Code, rr.Body.String())
			}
			if capturedTool != tool {
				t.Errorf("tool %q: mutator received %q", tool, capturedTool)
			}
		})
	}
}

func TestMutationNotifiesSSE(t *testing.T) {
	srv := NewServer(Config{
		ListenAddr:   "127.0.0.1:0",
		WebMutations: true,
	})
	srv.menuData = &fakeMenuDataLoader{snapshot: &MenuSnapshot{}}
	srv.mutator = &fakeMutator{
		stopSessionFn: func(id string) error { return nil },
	}

	ch := srv.subscribeMenuChanges()
	defer srv.unsubscribeMenuChanges(ch)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions/test-id/stop", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	select {
	case <-ch:
		// notification received
	case <-time.After(250 * time.Millisecond):
		t.Error("expected SSE notification within 250ms, got none")
	}
}
