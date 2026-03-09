package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Application directories to scan for .desktop files.
var appDirs = []string{
	"/usr/share/applications",
	"/var/lib/flatpak/exports/share/applications",
}

// Colors.
var (
	white = lipgloss.Color("#EBE8E2")
	gray  = lipgloss.Color("#444444")
	black = lipgloss.Color("#000000")
)

// Styles.
var (
	windowStyle   = lipgloss.NewStyle().Padding(1, 2).Background(black)
	selectedStyle = lipgloss.NewStyle().Foreground(white).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(gray)
	hatchedStyle  = lipgloss.NewStyle().Foreground(gray)
)

// -------------------------------------------------------------------
// Item
// -------------------------------------------------------------------

type item struct {
	name    string
	command string
}

func (i item) FilterValue() string { return i.name }

// -------------------------------------------------------------------
// Item delegate
// -------------------------------------------------------------------

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render("> "+i.name))
	} else {
		fmt.Fprint(w, normalStyle.Render("  "+i.name))
	}
}

// -------------------------------------------------------------------
// Model
// -------------------------------------------------------------------

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				selected, ok := m.list.SelectedItem().(item)
				if !ok {
					return m, tea.Quit
				}
				launchDetached(selected.command)
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width-4, msg.Height-4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return windowStyle.Render(
		m.list.View() + "\n\n" + hatchedStyle.Render(" /////////"),
	)
}

// -------------------------------------------------------------------
// Desktop entry parsing
// -------------------------------------------------------------------

// parseDesktopEntry extracts the application name and exec command from
// a .desktop file. It respects NoDisplay and hidden entries, returning
// empty strings when the entry should be skipped.
func parseDesktopEntry(path string) (name, cmd string) {
	file, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	inDesktopEntry := false

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		// Track section headers — only read [Desktop Entry].
		if strings.HasPrefix(line, "[") {
			inDesktopEntry = line == "[Desktop Entry]"
			continue
		}
		if !inDesktopEntry {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "Name":
			if name == "" {
				name = value
			}
		case "Exec":
			if cmd == "" {
				cmd = value
				// Strip field codes (%f, %F, %u, %U, etc.) and stray quotes.
				if i := strings.Index(cmd, " %"); i != -1 {
					cmd = cmd[:i]
				}
				cmd = strings.ReplaceAll(cmd, "\"", "")
			}
		case "NoDisplay", "Hidden":
			if strings.EqualFold(value, "true") {
				return "", ""
			}
		}
	}
	return name, cmd
}

// scanApplications walks application directories and returns a sorted
// slice of discovered items, deduplicating by name.
func scanApplications() []list.Item {
	seen := make(map[string]struct{})
	var items []list.Item

	for _, dir := range appDirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.desktop"))
		if err != nil {
			continue
		}

		for _, path := range files {
			name, cmd := parseDesktopEntry(path)
			if name == "" || cmd == "" {
				continue
			}
			if _, dup := seen[name]; dup {
				continue
			}
			seen[name] = struct{}{}
			items = append(items, item{name: name, command: cmd})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].(item).name) < strings.ToLower(items[j].(item).name)
	})

	return items
}

// -------------------------------------------------------------------
// Process launching
// -------------------------------------------------------------------

// launchDetached starts a command in a new session, fully detached
// from the current terminal so it survives after the launcher exits.
func launchDetached(command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	_ = cmd.Start()
}

// -------------------------------------------------------------------
// Single-instance lock
// -------------------------------------------------------------------

// acquireLock ensures only one instance of the launcher is running.
// Returns a cleanup function that must be deferred by the caller.
func acquireLock() (cleanup func(), err error) {
	const lockPath = "/tmp/golauncher.lock"

	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		return nil, fmt.Errorf("already running")
	}

	return func() {
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
		os.Remove(lockPath)
	}, nil
}

// -------------------------------------------------------------------
// Entrypoint
// -------------------------------------------------------------------

func main() {
	unlock, err := acquireLock()
	if err != nil {
		os.Exit(0)
	}
	defer unlock()

	items := scanApplications()
	if len(items) == 0 {
		fmt.Fprintln(os.Stderr, "golauncher: no applications found")
		os.Exit(1)
	}

	l := list.New(items, itemDelegate{}, 30, 10)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(white)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(white)
	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(white)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(white)

	p := tea.NewProgram(model{list: l}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "golauncher: failed to start TUI:", err)
		os.Exit(1)
	}
}
