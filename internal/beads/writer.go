package beads

import (
	"fmt"
	"os/exec"
	"strings"
)

// Writer executes bd commands to modify beads
type Writer struct {
	projectPath string
}

// NewWriter creates a new beads writer for the given project path
func NewWriter(projectPath string) *Writer {
	return &Writer{projectPath: projectPath}
}

// runBD executes a bd command in the project directory
func (w *Writer) runBD(args ...string) error {
	cmd := exec.Command("bd", args...)
	cmd.Dir = w.projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bd %s failed: %w\nOutput: %s", strings.Join(args, " "), err, output)
	}
	return nil
}

// runBDOutput executes a bd command and returns its output
func (w *Writer) runBDOutput(args ...string) (string, error) {
	cmd := exec.Command("bd", args...)
	cmd.Dir = w.projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("bd %s failed: %w\nOutput: %s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

// Create creates a new bead and returns its ID
func (w *Writer) Create(title string, priority int) (string, error) {
	output, err := w.runBDOutput("create", title, "-p", fmt.Sprintf("%d", priority))
	if err != nil {
		return "", err
	}

	// Parse the ID from output like "✓ Created issue: claude-deck-abc"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Created issue:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[len(parts)-1]), nil
			}
		}
	}

	return "", fmt.Errorf("could not parse bead ID from output: %s", output)
}

// CreateWithType creates a new bead with a specific type
func (w *Writer) CreateWithType(title string, priority int, issueType string) (string, error) {
	output, err := w.runBDOutput("create", title, "-p", fmt.Sprintf("%d", priority), "-t", issueType)
	if err != nil {
		return "", err
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Created issue:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[len(parts)-1]), nil
			}
		}
	}

	return "", fmt.Errorf("could not parse bead ID from output: %s", output)
}

// Claim marks a bead as in_progress and assigns it
func (w *Writer) Claim(id string) error {
	return w.runBD("update", id, "--claim")
}

// Close marks a bead as closed
func (w *Writer) Close(id string) error {
	return w.runBD("close", id)
}

// Reopen reopens a closed bead
func (w *Writer) Reopen(id string) error {
	return w.runBD("reopen", id)
}

// UpdateTitle updates the title of a bead
func (w *Writer) UpdateTitle(id, title string) error {
	return w.runBD("update", id, "--title", title)
}

// UpdateDescription updates the description of a bead
func (w *Writer) UpdateDescription(id, description string) error {
	return w.runBD("update", id, "--description", description)
}

// UpdatePriority updates the priority of a bead
func (w *Writer) UpdatePriority(id string, priority int) error {
	return w.runBD("update", id, "--priority", fmt.Sprintf("%d", priority))
}

// AddDependency adds a dependency between beads
func (w *Writer) AddDependency(childID, parentID string) error {
	return w.runBD("dep", "add", childID, parentID)
}

// RemoveDependency removes a dependency between beads
func (w *Writer) RemoveDependency(childID, parentID string) error {
	return w.runBD("dep", "rm", childID, parentID)
}

// Init initializes beads in a project directory
func (w *Writer) Init() error {
	return w.runBD("init", "--non-interactive")
}

// IsInstalled checks if the bd command is available
func IsInstalled() bool {
	_, err := exec.LookPath("bd")
	return err == nil
}
