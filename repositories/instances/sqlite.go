package instances

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/verbeux-ai/whatsmiau/interfaces"
	"github.com/verbeux-ai/whatsmiau/models"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// These verify if SQLiteInstance follows instances interface pattern
var _ interfaces.InstanceRepository = (*SQLiteInstance)(nil)

type SQLiteInstance struct {
	db     *sql.DB
	dbPath string
	mu     sync.Mutex
}

func NewSQLite(dbURL string) (*SQLiteInstance, error) {
	// SQLite connection string can be "file:data.db?_foreign_keys=on" or just "data.db"
	// We'll use it as-is since the sqlite3 driver handles both formats
	db, err := sql.Open("sqlite3", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Enable foreign keys and set connection pool settings
	db.SetMaxOpenConns(1) // SQLite works best with a single connection
	db.SetMaxIdleConns(1)

	instance := &SQLiteInstance{
		db:     db,
		dbPath: dbURL,
	}

	if err := instance.initTable(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize table: %w", err)
	}

	return instance, nil
}

func (s *SQLiteInstance) initTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS instances (
		id TEXT PRIMARY KEY,
		data TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_instances_id ON instances(id);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *SQLiteInstance) Create(ctx context.Context, instance *models.Instance) error {
	if instance.ID == "" {
		return ErrInstanceIDEmpty
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if instance already exists
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM instances WHERE id = ?", instance.ID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check instance existence: %w", err)
	}
	if count > 0 {
		return ErrorAlreadyExists
	}

	// Serialize instance to JSON
	data, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal instance: %w", err)
	}

	// Insert instance
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO instances (id, data, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		instance.ID, string(data))
	if err != nil {
		return fmt.Errorf("failed to insert instance: %w", err)
	}

	return nil
}

func (s *SQLiteInstance) Update(ctx context.Context, id string, toUpdate *models.Instance) (*models.Instance, error) {
	if id == "" {
		return nil, ErrInstanceIDEmpty
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get existing instance
	var dataStr string
	err := s.db.QueryRowContext(ctx, "SELECT data FROM instances WHERE id = ?", id).Scan(&dataStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorNotFound
		}
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	// Deserialize existing instance
	var oldInstance models.Instance
	if err := json.Unmarshal([]byte(dataStr), &oldInstance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal instance: %w", err)
	}

	// Update fields
	if len(toUpdate.RemoteJID) > 0 {
		oldInstance.RemoteJID = toUpdate.RemoteJID
	}
	if toUpdate.Webhook.Url != "" {
		oldInstance.Webhook.Url = toUpdate.Webhook.Url
	}
	if toUpdate.Webhook.ByEvents != nil {
		oldInstance.Webhook.ByEvents = toUpdate.Webhook.ByEvents
	}
	if toUpdate.Webhook.Base64 != nil {
		oldInstance.Webhook.Base64 = toUpdate.Webhook.Base64
	}
	if toUpdate.Webhook.Headers != nil {
		if oldInstance.Webhook.Headers == nil {
			oldInstance.Webhook.Headers = make(map[string]string)
		}
		for k, v := range toUpdate.Webhook.Headers {
			oldInstance.Webhook.Headers[k] = v
		}
	}
	if toUpdate.Webhook.Events != nil && len(toUpdate.Webhook.Events) > 0 {
		oldInstance.Webhook.Events = toUpdate.Webhook.Events
	}

	// Serialize updated instance
	updatedData, err := json.Marshal(oldInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated instance: %w", err)
	}

	// Update in database
	_, err = s.db.ExecContext(ctx,
		"UPDATE instances SET data = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(updatedData), id)
	if err != nil {
		return nil, fmt.Errorf("failed to update instance: %w", err)
	}

	return &oldInstance, nil
}

func (s *SQLiteInstance) List(ctx context.Context, id string) ([]models.Instance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var rows *sql.Rows
	var err error

	if id != "" {
		rows, err = s.db.QueryContext(ctx, "SELECT data FROM instances WHERE id = ?", id)
	} else {
		rows, err = s.db.QueryContext(ctx, "SELECT data FROM instances")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query instances: %w", err)
	}
	defer rows.Close()

	var instances []models.Instance
	for rows.Next() {
		var dataStr string
		if err := rows.Scan(&dataStr); err != nil {
			zap.L().Warn("failed to scan instance data", zap.Error(err))
			continue
		}

		var instance models.Instance
		if err := json.Unmarshal([]byte(dataStr), &instance); err != nil {
			zap.L().Warn("failed to unmarshal instance data", zap.Error(err))
			continue
		}

		instances = append(instances, instance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating instances: %w", err)
	}

	return instances, nil
}

func (s *SQLiteInstance) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInstanceIDEmpty
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.db.ExecContext(ctx, "DELETE FROM instances WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrorNotFound
	}

	return nil
}

func (s *SQLiteInstance) Close() error {
	return s.db.Close()
}

