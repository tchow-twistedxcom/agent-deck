package ui

import (
	"fmt"
	"strings"

	"github.com/asheshgoplani/agent-deck/internal/beads"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BeadsPanel displays beads (tasks) from the current project
type BeadsPanel struct {
	visible       bool
	width, height int
	cursor        int
	beadsList     []*beads.Bead
	projectPath   string
	reader        *beads.Reader
	errorMsg      string
	loading       bool
}

// NewBeadsPanel creates a new beads panel
func NewBeadsPanel() *BeadsPanel {
	return &BeadsPanel{
		beadsList: make([]*beads.Bead, 0),
	}
}

// Show displays the beads panel for a project
func (b *BeadsPanel) Show(projectPath string) tea.Cmd {
	b.visible = true
	b.projectPath = projectPath
	b.cursor = 0
	b.errorMsg = ""
	b.loading = true

	return b.loadBeads
}

// Hide hides the beads panel
func (b *BeadsPanel) Hide() {
	b.visible = false
}

// IsVisible returns true if the panel is visible
func (b *BeadsPanel) IsVisible() bool {
	return b.visible
}

// SetSize sets the panel dimensions
func (b *BeadsPanel) SetSize(width, height int) {
	b.width = width
	b.height = height
}

// GetSelectedBead returns the currently selected bead
func (b *BeadsPanel) GetSelectedBead() *beads.Bead {
	if b.cursor >= 0 && b.cursor < len(b.beadsList) {
		return b.beadsList[b.cursor]
	}
	return nil
}

// GetProjectPath returns the project path being viewed
func (b *BeadsPanel) GetProjectPath() string {
	return b.projectPath
}

// beadsLoadedMsg is sent when beads are loaded
type beadsLoadedMsg struct {
	beads []*beads.Bead
	err   error
}

// loadBeads loads beads from the project
func (b *BeadsPanel) loadBeads() tea.Msg {
	if b.projectPath == "" {
		return beadsLoadedMsg{err: fmt.Errorf("no project path")}
	}

	reader := beads.NewReader(b.projectPath)
	if !reader.HasBeads() {
		return beadsLoadedMsg{err: fmt.Errorf("no beads initialized in this project")}
	}

	beadsList, err := reader.ReadOpen()
	if err != nil {
		return beadsLoadedMsg{err: err}
	}

	return beadsLoadedMsg{beads: beadsList}
}

// Update handles keyboard input
func (b *BeadsPanel) Update(msg tea.Msg) (*BeadsPanel, tea.Cmd) {
	switch msg := msg.(type) {
	case beadsLoadedMsg:
		b.loading = false
		if msg.err != nil {
			b.errorMsg = msg.err.Error()
		} else {
			b.beadsList = msg.beads
		}
		return b, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "b", "q"))):
			b.Hide()
			return b, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if b.cursor > 0 {
				b.cursor--
			}
			return b, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if b.cursor < len(b.beadsList)-1 {
				b.cursor++
			}
			return b, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Return the selected bead (handled by parent)
			return b, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			// Claim the selected bead
			if bead := b.GetSelectedBead(); bead != nil {
				return b, b.claimBead(bead.ID)
			}
			return b, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
			// Close the selected bead
			if bead := b.GetSelectedBead(); bead != nil {
				return b, b.closeBead(bead.ID)
			}
			return b, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			// Refresh beads
			b.loading = true
			return b, b.loadBeads
		}
	}

	return b, nil
}

// beadClaimedMsg is sent when a bead is claimed
type beadClaimedMsg struct {
	id  string
	err error
}

// beadClosedMsg is sent when a bead is closed
type beadClosedMsg struct {
	id  string
	err error
}

func (b *BeadsPanel) claimBead(id string) tea.Cmd {
	return func() tea.Msg {
		writer := beads.NewWriter(b.projectPath)
		err := writer.Claim(id)
		return beadClaimedMsg{id: id, err: err}
	}
}

func (b *BeadsPanel) closeBead(id string) tea.Cmd {
	return func() tea.Msg {
		writer := beads.NewWriter(b.projectPath)
		err := writer.Close(id)
		return beadClosedMsg{id: id, err: err}
	}
}

// View renders the beads panel
func (b *BeadsPanel) View() string {
	if !b.visible {
		return ""
	}

	// Panel styling
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2).
		Width(b.width - 4).
		Height(b.height - 4)

	titleStyle := lipgloss.NewStyle().
		Foreground(ColorCyan).
		Bold(true)

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("BEADS - Task Tracker"))
	content.WriteString("\n\n")

	// Error or loading state
	if b.loading {
		content.WriteString(lipgloss.NewStyle().Foreground(ColorComment).Render("Loading beads..."))
	} else if b.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(ColorRed).Render(b.errorMsg))
		content.WriteString("\n\nPress 'q' or 'esc' to close")
	} else if len(b.beadsList) == 0 {
		content.WriteString(lipgloss.NewStyle().Foreground(ColorComment).Render("No open beads"))
		content.WriteString("\n\nCreate beads with: bd create \"Task title\" -p 0")
	} else {
		// Render beads list
		for i, bead := range b.beadsList {
			line := b.renderBeadLine(bead, i == b.cursor)
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	// Help footer
	content.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(ColorComment)
	content.WriteString(helpStyle.Render("j/k:navigate  c:claim  x:close  r:refresh  esc:back"))

	return lipgloss.Place(
		b.width, b.height,
		lipgloss.Center, lipgloss.Center,
		panelStyle.Render(content.String()),
	)
}

// renderBeadLine renders a single bead line
func (b *BeadsPanel) renderBeadLine(bead *beads.Bead, selected bool) string {
	// Status icon
	statusIcon := bead.StatusIcon()
	statusStyle := lipgloss.NewStyle()
	switch bead.Status {
	case beads.StatusOpen:
		statusStyle = statusStyle.Foreground(ColorText)
	case beads.StatusInProgress:
		statusStyle = statusStyle.Foreground(ColorYellow)
	case beads.StatusClosed:
		statusStyle = statusStyle.Foreground(ColorGreen)
	}

	// Priority badge
	priorityStyle := lipgloss.NewStyle().Bold(true)
	switch bead.Priority {
	case 0:
		priorityStyle = priorityStyle.Foreground(ColorRed)
	case 1:
		priorityStyle = priorityStyle.Foreground(ColorOrange)
	case 2:
		priorityStyle = priorityStyle.Foreground(ColorYellow)
	default:
		priorityStyle = priorityStyle.Foreground(ColorComment)
	}

	// Type badge
	typeStyle := lipgloss.NewStyle().Foreground(ColorPurple)
	typeStr := ""
	switch bead.IssueType {
	case beads.TypeEpic:
		typeStr = "[epic]"
	case beads.TypeTask:
		typeStr = "[task]"
	case beads.TypeSubtask:
		typeStr = "[sub]"
	}

	// Build line
	var line strings.Builder

	// Selection indicator
	if selected {
		line.WriteString(lipgloss.NewStyle().Foreground(ColorAccent).Render("▶ "))
	} else {
		line.WriteString("  ")
	}

	// Status icon
	line.WriteString(statusStyle.Render(statusIcon))
	line.WriteString(" ")

	// Priority
	line.WriteString(priorityStyle.Render(bead.PriorityLabel()))
	line.WriteString(" ")

	// Type
	line.WriteString(typeStyle.Render(typeStr))
	line.WriteString(" ")

	// ID (dimmed)
	line.WriteString(lipgloss.NewStyle().Foreground(ColorComment).Render(bead.ID))
	line.WriteString(": ")

	// Title
	titleStyle := lipgloss.NewStyle().Foreground(ColorText)
	if selected {
		titleStyle = titleStyle.Foreground(ColorAccent).Bold(true)
	}
	line.WriteString(titleStyle.Render(bead.Title))

	return line.String()
}

// HasBeads returns true if beads are available
func (b *BeadsPanel) HasBeads() bool {
	return len(b.beadsList) > 0
}

// Refresh reloads the beads
func (b *BeadsPanel) Refresh() tea.Cmd {
	b.loading = true
	return b.loadBeads
}
