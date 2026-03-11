package main

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type processedMsg struct{ name string }
type errorMsg struct{ name, err string }
type doneMsg struct{}

var (
	subtle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	bold   = lipgloss.NewStyle().Bold(true)
)

type model struct {
	jobs      []ImageJob
	progress  progress.Model
	current   int
	succeeded int
	failed    int
	width     int
	curName   string
	errors    []string
	quitting  bool
}

func newModel(jobs []ImageJob) model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)
	return model{
		jobs:     jobs,
		progress: p,
		curName:  filepath.Base(jobs[0].Input),
	}
}

func (m model) processNext() tea.Cmd {
	if m.current >= len(m.jobs) {
		return func() tea.Msg { return doneMsg{} }
	}
	job := m.jobs[m.current]
	return func() tea.Msg {
		name := filepath.Base(job.Input)
		if err := processFile(job.Input, job.Output); err != nil {
			return errorMsg{name: name, err: err.Error()}
		}
		return processedMsg{name: name}
	}
}

func (m model) Init() tea.Cmd {
	return m.processNext()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.progress.Width = msg.Width - 10
		if m.progress.Width < 20 {
			m.progress.Width = 20
		}
		return m, nil

	case processedMsg:
		m.succeeded++
		m.current++
		if m.current < len(m.jobs) {
			m.curName = filepath.Base(m.jobs[m.current].Input)
		}
		pct := float64(m.current) / float64(len(m.jobs))
		cmd := m.progress.SetPercent(pct)
		return m, tea.Batch(cmd, m.processNext())

	case errorMsg:
		m.failed++
		m.current++
		m.errors = append(m.errors, fmt.Sprintf("%s: %s", msg.name, msg.err))
		if m.current < len(m.jobs) {
			m.curName = filepath.Base(m.jobs[m.current].Input)
		}
		pct := float64(m.current) / float64(len(m.jobs))
		cmd := m.progress.SetPercent(pct)
		return m, tea.Batch(cmd, m.processNext())

	case doneMsg:
		m.quitting = true
		return m, tea.Quit

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		s := bold.Render(fmt.Sprintf("Done! %d converted", m.succeeded))
		if m.failed > 0 {
			s += subtle.Render(fmt.Sprintf(", %d failed", m.failed))
		}
		s += "\n"
		for _, e := range m.errors {
			s += subtle.Render("  ✗ "+e) + "\n"
		}
		return s
	}

	total := len(m.jobs)
	counter := subtle.Render(fmt.Sprintf(" %d/%d", m.current, total))
	name := subtle.Render("  " + m.curName)
	return "\n" + m.progress.View() + counter + name + "\n"
}

func RunProgressUI(jobs []ImageJob) error {
	p := tea.NewProgram(newModel(jobs))
	_, err := p.Run()
	return err
}
