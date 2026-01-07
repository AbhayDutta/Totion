package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* ---------------- STYLES ---------------- */

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(1, 2)

	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

/* ---------------- GLOBAL ---------------- */

var vaultDir string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	vaultDir = fmt.Sprintf("%s/.totion", home)
	os.MkdirAll(vaultDir, 0750)
}

/* ---------------- LIST ITEM ---------------- */

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

/* ---------------- MODEL ---------------- */

type model struct {
	input       textinput.Model
	editor      textarea.Model
	list        list.Model
	currentFile *os.File

	showInput bool
	showList  bool
}

/* ---------------- INIT ---------------- */

func (m model) Init() tea.Cmd {
	return nil
}

/* ---------------- UPDATE ---------------- */

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "ctrl+n":
			m.showInput = true
			m.showList = false
			return m, nil

		case "ctrl+l":
			m.list.SetItems(listFiles())
			m.showList = true
			m.showInput = false
			return m, nil

		case "ctrl+s":
			if m.currentFile != nil {
				m.currentFile.Truncate(0)
				m.currentFile.Seek(0, 0)
				m.currentFile.WriteString(m.editor.Value())
				m.currentFile.Close()
				m.currentFile = nil
				m.editor.SetValue("")
			}
			return m, nil

		case "ctrl+d": // 🔥 DELETE NOTE
			if m.currentFile != nil {
				name := m.currentFile.Name()
				m.currentFile.Close()
				os.Remove(name)
				m.currentFile = nil
				m.editor.SetValue("")
				m.list.SetItems(listFiles())
			}
			return m, nil

		case "esc":
			m.showInput = false
			m.showList = false
			return m, nil

		case "enter":
			if m.showInput {
				filename := m.input.Value()
				if filename != "" {
					path := fmt.Sprintf("%s/%s.md", vaultDir, filename)
					f, _ := os.Create(path)
					m.currentFile = f
					m.input.SetValue("")
					m.showInput = false
				}
				return m, nil
			}

			if m.showList {
				it, ok := m.list.SelectedItem().(item)
				if ok {
					path := fmt.Sprintf("%s/%s", vaultDir, it.title)
					content, _ := os.ReadFile(path)
					m.editor.SetValue(string(content))
					f, _ := os.OpenFile(path, os.O_RDWR, 0644)
					m.currentFile = f
					m.showList = false
				}
				return m, nil
			}
		}
	}

	if m.showInput {
		m.input, cmd = m.input.Update(msg)
	}
	if m.currentFile != nil {
		m.editor, cmd = m.editor.Update(msg)
	}
	if m.showList {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

/* ---------------- VIEW ---------------- */

func (m model) View() string {
	header := headerStyle.Render("📘 Totion — Terminal Notes")

	help := helpStyle.Render(
		"Ctrl+N New  •  Ctrl+L List  •  Ctrl+S Save  •  Ctrl+D Delete  •  Esc Back  •  Q Quit",
	)

	content := ""

	switch {
	case m.showInput:
		content = boxStyle.Render(m.input.View())
	case m.showList:
		content = boxStyle.Render(m.list.View())
	case m.currentFile != nil:
		content = boxStyle.Render(m.editor.View())
	default:
		content = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Render("Press Ctrl+N to create a note or Ctrl+L to open one")
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, content, help)
}

/* ---------------- INIT MODEL ---------------- */

func initializeModel() model {
	ti := textinput.New()
	ti.Placeholder = "New note name..."
	ti.Cursor.Style = cursorStyle
	ti.Focus()

	ta := textarea.New()
	ta.Placeholder = "Write your note..."
	ta.Focus()

	l := list.New(listFiles(), list.NewDefaultDelegate(), 40, 10)
	l.Title = "Your Notes"

	return model{
		input:  ti,
		editor: ta,
		list:   l,
	}
}

/* ---------------- FILE LIST ---------------- */

func listFiles() []list.Item {
	items := []list.Item{}
	entries, _ := os.ReadDir(vaultDir)

	for _, e := range entries {
		if !e.IsDir() {
			items = append(items, item{
				title: e.Name(),
				desc:  "Markdown note",
			})
		}
	}
	return items
}

/* ---------------- MAIN ---------------- */

func main() {
	p := tea.NewProgram(initializeModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
