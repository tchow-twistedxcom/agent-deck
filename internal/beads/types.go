package beads

import (
	"time"
)

// Bead represents an issue/task in the beads system
type Bead struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Status      string       `json:"status"` // open, in_progress, closed
	Priority    int          `json:"priority"`
	IssueType   string       `json:"issue_type"` // epic, task, subtask
	Owner       string       `json:"owner,omitempty"`
	AssignedTo  string       `json:"assigned_to,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	CreatedBy   string       `json:"created_by,omitempty"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Dependencies []Dependency `json:"dependencies,omitempty"`
}

// Dependency represents a relationship between beads
type Dependency struct {
	IssueID     string    `json:"issue_id"`
	DependsOnID string    `json:"depends_on_id"`
	Type        string    `json:"type"` // parent-child, blocks, relates_to
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by,omitempty"`
}

// BeadStatus constants
const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusClosed     = "closed"
)

// BeadType constants
const (
	TypeEpic    = "epic"
	TypeTask    = "task"
	TypeSubtask = "subtask"
)

// IsBlocked returns true if this bead has blocking dependencies
func (b *Bead) IsBlocked() bool {
	for _, dep := range b.Dependencies {
		if dep.Type == "blocks" || dep.Type == "parent-child" {
			return true
		}
	}
	return false
}

// GetBlockers returns the IDs of beads that block this one
func (b *Bead) GetBlockers() []string {
	var blockers []string
	for _, dep := range b.Dependencies {
		if dep.Type == "blocks" || dep.Type == "parent-child" {
			blockers = append(blockers, dep.DependsOnID)
		}
	}
	return blockers
}

// PriorityLabel returns a human-readable priority label
func (b *Bead) PriorityLabel() string {
	switch b.Priority {
	case 0:
		return "P0"
	case 1:
		return "P1"
	case 2:
		return "P2"
	case 3:
		return "P3"
	default:
		return "P?"
	}
}

// StatusIcon returns a status indicator icon
func (b *Bead) StatusIcon() string {
	switch b.Status {
	case StatusOpen:
		return "○"
	case StatusInProgress:
		return "◐"
	case StatusClosed:
		return "●"
	default:
		return "?"
	}
}
