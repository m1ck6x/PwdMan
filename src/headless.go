package main

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
)

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("128"))
	noStyle          = lipgloss.NewStyle()
	blurred          = false
	pwCopied         = false
	pwAdditive       = actualPwAdditive
	actualPwAdditive = baseStyle.Foreground(lipgloss.Color("127")).Render("Password copied!") + "\n"
	savedAdditive    = baseStyle.Foreground(lipgloss.Color("127")).Render("Saved to disk!") + "\n"
)

type model struct {
	table          table.Model
	inpService     textinput.Model
	inpDescription textinput.Model
	inpNotes       textarea.Model
	inpUser        textinput.Model
	inpPw          textinput.Model

	KeyMap    customKeyMap
	SelKeyMap customSelKeyMap
	Help      help.Model
	SelHelp   help.Model
	accounts  *[]account
	selected  *account
	focus     int
}

type customKeyMap struct {
	Blur       key.Binding
	LineUp     key.Binding
	LineDown   key.Binding
	Select     key.Binding
	Quit       key.Binding
	CopyPw     key.Binding
	GotoTop    key.Binding
	GotoBottom key.Binding
}

type customSelKeyMap struct {
	Blur   key.Binding
	Next   key.Binding
	Quit   key.Binding
	Back   key.Binding
	Save   key.Binding
	Delete key.Binding
	NewPw  key.Binding
	ShowPw key.Binding
	CopyPw key.Binding
}

func (m model) Init() tea.Cmd { return nil }

// 𝕸𝖆𝖞 𝖙𝖍𝖊 𝖑𝖔𝖗𝖉'𝖘 𝖒𝖊𝖗𝖈𝖞 𝖇𝖊 𝖚𝖕𝖔𝖓 𝖙𝖍𝖊𝖊, 𝖋𝖔𝖗 𝖙𝖍𝖔𝖚 𝖑𝖆𝖞𝖊𝖘𝖙 𝖍𝖆𝖓𝖉 𝖚𝖕𝖔𝖓 𝖙𝖍𝖎𝖘 𝖜𝖔𝖊𝖋𝖚𝖑 𝖈𝖗𝖆𝖋𝖙, 𝖋𝖎𝖙 𝖋𝖔𝖗 𝖓𝖊𝖎𝖙𝖍𝖊𝖗 𝖒𝖆𝖓 𝖓𝖔𝖗 𝖇𝖊𝖆𝖘𝖙.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {

		case key.Matches(msg, m.KeyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.SelKeyMap.NewPw):
			pwBuf := generatePw()
			m.inpPw.SetValue(string(*pwBuf))
			zero(pwBuf)

		case key.Matches(msg, m.KeyMap.Blur), key.Matches(msg, m.SelKeyMap.Blur):
			if blurred {
				blurred = false

				m.KeyMap.CopyPw.SetEnabled(true)
				m.KeyMap.GotoBottom.SetEnabled(true)
				m.KeyMap.GotoTop.SetEnabled(true)
				m.KeyMap.LineDown.SetEnabled(true)
				m.KeyMap.LineUp.SetEnabled(true)
				m.KeyMap.Quit.SetEnabled(true)
				m.KeyMap.Select.SetEnabled(true)

				m.SelKeyMap.Back.SetEnabled(true)
				m.SelKeyMap.CopyPw.SetEnabled(true)
				m.SelKeyMap.NewPw.SetEnabled(true)
				m.SelKeyMap.Next.SetEnabled(true)
				m.SelKeyMap.Quit.SetEnabled(true)
				m.SelKeyMap.Save.SetEnabled(true)
				m.SelKeyMap.ShowPw.SetEnabled(true)

				m.table.Blur()
				m.inpService.Blur()
				m.inpDescription.Blur()
				m.inpNotes.Blur()
				m.inpUser.Blur()
				m.inpPw.Blur()

				if m.selected == nil {
					m.table.Focus()
				}
			} else {
				blurred = true

				m.KeyMap.CopyPw.SetEnabled(false)
				m.KeyMap.GotoBottom.SetEnabled(false)
				m.KeyMap.GotoTop.SetEnabled(false)
				m.KeyMap.LineDown.SetEnabled(false)
				m.KeyMap.LineUp.SetEnabled(false)
				m.KeyMap.Quit.SetEnabled(false)
				m.KeyMap.Select.SetEnabled(false)

				m.SelKeyMap.Back.SetEnabled(false)
				m.SelKeyMap.CopyPw.SetEnabled(false)
				m.SelKeyMap.NewPw.SetEnabled(false)
				m.SelKeyMap.Next.SetEnabled(false)
				m.SelKeyMap.Quit.SetEnabled(false)
				m.SelKeyMap.Save.SetEnabled(false)
				m.SelKeyMap.ShowPw.SetEnabled(false)

				m.table.Blur()
				m.inpService.Blur()
				m.inpDescription.Blur()
				m.inpNotes.Blur()
				m.inpUser.Blur()
				m.inpPw.Blur()

				if m.selected != nil {
					switch m.focus {

					// case 0
					default:
						m.inpService.PromptStyle = focusedStyle
						return m, m.inpService.Focus()

					case 1:
						m.inpDescription.PromptStyle = focusedStyle
						return m, m.inpDescription.Focus()

					case 2:
						return m, m.inpNotes.Focus()

					case 3:
						m.inpUser.PromptStyle = focusedStyle
						return m, m.inpUser.Focus()

					case 4:
						m.inpPw.PromptStyle = focusedStyle
						return m, m.inpPw.Focus()
					}
				}
			}

		case key.Matches(msg, m.SelKeyMap.ShowPw):
			if m.inpPw.EchoMode == textinput.EchoNormal {
				m.inpPw.EchoMode = textinput.EchoPassword
			} else {
				m.inpPw.EchoMode = textinput.EchoNormal
			}

		case key.Matches(msg, m.SelKeyMap.Back):
			m.inpService.Blur()
			m.inpDescription.Blur()
			m.inpNotes.Blur()
			m.inpUser.Blur()
			m.inpPw.Blur()

			m.focus = 0
			m.selected = nil

			m.table.Focus()

		case key.Matches(msg, m.SelKeyMap.Delete):
			// TODO: Ask for confirmation from the user like pwAdditive

			index, _ := strconv.ParseInt(m.table.SelectedRow()[0], 10, 32)

			if index == 0 {
				m.inpService.Blur()
				m.inpDescription.Blur()
				m.inpNotes.Blur()
				m.inpUser.Blur()
				m.inpPw.Blur()

				m.focus = 0
				m.selected = nil

				m.table.Focus()

				break
			}

			*m.accounts = slices.Delete(*m.accounts, int(index), int(index)+1)

			temp := (*m.accounts)[1:]
			saveAccountsToDisk(&temp)

			m.inpService.Blur()
			m.inpDescription.Blur()
			m.inpNotes.Blur()
			m.inpUser.Blur()
			m.inpPw.Blur()

			m.focus = 0
			m.selected = nil

			rows := []table.Row{}

			for index, account := range *m.accounts {
				rows = append(rows, table.Row{
					fmt.Sprint(index),
					account.Service,
					account.Description,
					account.Notes,
				})
			}

			m.table.SetRows(rows)
			m.table.Focus()

		case key.Matches(msg, m.SelKeyMap.Save):
			if m.selected == nil {
				break
			}

			index, _ := strconv.ParseInt(m.table.SelectedRow()[0], 10, 32)

			if index != 0 {
				m.selected.Service = m.inpService.Value()
				m.selected.Description = m.inpDescription.Value()
				m.selected.Notes = m.inpNotes.Value()
				m.selected.User = m.inpUser.Value()
				m.selected.Pw = m.inpPw.Value()
			} else if len(m.inpService.Value()) > 0 && len(m.inpPw.Value()) > 0 {
				*m.accounts = append(*m.accounts, account{
					Service:     m.inpService.Value(),
					Description: m.inpDescription.Value(),
					Notes:       m.inpNotes.Value(),
					User:        m.inpUser.Value(),
					Pw:          m.inpPw.Value(),
				})
			}

			// TODO: Add confirmation with help display like pwAdditive
			temp := (*m.accounts)[1:]
			saveAccountsToDisk(&temp)

			pwAdditive = savedAdditive
			pwCopied = true

			time.AfterFunc(3*time.Second, func() {
				pwCopied = false
			})

			m.inpService.Blur()
			m.inpDescription.Blur()
			m.inpNotes.Blur()
			m.inpUser.Blur()
			m.inpPw.Blur()

			m.focus = 0
			m.selected = nil

			rows := []table.Row{}

			for index, account := range *m.accounts {
				rows = append(rows, table.Row{
					fmt.Sprint(index),
					account.Service,
					account.Description,
					account.Notes,
				})
			}

			m.table.SetRows(rows)
			m.table.Focus()

		case key.Matches(msg, m.KeyMap.CopyPw), key.Matches(msg, m.SelKeyMap.CopyPw):
			if m.selected == nil {
				index, _ := strconv.ParseInt(m.table.SelectedRow()[0], 10, 32)
				account := (*m.accounts)[index]

				clipboard.Write(clipboard.FmtText, []byte(account.Pw))
			} else {
				clipboard.Write(clipboard.FmtText, []byte(m.selected.Pw))
			}

			pwAdditive = actualPwAdditive
			pwCopied = true

			time.AfterFunc(3*time.Second, func() {
				pwCopied = false
			})

		case key.Matches(msg, m.KeyMap.Select), key.Matches(msg, m.SelKeyMap.Next), key.Matches(msg, m.SelKeyMap.Back):
			if m.selected == nil {
				index, _ := strconv.ParseInt(m.table.SelectedRow()[0], 10, 32)
				m.selected = &(*m.accounts)[index]

				if index != 0 {
					account := *m.selected

					m.inpService.SetValue(account.Service)
					m.inpDescription.SetValue(account.Description)
					m.inpNotes.SetValue(account.Notes)
					m.inpUser.SetValue(string(account.User))
					m.inpPw.SetValue(string(account.Pw))
				} else {
					m.inpService.SetValue("")
					m.inpDescription.SetValue("")
					m.inpNotes.SetValue("")
					m.inpUser.SetValue("")
					m.inpPw.SetValue("")
				}

				m.focus = 0
				m.table.Blur()
				m.inpService.PromptStyle = focusedStyle

				return m, tea.ClearScreen
			} else {
				m.focus = (m.focus + 1) % 5

				m.inpService.Blur()
				m.inpDescription.Blur()
				m.inpNotes.Blur()
				m.inpUser.Blur()
				m.inpPw.Blur()

				m.inpService.PromptStyle = noStyle
				m.inpDescription.PromptStyle = noStyle
				m.inpNotes.BlurredStyle.Prompt = noStyle
				m.inpUser.PromptStyle = noStyle
				m.inpPw.PromptStyle = noStyle

				switch m.focus {

				// case 0
				default:
					m.inpService.PromptStyle = focusedStyle

				case 1:
					m.inpDescription.PromptStyle = focusedStyle

				case 2:
					m.inpNotes.BlurredStyle.Prompt = focusedStyle

				case 3:
					m.inpUser.PromptStyle = focusedStyle

				case 4:
					m.inpPw.PromptStyle = focusedStyle
				}
			}
		}

	case tea.WindowSizeMsg:
		workableWidth := msg.Width
		workableHeight := msg.Height

		extraSpace := 13

		m.table.SetColumns([]table.Column{
			{Title: "ID", Width: 3},
			{Title: "Service", Width: workableWidth / 5},
			{Title: "Description", Width: workableWidth / 5 * 2},
			{Title: "Notes", Width: workableWidth/5*2 - extraSpace},
		})

		m.table.SetHeight(workableHeight / 3 * 2)

		m.inpNotes.SetWidth(workableWidth - extraSpace)
		m.inpPw.Width = workableWidth - extraSpace
		m.inpUser.Width = workableWidth - extraSpace
		m.inpService.Width = workableWidth - extraSpace
		m.inpDescription.Width = workableWidth - extraSpace
		m.inpUser.Width = workableWidth - extraSpace
		m.inpPw.Width = workableWidth - extraSpace
	}

	if m.table.Focused() {
		m.table, cmd = m.table.Update(msg)
	} else {
		switch m.focus {

		// case 0
		default:
			m.inpService, cmd = m.inpService.Update(msg)

		case 1:
			m.inpDescription, cmd = m.inpDescription.Update(msg)

		case 2:
			m.inpNotes, cmd = m.inpNotes.Update(msg)

		case 3:
			m.inpUser, cmd = m.inpUser.Update(msg)

		case 4:
			m.inpPw, cmd = m.inpPw.Update(msg)
		}
	}

	return m, cmd
}

func (m model) View() string {
	if m.selected == nil {
		finalRender := baseStyle.Render(m.table.View()) + "\n"

		if blurred {
			finalRender += baseStyle.Render(m.Help.ShortHelpView(m.KeyMap.ShortHelp())) + "\n"
		} else {
			finalRender += baseStyle.Render(m.Help.FullHelpView(m.KeyMap.FullHelp())) + "\n"
		}

		if pwCopied {
			finalRender += pwAdditive
		}

		return finalRender
	}

	render := fmt.Sprintf(
		"Service\n%s\n\nDescription\n%s\n\nNotes\n%s\n\nUser\n%s\n\nPassword\n%s\n",
		m.inpService.View(),
		m.inpDescription.View(),
		m.inpNotes.View(),
		m.inpUser.View(),
		m.inpPw.View(),
	)

	var help string

	if blurred {
		help = m.SelHelp.ShortHelpView(m.SelKeyMap.ShortHelp())
	} else {
		help += m.SelHelp.FullHelpView(m.SelKeyMap.FullHelp())
	}

	finalRender := fmt.Sprintf(
		"%s\n%s\n",
		baseStyle.Render(render),
		baseStyle.Render(help),
	)

	if pwCopied {
		finalRender += pwAdditive
	}

	return finalRender
}

func (km customKeyMap) ShortHelp() []key.Binding {
	km.Blur.SetHelp("esc", "Unlock focus")
	return []key.Binding{km.Blur}
}

func (km customKeyMap) FullHelp() [][]key.Binding {
	km.Blur.SetHelp("esc", "Lock focus")
	return [][]key.Binding{
		{km.Blur, km.Select, km.CopyPw, km.LineUp, km.LineDown},
		{km.GotoTop, km.GotoBottom, km.Quit},
	}
}

func (km customSelKeyMap) ShortHelp() []key.Binding {
	km.Blur.SetHelp("esc", "Unlock focus")
	return []key.Binding{km.Blur}
}

func (km customSelKeyMap) FullHelp() [][]key.Binding {
	km.Blur.SetHelp("esc", "Lock focus")
	return [][]key.Binding{
		{km.Blur, km.Next, km.NewPw, km.ShowPw, km.CopyPw},
		{km.Save, km.Delete, km.Back, km.Quit},
	}
}

func setupHeadless() {
	accountsPtr := &[]account{
		{Service: "New", Description: "New entry", Notes: "", User: "", Pw: ""},
	}

	*accountsPtr = append(*accountsPtr, *getAllAccounts()...)

	columns := []table.Column{
		{Title: "ID", Width: 3},
		{Title: "Service", Width: 10},
		{Title: "Description", Width: 20},
		{Title: "Notes", Width: 10},
	}

	rows := []table.Row{}

	for index, account := range *accountsPtr {
		rows = append(rows, table.Row{
			fmt.Sprint(index),
			account.Service,
			account.Description,
			account.Notes,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("128")).
		Background(lipgloss.Color("234")).
		Bold(false)
	t.SetStyles(s)

	km := customKeyMap{
		Blur: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "Focus"),
		),
		LineUp: key.NewBinding(
			key.WithKeys("w", "up"),
			key.WithHelp("w/↑", "Up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("s", "down"),
			key.WithHelp("s/↓", "Down"),
		),
		Select: key.NewBinding(
			key.WithKeys("tab", "enter"),
			key.WithHelp("tab/enter", "Select"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "Quit"),
		),
		CopyPw: key.NewBinding(
			key.WithKeys("ctrl+q"),
			key.WithHelp("ctrl+q", "Copy password"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "Go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "Go to end"),
		),
	}

	selKm := customSelKeyMap{
		Blur:   km.Blur,
		Next:   km.Select,
		CopyPw: km.CopyPw,
		Quit:   km.Quit,
		NewPw: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "Generate new password"),
		),
		ShowPw: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "Show / hide password"),
		),
		Back: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "Leave this menu without saving"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "Save changes and leave"),
		),
		Delete: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "Delete this entry"),
		),
	}

	tiServ := textinput.New()
	tiServ.Placeholder = "Google"

	tiDesc := textinput.New()
	tiDesc.Placeholder = "Some description"

	taNotes := textarea.New()
	taNotes.Placeholder = "Some additional notes"
	taNotes.FocusedStyle.CursorLineNumber = focusedStyle

	tiUser := textinput.New()
	tiUser.Placeholder = "Username / Email address"

	tiPw := textinput.New()
	tiPw.Placeholder = "Password"
	tiPw.EchoMode = textinput.EchoPassword
	tiPw.EchoCharacter = '•'

	m := model{t, tiServ, tiDesc, taNotes, tiUser, tiPw, km, selKm, help.New(), help.New(), accountsPtr, nil, 0}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
