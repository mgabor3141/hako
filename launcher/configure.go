package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// `hako configure` -- a small TUI to enable/disable integrations, set their
// typed settings, and seal secrets, so the user never hand-edits hako.toml.
// On save it writes hako.toml (the enabled set + non-default settings).

func runConfigure(cfg *Config) {
	if len(cfg.Ints) == 0 {
		fatal("no integrations found in integrations/")
	}
	m := newCfgModel(cfg)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fatal("configure: " + err.Error())
	}
}

type rowKind int

const (
	rowInt rowKind = iota
	rowSetting
	rowSecret
)

type row struct {
	kind rowKind
	it   *Integration
	key  string // rowSetting: the setting name
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	dimStyle   = lipgloss.NewStyle().Faint(true)
	curStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	onStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

type cfgModel struct {
	cfg      *Config
	expanded map[string]bool
	rows     []row
	cursor   int
	editing  bool
	input    textinput.Model
	editIt   *Integration
	editKey  string
	dirty    bool
	status   string
}

type sealDoneMsg struct {
	name string
	err  error
}

type shellDoneMsg struct{ err error }

func newCfgModel(cfg *Config) *cfgModel {
	ti := textinput.New()
	ti.CharLimit = 512
	ti.Width = 48
	m := &cfgModel{cfg: cfg, expanded: map[string]bool{}, input: ti}
	m.rebuild()
	return m
}

// rebuild flattens integrations (+ expanded settings/secrets) into rows.
func (m *cfgModel) rebuild() {
	m.rows = m.rows[:0]
	for _, it := range m.cfg.Ints {
		m.rows = append(m.rows, row{kind: rowInt, it: it})
		if !m.expanded[it.Name] {
			continue
		}
		for _, k := range sortedKeys(it.Settings) {
			m.rows = append(m.rows, row{kind: rowSetting, it: it, key: k})
		}
		if len(it.Secrets) > 0 {
			m.rows = append(m.rows, row{kind: rowSecret, it: it})
		}
	}
	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}
}

func (m *cfgModel) Init() tea.Cmd { return nil }

func (m *cfgModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sealDoneMsg:
		if msg.err != nil {
			m.status = warnStyle.Render("seal " + msg.name + " failed or cancelled")
		} else {
			m.status = onStyle.Render("sealed a secret for " + msg.name)
		}
		return m, nil
	case shellDoneMsg:
		if msg.err != nil {
			m.status = warnStyle.Render("shell exited with an error -- is the stack up? (hako up)")
		} else {
			m.status = ""
		}
		return m, nil
	case tea.KeyMsg:
		if m.editing {
			return m.updateEditing(msg)
		}
		return m.updateNav(msg)
	}
	return m, nil
}

func (m *cfgModel) updateNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.rows)-1 {
			m.cursor++
		}
	case " ":
		m.toggle()
	case "enter":
		return m.activate()
	case "s":
		if it := m.rows[m.cursor].it; len(it.Secrets) > 0 {
			return m, m.seal(it.Name)
		}
	case "!":
		return m, m.shell()
	case "w":
		if err := writeHako(m.cfg); err != nil {
			m.status = warnStyle.Render("save failed: " + err.Error())
		} else {
			m.dirty = false
			m.status = onStyle.Render("saved hako.toml -- run `hako up` to apply")
		}
	}
	return m, nil
}

// toggle flips an integration on/off or a bool setting.
func (m *cfgModel) toggle() {
	r := m.rows[m.cursor]
	switch r.kind {
	case rowInt:
		r.it.Enabled = !r.it.Enabled
		m.dirty = true
	case rowSetting:
		if r.it.Settings[r.key].Type == "bool" {
			if r.it.Values[r.key] == "true" {
				r.it.Values[r.key] = "false"
			} else {
				r.it.Values[r.key] = "true"
			}
			m.dirty = true
		}
	}
}

// activate expands an integration, edits a setting, or seals a secret.
func (m *cfgModel) activate() (tea.Model, tea.Cmd) {
	r := m.rows[m.cursor]
	switch r.kind {
	case rowInt:
		m.expanded[r.it.Name] = !m.expanded[r.it.Name]
		m.rebuild()
	case rowSetting:
		if r.it.Settings[r.key].Type == "bool" {
			m.toggle()
			return m, nil
		}
		m.editing = true
		m.editIt, m.editKey = r.it, r.key
		m.input.SetValue(r.it.Values[r.key])
		m.input.CursorEnd()
		m.input.Focus()
		return m, textinput.Blink
	case rowSecret:
		return m, m.seal(r.it.Name)
	}
	return m, nil
}

func (m *cfgModel) updateEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		v := strings.TrimSpace(m.input.Value())
		if m.editIt.Settings[m.editKey].Type == "number" && v != "" {
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				m.status = warnStyle.Render("not a number: " + v)
				return m, nil
			}
		}
		m.editIt.Values[m.editKey] = v
		m.dirty = true
		m.editing = false
		m.input.Blur()
		return m, nil
	case "esc":
		m.editing = false
		m.input.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// shell drops the user into the agent container's zsh, then returns to the TUI
// (tea.ExecProcess releases the terminal for the duration). Same invocation as
// the `hako shell` verb (shellArgs).
func (m *cfgModel) shell() tea.Cmd {
	c := exec.Command("docker", shellArgs()...)
	return tea.ExecProcess(c, func(err error) tea.Msg { return shellDoneMsg{err} })
}

// seal hands the terminal to `hako seal <name>` (the masked-prompt flow).
func (m *cfgModel) seal(name string) tea.Cmd {
	self, err := os.Executable()
	if err != nil {
		return func() tea.Msg { return sealDoneMsg{name, err} }
	}
	c := exec.Command(self, "seal", name)
	return tea.ExecProcess(c, func(err error) tea.Msg { return sealDoneMsg{name, err} })
}

func (m *cfgModel) View() string {
	var b strings.Builder
	fmt.Fprintln(&b, titleStyle.Render("hako configure"))
	fmt.Fprintln(&b, dimStyle.Render("enable integrations, set their options, seal secrets"))
	b.WriteByte('\n')

	for i, r := range m.rows {
		cur := "  "
		if i == m.cursor {
			cur = curStyle.Render("> ")
		}
		switch r.kind {
		case rowInt:
			box := "[ ]"
			if r.it.Enabled {
				box = onStyle.Render("[x]")
			}
			arrow := "▸"
			if m.expanded[r.it.Name] {
				arrow = "▾"
			}
			fmt.Fprintf(&b, "%s%s %s %-12s %s\n", cur, box, arrow, r.it.Name, dimStyle.Render(r.it.Summary))
		case rowSetting:
			s := r.it.Settings[r.key]
			val := r.it.Values[r.key]
			if m.editing && m.editIt == r.it && m.editKey == r.key {
				fmt.Fprintf(&b, "%s      %s = %s\n", cur, r.key, m.input.View())
			} else {
				shown := val
				if s.Type == "bool" {
					shown = "[x]"
					if val != "true" {
						shown = "[ ]"
					}
				}
				fmt.Fprintf(&b, "%s      %s = %s  %s\n", cur, r.key, shown, dimStyle.Render(s.Description))
			}
		case rowSecret:
			envs := make([]string, len(r.it.Secrets))
			for j, s := range r.it.Secrets {
				envs[j] = s.Env
			}
			fmt.Fprintf(&b, "%s      %s %s\n", cur, warnStyle.Render("secret:"),
				dimStyle.Render(strings.Join(envs, ", ")+"  (enter/s to seal)"))
		}
	}

	b.WriteByte('\n')
	help := "↑/↓ move · space toggle · enter expand/edit · s seal · ! shell · w save · q quit"
	if m.dirty {
		help = warnStyle.Render("● unsaved") + " · " + help
	}
	fmt.Fprintln(&b, dimStyle.Render(help))
	if m.status != "" {
		fmt.Fprintln(&b, m.status)
	}
	return b.String()
}

// writeHako serializes the enabled set + non-default settings to hako.toml.
func writeHako(cfg *Config) error {
	var b strings.Builder
	b.WriteString("# hako.toml -- your enabled integrations and their settings.\n")
	b.WriteString("# Managed by `hako configure`; safe to hand-edit. Gitignored.\n")
	for _, it := range cfg.Ints {
		fmt.Fprintf(&b, "\n[integrations.%s]\n", it.Name)
		fmt.Fprintf(&b, "enabled = %t\n", it.Enabled)
		for _, k := range sortedKeys(it.Settings) {
			s := it.Settings[k]
			v := it.Values[k]
			if v == toStr(s.Default) {
				continue // leave defaults to the manifest
			}
			fmt.Fprintf(&b, "%s = %s\n", k, tomlValue(s.Type, v))
		}
	}
	return os.WriteFile(filepath.Join(cfg.Root, "hako.toml"), []byte(b.String()), 0o644)
}

func tomlValue(typ, v string) string {
	switch typ {
	case "bool", "number":
		if v == "" {
			return `""`
		}
		return v
	default:
		return strconv.Quote(v)
	}
}

func sortedKeys(m map[string]Setting) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
