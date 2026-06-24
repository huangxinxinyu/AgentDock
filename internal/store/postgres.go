package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/huangxinxinyu/agentdock/internal/domain"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrNotFound = errors.New("resource not found")

type PostgresStore struct {
	db *sql.DB
}

func Open(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &PostgresStore{db: db}, nil
}

func (store *PostgresStore) DB() *sql.DB {
	return store.db
}

func (store *PostgresStore) Close() error {
	return store.db.Close()
}

type CreateWorkspaceParams struct {
	Name string
}

type CreateRepositoryParams struct {
	WorkspaceID string
	Name        string
	URL         string
}

type CreateAgentParams struct {
	WorkspaceID string
	Name        string
	RuntimeKey  string
}

type CreateIssueParams struct {
	WorkspaceID  string
	RepositoryID string
	AgentID      string
	Title        string
	Prompt       string
}

type CreateRunParams struct {
	IssueID        string
	IdempotencyKey string
}

type CreateSandboxParams struct {
	Name              string
	Provider          string
	ProviderSessionID string
	State             domain.SandboxState
	DefaultWorkdir    string
	AgentOSImage      string
	Metadata          string
	LastError         string
}

type CreateSandboxTaskParams struct {
	SandboxSessionID string
	Prompt           string
	Entrypoint       string
	Workdir          string
}

func (store *PostgresStore) CreateWorkspace(ctx context.Context, params CreateWorkspaceParams) (domain.Workspace, error) {
	workspace := domain.Workspace{ID: newID(), Name: strings.TrimSpace(params.Name)}
	err := store.db.QueryRowContext(ctx, `
		INSERT INTO workspaces(id, name)
		VALUES ($1, $2)
		RETURNING created_at
	`, workspace.ID, workspace.Name).Scan(&workspace.CreatedAt)
	return workspace, wrapSQLError(err)
}

func (store *PostgresStore) GetWorkspace(ctx context.Context, id string) (domain.Workspace, error) {
	var workspace domain.Workspace
	err := store.db.QueryRowContext(ctx, `
		SELECT id, name, created_at
		FROM workspaces
		WHERE id = $1
	`, id).Scan(&workspace.ID, &workspace.Name, &workspace.CreatedAt)
	return workspace, wrapSQLError(err)
}

func (store *PostgresStore) CreateRepository(ctx context.Context, params CreateRepositoryParams) (domain.Repository, error) {
	repository := domain.Repository{
		ID:          newID(),
		WorkspaceID: params.WorkspaceID,
		Name:        strings.TrimSpace(params.Name),
		URL:         strings.TrimSpace(params.URL),
	}
	err := store.db.QueryRowContext(ctx, `
		INSERT INTO repositories(id, workspace_id, name, url)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`, repository.ID, repository.WorkspaceID, repository.Name, repository.URL).Scan(&repository.CreatedAt)
	return repository, wrapSQLError(err)
}

func (store *PostgresStore) CreateAgent(ctx context.Context, params CreateAgentParams) (domain.Agent, error) {
	agent := domain.Agent{
		ID:          newID(),
		WorkspaceID: params.WorkspaceID,
		Name:        strings.TrimSpace(params.Name),
		RuntimeKey:  strings.TrimSpace(params.RuntimeKey),
	}
	err := store.db.QueryRowContext(ctx, `
		INSERT INTO agents(id, workspace_id, name, runtime_key)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`, agent.ID, agent.WorkspaceID, agent.Name, agent.RuntimeKey).Scan(&agent.CreatedAt)
	return agent, wrapSQLError(err)
}

func (store *PostgresStore) CreateIssue(ctx context.Context, params CreateIssueParams) (domain.Issue, error) {
	issue := domain.Issue{
		ID:           newID(),
		WorkspaceID:  params.WorkspaceID,
		RepositoryID: params.RepositoryID,
		AgentID:      params.AgentID,
		Title:        strings.TrimSpace(params.Title),
		Prompt:       strings.TrimSpace(params.Prompt),
		Status:       domain.IssueStatusOpen,
	}
	err := store.db.QueryRowContext(ctx, `
		INSERT INTO issues(id, workspace_id, repository_id, agent_id, title, prompt, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`, issue.ID, issue.WorkspaceID, issue.RepositoryID, issue.AgentID, issue.Title, issue.Prompt, issue.Status).Scan(&issue.CreatedAt, &issue.UpdatedAt)
	return issue, wrapSQLError(err)
}

func (store *PostgresStore) GetIssue(ctx context.Context, id string) (domain.Issue, error) {
	var issue domain.Issue
	err := store.db.QueryRowContext(ctx, `
		SELECT id, workspace_id, repository_id, agent_id, title, prompt, status, created_at, updated_at
		FROM issues
		WHERE id = $1
	`, id).Scan(&issue.ID, &issue.WorkspaceID, &issue.RepositoryID, &issue.AgentID, &issue.Title, &issue.Prompt, &issue.Status, &issue.CreatedAt, &issue.UpdatedAt)
	return issue, wrapSQLError(err)
}

func (store *PostgresStore) CreateRun(ctx context.Context, params CreateRunParams) (domain.Run, error) {
	idempotencyKey := strings.TrimSpace(params.IdempotencyKey)
	if idempotencyKey != "" {
		existing, err := store.getRunByIdempotencyKey(ctx, params.IssueID, idempotencyKey)
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, ErrNotFound) {
			return domain.Run{}, err
		}
	}

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Run{}, fmt.Errorf("begin create run: %w", err)
	}
	defer rollbackUnlessCommitted(tx)

	var issue domain.Issue
	err = tx.QueryRowContext(ctx, `
		SELECT id, repository_id, agent_id, prompt
		FROM issues
		WHERE id = $1
		FOR SHARE
	`, params.IssueID).Scan(&issue.ID, &issue.RepositoryID, &issue.AgentID, &issue.Prompt)
	if err != nil {
		return domain.Run{}, wrapSQLError(err)
	}

	run := domain.Run{
		ID:             newID(),
		IssueID:        issue.ID,
		RepositoryID:   issue.RepositoryID,
		AgentID:        issue.AgentID,
		Prompt:         issue.Prompt,
		State:          domain.RunStateQueued,
		IdempotencyKey: idempotencyKey,
	}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO runs(id, issue_id, repository_id, agent_id, prompt, state, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''))
		RETURNING created_at, updated_at, last_transition_at
	`, run.ID, run.IssueID, run.RepositoryID, run.AgentID, run.Prompt, run.State, run.IdempotencyKey).Scan(&run.CreatedAt, &run.UpdatedAt, &run.LastTransitionAt)
	if err != nil {
		_ = tx.Rollback()
		if idempotencyKey != "" {
			existing, lookupErr := store.getRunByIdempotencyKey(ctx, params.IssueID, idempotencyKey)
			if lookupErr == nil {
				return existing, nil
			}
		}
		return domain.Run{}, wrapSQLError(err)
	}
	if _, err := appendRunEventTx(ctx, tx, run.ID, domain.RunEventQueued, "run queued"); err != nil {
		return domain.Run{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Run{}, fmt.Errorf("commit create run: %w", err)
	}
	return run, nil
}

func (store *PostgresStore) GetRun(ctx context.Context, id string) (domain.Run, error) {
	return store.scanRun(store.db.QueryRowContext(ctx, runSelectSQL()+` WHERE id = $1`, id))
}

func (store *PostgresStore) ListRunEvents(ctx context.Context, runID string) ([]domain.RunEvent, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, run_id, sequence, type, message, created_at
		FROM run_events
		WHERE run_id = $1
		ORDER BY sequence
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("list run events: %w", err)
	}
	defer rows.Close()

	var events []domain.RunEvent
	for rows.Next() {
		var event domain.RunEvent
		if err := rows.Scan(&event.ID, &event.RunID, &event.Sequence, &event.Type, &event.Message, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan run event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate run events: %w", err)
	}
	return events, nil
}

func (store *PostgresStore) ClaimQueuedRun(ctx context.Context) (domain.Run, bool, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Run{}, false, fmt.Errorf("begin claim run: %w", err)
	}
	defer rollbackUnlessCommitted(tx)

	var id string
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM runs
		WHERE state = $1
		ORDER BY created_at
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, domain.RunStateQueued).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		if err := tx.Commit(); err != nil {
			return domain.Run{}, false, fmt.Errorf("commit empty claim: %w", err)
		}
		return domain.Run{}, false, nil
	}
	if err != nil {
		return domain.Run{}, false, fmt.Errorf("select queued run: %w", err)
	}
	run, err := advanceRunTx(ctx, tx, id, domain.RunStateQueued, domain.RunStateProvisioning)
	if err != nil {
		return domain.Run{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Run{}, false, fmt.Errorf("commit claim run: %w", err)
	}
	return run, true, nil
}

func (store *PostgresStore) AdvanceRun(ctx context.Context, runID string, from domain.RunState, to domain.RunState) (domain.Run, error) {
	if err := domain.ValidateRunTransition(from, to); err != nil {
		return domain.Run{}, err
	}
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Run{}, fmt.Errorf("begin advance run: %w", err)
	}
	defer rollbackUnlessCommitted(tx)
	run, err := advanceRunTx(ctx, tx, runID, from, to)
	if err != nil {
		return domain.Run{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Run{}, fmt.Errorf("commit advance run: %w", err)
	}
	return run, nil
}

func (store *PostgresStore) CompleteRun(ctx context.Context, runID string, summary string) (domain.Run, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Run{}, fmt.Errorf("begin complete run: %w", err)
	}
	defer rollbackUnlessCommitted(tx)
	run, err := scanRun(tx.QueryRowContext(ctx, `
		UPDATE runs
		SET state = $2,
			result_summary = $3,
			updated_at = now(),
			completed_at = now(),
			last_transition_at = now()
		WHERE id = $1 AND state = $4
		RETURNING id, issue_id, repository_id, agent_id, prompt, state, result_summary, COALESCE(idempotency_key, ''), created_at, updated_at, COALESCE(started_at, '0001-01-01'::timestamptz), COALESCE(completed_at, '0001-01-01'::timestamptz), last_transition_at
	`, runID, domain.RunStateCompleted, summary, domain.RunStateRunning))
	if err != nil {
		return domain.Run{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Run{}, fmt.Errorf("commit complete run: %w", err)
	}
	return run, nil
}

func (store *PostgresStore) AppendRunEvent(ctx context.Context, runID string, eventType domain.RunEventType, message string) (domain.RunEvent, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.RunEvent{}, fmt.Errorf("begin append event: %w", err)
	}
	defer rollbackUnlessCommitted(tx)
	event, err := appendRunEventTx(ctx, tx, runID, eventType, message)
	if err != nil {
		return domain.RunEvent{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.RunEvent{}, fmt.Errorf("commit append event: %w", err)
	}
	return event, nil
}

func (store *PostgresStore) RecordSandboxSession(ctx context.Context, params domain.RecordSandboxSessionParams) (domain.SandboxSession, error) {
	return store.CreateSandbox(ctx, CreateSandboxParams{
		Name:              params.Name,
		Provider:          params.Provider,
		ProviderSessionID: params.ProviderSessionID,
		State:             params.State,
		DefaultWorkdir:    params.DefaultWorkdir,
		AgentOSImage:      params.AgentOSImage,
		Metadata:          params.Metadata,
		LastError:         params.LastError,
	})
}

func (store *PostgresStore) CreateSandbox(ctx context.Context, params CreateSandboxParams) (domain.SandboxSession, error) {
	session := domain.SandboxSession{
		ID:                newID(),
		Name:              strings.TrimSpace(params.Name),
		Provider:          strings.TrimSpace(params.Provider),
		ProviderSessionID: strings.TrimSpace(params.ProviderSessionID),
		State:             params.State,
		DefaultWorkdir:    strings.TrimSpace(params.DefaultWorkdir),
		AgentOSImage:      strings.TrimSpace(params.AgentOSImage),
		Metadata:          strings.TrimSpace(params.Metadata),
		LastError:         strings.TrimSpace(params.LastError),
	}
	if session.Provider == "" {
		session.Provider = "noop"
	}
	if session.State == "" {
		session.State = domain.SandboxStateReady
	}
	if session.DefaultWorkdir == "" {
		session.DefaultWorkdir = "/workspace"
	}
	if session.Metadata == "" {
		session.Metadata = "{}"
	}
	err := store.db.QueryRowContext(ctx, `
		INSERT INTO sandbox_sessions(id, name, provider, provider_session_id, state, default_workdir, agentos_image, metadata, last_error)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, NULLIF($7, ''), $8::jsonb, $9)
		RETURNING created_at, updated_at, COALESCE(last_started_at, '0001-01-01'::timestamptz), COALESCE(last_paused_at, '0001-01-01'::timestamptz), COALESCE(closed_at, '0001-01-01'::timestamptz)
	`, session.ID, session.Name, session.Provider, session.ProviderSessionID, session.State, session.DefaultWorkdir, session.AgentOSImage, session.Metadata, session.LastError).Scan(&session.CreatedAt, &session.UpdatedAt, &session.LastStartedAt, &session.LastPausedAt, &session.ClosedAt)
	return session, wrapSQLError(err)
}

func (store *PostgresStore) ListSandboxes(ctx context.Context) ([]domain.SandboxSession, error) {
	rows, err := store.db.QueryContext(ctx, sandboxSelectSQL()+` ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list sandboxes: %w", err)
	}
	defer rows.Close()

	var sessions []domain.SandboxSession
	for rows.Next() {
		session, err := scanSandbox(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sandboxes: %w", err)
	}
	return sessions, nil
}

func (store *PostgresStore) GetSandbox(ctx context.Context, id string) (domain.SandboxSession, error) {
	return scanSandboxRow(store.db.QueryRowContext(ctx, sandboxSelectSQL()+` WHERE id = $1`, id))
}

func (store *PostgresStore) UpdateSandboxState(ctx context.Context, id string, from domain.SandboxState, to domain.SandboxState, lastError string) (domain.SandboxSession, error) {
	if err := domain.ValidateSandboxTransition(from, to); err != nil {
		return domain.SandboxSession{}, err
	}
	pausedExpr := "last_paused_at"
	if to == domain.SandboxStatePaused {
		pausedExpr = "now()"
	}
	closedExpr := "closed_at"
	if to == domain.SandboxStateClosed {
		closedExpr = "now()"
	}
	return scanSandboxRow(store.db.QueryRowContext(ctx, fmt.Sprintf(`
		UPDATE sandbox_sessions
		SET state = $2,
			last_error = $3,
			updated_at = now(),
			last_paused_at = %s,
			closed_at = %s
		WHERE id = $1 AND state = $4
		RETURNING id, name, provider, COALESCE(provider_session_id, ''), state, default_workdir, COALESCE(agentos_image, ''), metadata::text, last_error, created_at, updated_at, COALESCE(last_started_at, '0001-01-01'::timestamptz), COALESCE(last_paused_at, '0001-01-01'::timestamptz), COALESCE(closed_at, '0001-01-01'::timestamptz)
	`, pausedExpr, closedExpr), id, to, lastError, from))
}

func (store *PostgresStore) CreateSandboxTask(ctx context.Context, params CreateSandboxTaskParams) (domain.SandboxTask, error) {
	task := domain.SandboxTask{
		ID:               newID(),
		SandboxSessionID: params.SandboxSessionID,
		Prompt:           strings.TrimSpace(params.Prompt),
		State:            domain.SandboxTaskStateQueued,
		Entrypoint:       strings.TrimSpace(params.Entrypoint),
		Workdir:          strings.TrimSpace(params.Workdir),
	}
	if task.Entrypoint == "" {
		task.Entrypoint = "agentos-sdk-python"
	}
	if task.Workdir == "" {
		task.Workdir = "/workspace"
	}
	err := store.db.QueryRowContext(ctx, `
		INSERT INTO sandbox_tasks(id, sandbox_session_id, prompt, state, entrypoint, workdir)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at, COALESCE(started_at, '0001-01-01'::timestamptz), COALESCE(completed_at, '0001-01-01'::timestamptz)
	`, task.ID, task.SandboxSessionID, task.Prompt, task.State, task.Entrypoint, task.Workdir).Scan(&task.CreatedAt, &task.UpdatedAt, &task.StartedAt, &task.CompletedAt)
	return task, wrapSQLError(err)
}

func (store *PostgresStore) ListSandboxTasks(ctx context.Context, sandboxID string) ([]domain.SandboxTask, error) {
	rows, err := store.db.QueryContext(ctx, sandboxTaskSelectSQL()+` WHERE sandbox_session_id = $1 ORDER BY created_at DESC`, sandboxID)
	if err != nil {
		return nil, fmt.Errorf("list sandbox tasks: %w", err)
	}
	defer rows.Close()
	var tasks []domain.SandboxTask
	for rows.Next() {
		task, err := scanSandboxTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sandbox tasks: %w", err)
	}
	return tasks, nil
}

func (store *PostgresStore) GetSandboxTask(ctx context.Context, id string) (domain.SandboxTask, error) {
	return scanSandboxTaskRow(store.db.QueryRowContext(ctx, sandboxTaskSelectSQL()+` WHERE id = $1`, id))
}

func (store *PostgresStore) UpdateSandboxTaskState(ctx context.Context, id string, from domain.SandboxTaskState, to domain.SandboxTaskState, summary string, outputRef string, lastError string) (domain.SandboxTask, error) {
	if err := domain.ValidateSandboxTaskTransition(from, to); err != nil {
		return domain.SandboxTask{}, err
	}
	startedExpr := "started_at"
	if to == domain.SandboxTaskStateRunning {
		startedExpr = "COALESCE(started_at, now())"
	}
	completedExpr := "completed_at"
	if to == domain.SandboxTaskStateSucceeded || to == domain.SandboxTaskStateFailed || to == domain.SandboxTaskStateCancelled {
		completedExpr = "now()"
	}
	return scanSandboxTaskRow(store.db.QueryRowContext(ctx, fmt.Sprintf(`
		UPDATE sandbox_tasks
		SET state = $2,
			summary = CASE WHEN $3 = '' THEN summary ELSE $3 END,
			output_ref = CASE WHEN $4 = '' THEN output_ref ELSE $4 END,
			last_error = $5,
			updated_at = now(),
			started_at = %s,
			completed_at = %s
		WHERE id = $1 AND state = $6
		RETURNING id, sandbox_session_id, prompt, state, entrypoint, workdir, summary, output_ref, last_error, created_at, updated_at, COALESCE(started_at, '0001-01-01'::timestamptz), COALESCE(completed_at, '0001-01-01'::timestamptz)
	`, startedExpr, completedExpr), id, to, summary, outputRef, lastError, from))
}

func (store *PostgresStore) AppendSandboxTaskEvent(ctx context.Context, taskID string, eventType domain.SandboxTaskEventType, message string, payload string) (domain.SandboxTaskEvent, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.SandboxTaskEvent{}, fmt.Errorf("begin append sandbox task event: %w", err)
	}
	defer rollbackUnlessCommitted(tx)
	event, err := appendSandboxTaskEventTx(ctx, tx, taskID, eventType, message, payload)
	if err != nil {
		return domain.SandboxTaskEvent{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.SandboxTaskEvent{}, fmt.Errorf("commit append sandbox task event: %w", err)
	}
	return event, nil
}

func (store *PostgresStore) ListSandboxTaskEvents(ctx context.Context, taskID string) ([]domain.SandboxTaskEvent, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, sandbox_task_id, sequence, type, message, payload::text, created_at
		FROM sandbox_task_events
		WHERE sandbox_task_id = $1
		ORDER BY sequence
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("list sandbox task events: %w", err)
	}
	defer rows.Close()
	var events []domain.SandboxTaskEvent
	for rows.Next() {
		var event domain.SandboxTaskEvent
		if err := rows.Scan(&event.ID, &event.SandboxTaskID, &event.Sequence, &event.Type, &event.Message, &event.Payload, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan sandbox task event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sandbox task events: %w", err)
	}
	return events, nil
}

func (store *PostgresStore) TruncateForTest(ctx context.Context) error {
	_, err := store.db.ExecContext(ctx, `
		TRUNCATE sandbox_task_events, sandbox_tasks, sandbox_sessions, run_events, runs, issues, agents, repositories, workspaces
		RESTART IDENTITY CASCADE
	`)
	return err
}

func (store *PostgresStore) getRunByIdempotencyKey(ctx context.Context, issueID string, key string) (domain.Run, error) {
	return store.scanRun(store.db.QueryRowContext(ctx, runSelectSQL()+` WHERE issue_id = $1 AND idempotency_key = $2`, issueID, key))
}

func (store *PostgresStore) scanRun(row *sql.Row) (domain.Run, error) {
	return scanRun(row)
}

func scanRun(row *sql.Row) (domain.Run, error) {
	var run domain.Run
	err := row.Scan(
		&run.ID,
		&run.IssueID,
		&run.RepositoryID,
		&run.AgentID,
		&run.Prompt,
		&run.State,
		&run.ResultSummary,
		&run.IdempotencyKey,
		&run.CreatedAt,
		&run.UpdatedAt,
		&run.StartedAt,
		&run.CompletedAt,
		&run.LastTransitionAt,
	)
	return run, wrapSQLError(err)
}

func runSelectSQL() string {
	return `
		SELECT id, issue_id, repository_id, agent_id, prompt, state, result_summary, COALESCE(idempotency_key, ''), created_at, updated_at, COALESCE(started_at, '0001-01-01'::timestamptz), COALESCE(completed_at, '0001-01-01'::timestamptz), last_transition_at
		FROM runs
	`
}

func advanceRunTx(ctx context.Context, tx *sql.Tx, runID string, from domain.RunState, to domain.RunState) (domain.Run, error) {
	if err := domain.ValidateRunTransition(from, to); err != nil {
		return domain.Run{}, err
	}
	startedExpr := "started_at"
	if to == domain.RunStateRunning {
		startedExpr = "COALESCE(started_at, now())"
	}
	return scanRun(tx.QueryRowContext(ctx, fmt.Sprintf(`
		UPDATE runs
		SET state = $2,
			updated_at = now(),
			started_at = %s,
			last_transition_at = now()
		WHERE id = $1 AND state = $3
		RETURNING id, issue_id, repository_id, agent_id, prompt, state, result_summary, COALESCE(idempotency_key, ''), created_at, updated_at, COALESCE(started_at, '0001-01-01'::timestamptz), COALESCE(completed_at, '0001-01-01'::timestamptz), last_transition_at
	`, startedExpr), runID, to, from))
}

func appendRunEventTx(ctx context.Context, tx *sql.Tx, runID string, eventType domain.RunEventType, message string) (domain.RunEvent, error) {
	if _, err := tx.ExecContext(ctx, `SELECT id FROM runs WHERE id = $1 FOR UPDATE`, runID); err != nil {
		return domain.RunEvent{}, fmt.Errorf("lock run for event: %w", err)
	}
	event := domain.RunEvent{ID: newID(), RunID: runID, Type: eventType, Message: message}
	err := tx.QueryRowContext(ctx, `
		INSERT INTO run_events(id, run_id, sequence, type, message)
		VALUES (
			$1,
			$2,
			(SELECT COALESCE(MAX(sequence), 0) + 1 FROM run_events WHERE run_id = $2),
			$3,
			$4
		)
		RETURNING sequence, created_at
	`, event.ID, event.RunID, event.Type, event.Message).Scan(&event.Sequence, &event.CreatedAt)
	if err != nil {
		return domain.RunEvent{}, fmt.Errorf("append run event: %w", err)
	}
	return event, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSandboxRow(row rowScanner) (domain.SandboxSession, error) {
	var session domain.SandboxSession
	err := row.Scan(
		&session.ID,
		&session.Name,
		&session.Provider,
		&session.ProviderSessionID,
		&session.State,
		&session.DefaultWorkdir,
		&session.AgentOSImage,
		&session.Metadata,
		&session.LastError,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.LastStartedAt,
		&session.LastPausedAt,
		&session.ClosedAt,
	)
	return session, wrapSQLError(err)
}

func scanSandbox(row rowScanner) (domain.SandboxSession, error) {
	session, err := scanSandboxRow(row)
	if err != nil {
		return domain.SandboxSession{}, fmt.Errorf("scan sandbox: %w", err)
	}
	return session, nil
}

func sandboxSelectSQL() string {
	return `
		SELECT id, name, provider, COALESCE(provider_session_id, ''), state, default_workdir, COALESCE(agentos_image, ''), metadata::text, last_error, created_at, updated_at, COALESCE(last_started_at, '0001-01-01'::timestamptz), COALESCE(last_paused_at, '0001-01-01'::timestamptz), COALESCE(closed_at, '0001-01-01'::timestamptz)
		FROM sandbox_sessions
	`
}

func sandboxTaskSelectSQL() string {
	return `
		SELECT id, sandbox_session_id, prompt, state, entrypoint, workdir, summary, output_ref, last_error, created_at, updated_at, COALESCE(started_at, '0001-01-01'::timestamptz), COALESCE(completed_at, '0001-01-01'::timestamptz)
		FROM sandbox_tasks
	`
}

func scanSandboxTaskRow(row rowScanner) (domain.SandboxTask, error) {
	var task domain.SandboxTask
	err := row.Scan(
		&task.ID,
		&task.SandboxSessionID,
		&task.Prompt,
		&task.State,
		&task.Entrypoint,
		&task.Workdir,
		&task.Summary,
		&task.OutputRef,
		&task.LastError,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.StartedAt,
		&task.CompletedAt,
	)
	return task, wrapSQLError(err)
}

func scanSandboxTask(row rowScanner) (domain.SandboxTask, error) {
	task, err := scanSandboxTaskRow(row)
	if err != nil {
		return domain.SandboxTask{}, fmt.Errorf("scan sandbox task: %w", err)
	}
	return task, nil
}

func appendSandboxTaskEventTx(ctx context.Context, tx *sql.Tx, taskID string, eventType domain.SandboxTaskEventType, message string, payload string) (domain.SandboxTaskEvent, error) {
	if _, err := tx.ExecContext(ctx, `SELECT id FROM sandbox_tasks WHERE id = $1 FOR UPDATE`, taskID); err != nil {
		return domain.SandboxTaskEvent{}, fmt.Errorf("lock sandbox task for event: %w", err)
	}
	if strings.TrimSpace(payload) == "" {
		payload = "{}"
	}
	event := domain.SandboxTaskEvent{ID: newID(), SandboxTaskID: taskID, Type: eventType, Message: message, Payload: payload}
	err := tx.QueryRowContext(ctx, `
		INSERT INTO sandbox_task_events(id, sandbox_task_id, sequence, type, message, payload)
		VALUES (
			$1,
			$2,
			(SELECT COALESCE(MAX(sequence), 0) + 1 FROM sandbox_task_events WHERE sandbox_task_id = $2),
			$3,
			$4,
			$5::jsonb
		)
		RETURNING sequence, created_at
	`, event.ID, event.SandboxTaskID, event.Type, event.Message, event.Payload).Scan(&event.Sequence, &event.CreatedAt)
	if err != nil {
		return domain.SandboxTaskEvent{}, fmt.Errorf("append sandbox task event: %w", err)
	}
	return event, nil
}

func wrapSQLError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func rollbackUnlessCommitted(tx *sql.Tx) {
	_ = tx.Rollback()
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic("generate uuid: " + err.Error())
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(bytes[:])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
}
