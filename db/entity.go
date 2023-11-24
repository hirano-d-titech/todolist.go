package db

// schema.go provides data models in DB
import (
	"time"
)

// Task corresponds to a row in `tasks` table
type Task struct {
	ID        uint64    `db:"id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	IsDone    bool      `db:"is_done"`
}

type User struct {
	ID       uint64 `db:"id"`
	Name     string `db:"name"`
	Password []byte `db:"password"`
	Deleted  bool   `db:"is_deleted"`
}

type Ownership struct {
	UserID uint64 `db:"user_id"`
	TaskID uint64 `db:"task_id"`
}
