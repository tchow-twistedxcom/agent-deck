package beads

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const (
	BeadsDir      = ".beads"
	IssuesFile    = "issues.jsonl"
)

// Reader reads beads from a project's .beads directory
type Reader struct {
	projectPath string
}

// NewReader creates a new beads reader for the given project path
func NewReader(projectPath string) *Reader {
	return &Reader{projectPath: projectPath}
}

// BeadsPath returns the path to the .beads directory
func (r *Reader) BeadsPath() string {
	return filepath.Join(r.projectPath, BeadsDir)
}

// IssuesPath returns the path to the issues.jsonl file
func (r *Reader) IssuesPath() string {
	return filepath.Join(r.BeadsPath(), IssuesFile)
}

// HasBeads returns true if the project has a .beads directory
func (r *Reader) HasBeads() bool {
	_, err := os.Stat(r.BeadsPath())
	return err == nil
}

// ReadAll reads all beads from the project
func (r *Reader) ReadAll() ([]*Bead, error) {
	path := r.IssuesPath()
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No beads yet
		}
		return nil, fmt.Errorf("failed to open beads file: %w", err)
	}
	defer file.Close()

	var beads []*Bead
	scanner := bufio.NewScanner(file)
	// Increase buffer size for large JSONL lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var bead Bead
		if err := json.Unmarshal(line, &bead); err != nil {
			return nil, fmt.Errorf("failed to parse bead at line %d: %w", lineNum, err)
		}
		beads = append(beads, &bead)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading beads file: %w", err)
	}

	return beads, nil
}

// ReadOpen reads only open beads
func (r *Reader) ReadOpen() ([]*Bead, error) {
	all, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var open []*Bead
	for _, b := range all {
		if b.Status == StatusOpen || b.Status == StatusInProgress {
			open = append(open, b)
		}
	}
	return open, nil
}

// ReadReady reads beads that are ready to work on (no blockers)
func (r *Reader) ReadReady() ([]*Bead, error) {
	all, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	// Build a set of closed bead IDs
	closedIDs := make(map[string]bool)
	for _, b := range all {
		if b.Status == StatusClosed {
			closedIDs[b.ID] = true
		}
	}

	var ready []*Bead
	for _, b := range all {
		if b.Status == StatusClosed {
			continue
		}

		// Check if all blockers are closed
		blocked := false
		for _, dep := range b.Dependencies {
			if !closedIDs[dep.DependsOnID] {
				blocked = true
				break
			}
		}

		if !blocked {
			ready = append(ready, b)
		}
	}

	// Sort by priority
	sort.Slice(ready, func(i, j int) bool {
		return ready[i].Priority < ready[j].Priority
	})

	return ready, nil
}

// GetByID returns a bead by its ID
func (r *Reader) GetByID(id string) (*Bead, error) {
	all, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, b := range all {
		if b.ID == id {
			return b, nil
		}
	}
	return nil, fmt.Errorf("bead not found: %s", id)
}

// ReadByProject reads beads grouped by their project path
// This is useful when aggregating beads from multiple projects
func ReadByProject(projectPaths []string) (map[string][]*Bead, error) {
	result := make(map[string][]*Bead)

	for _, path := range projectPaths {
		reader := NewReader(path)
		if !reader.HasBeads() {
			continue
		}

		beads, err := reader.ReadOpen()
		if err != nil {
			return nil, fmt.Errorf("failed to read beads from %s: %w", path, err)
		}

		if len(beads) > 0 {
			result[path] = beads
		}
	}

	return result, nil
}
