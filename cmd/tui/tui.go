// Package tui implements an interactive terminal UI for bip32-ssh-keygen.
package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/valerius21/bip32-ssh-keygen/internal/keygen"
	"github.com/valerius21/bip32-ssh-keygen/internal/mnemonic"
	"github.com/valerius21/bip32-ssh-keygen/internal/path"
	"github.com/valerius21/bip32-ssh-keygen/internal/slip10"
)

// screen represents the current UI screen.
type screen int

const (
	screenMenu screen = iota
	screenGenerate
	screenDeriveInput
	screenDerivePath
	screenDeriveOutput
	screenDeriveConfirm
	screenResult
)

// model represents the application state.
type model struct {
	screen      screen
	width       int
	height      int
	cursor      int
	menuItems   []string

// Generate state
	generatedMnemonic string
	wordCount         int
	showDerivePrompt  bool // Tracks if we're showing the derive prompt on generate screen

	// Derive state
	mnemonicInput     textinput.Model
	pathInput         textinput.Model
	outputInput       textinput.Model
	passphraseInput   textinput.Model
	mnemonic          string
	derivePath        string
	outputPath        string
	passphrase        string

	// Result state
	resultMessage string
	resultError   error
	fingerprint   string
}

// Styles
type styles struct {
	title       lipgloss.Style
	subtitle    lipgloss.Style
	menuItem    lipgloss.Style
	menuCursor  lipgloss.Style
	mnemonic    lipgloss.Style
	errorStyle  lipgloss.Style
	success     lipgloss.Style
	help        lipgloss.Style
}

func newStyles() styles {
	return styles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1),
		subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B8B8B8")).
			MarginBottom(1),
		menuItem: lipgloss.NewStyle().
			PaddingLeft(2),
		menuCursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true).
			PaddingRight(1).
			SetString(">"),
		mnemonic: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Background(lipgloss.Color("#0D1117")).
			Padding(1).
			MarginTop(1).
			MarginBottom(1),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")),
		success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
		help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
	}
}

func initialModel() model {
	// Setup mnemonic input
	mnInput := textinput.New()
	mnInput.Placeholder = "Enter your 12-24 word mnemonic..."
	mnInput.Focus()
	mnInput.Width = 60

	// Setup path input
	pInput := textinput.New()
	pInput.Placeholder = path.DefaultPath
	pInput.SetValue(path.DefaultPath)
	pInput.Width = 60

	// Setup output input
	oInput := textinput.New()
	oInput.Placeholder = "id_ed25519"
	oInput.SetValue("id_ed25519")
	oInput.Width = 60

	// Setup passphrase input
	passInput := textinput.New()
	passInput.Placeholder = "Optional passphrase"
	passInput.EchoMode = textinput.EchoPassword
	passInput.Width = 60

	return model{
		screen:    screenMenu,
		menuItems: []string{"Generate new mnemonic", "Derive SSH key from mnemonic", "Quit"},
		wordCount: 24,
		mnemonicInput:   mnInput,
		pathInput:       pInput,
		outputInput:     oInput,
		passphraseInput: passInput,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.screen == screenMenu {
				return m, tea.Quit
			}
			m.screen = screenMenu
			return m, nil

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyUp, tea.KeyDown:
			if m.screen == screenMenu {
				if msg.Type == tea.KeyUp {
					m.cursor--
					if m.cursor < 0 {
						m.cursor = len(m.menuItems) - 1
					}
				} else {
					m.cursor++
					if m.cursor >= len(m.menuItems) {
						m.cursor = 0
					}
				}
			}
		case tea.KeyRunes:
			// Handle y/n on generate screen with derive prompt
			if m.screen == screenGenerate && m.showDerivePrompt {
				runes := msg.Runes
				if len(runes) > 0 {
					switch runes[0] {
					case 'y', 'Y':
						// Yes: transition to derive with pre-filled mnemonic
						m.screen = screenDeriveInput
						m.mnemonicInput.SetValue(m.generatedMnemonic)
						m.mnemonic = m.generatedMnemonic
						m.mnemonicInput.Focus()
						return m, nil
					case 'n', 'N':
						// No: return to menu
						m.screen = screenMenu
						return m, nil
					}
				}
			}
		} // closes case tea.KeyMsg (switch msg.Type)
	} // closes switch msg.(type)

	// Update text inputs based on current screen
	switch m.screen {
	case screenDeriveInput:
		m.mnemonicInput, cmd = m.mnemonicInput.Update(msg)
		cmds = append(cmds, cmd)
	case screenDerivePath:
		m.pathInput, cmd = m.pathInput.Update(msg)
		cmds = append(cmds, cmd)
	case screenDeriveOutput:
		m.outputInput, cmd = m.outputInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenMenu:
		switch m.cursor {
			case 0: // Generate
				m.screen = screenGenerate
				mn, err := mnemonic.Generate(m.wordCount)
				if err != nil {
					m.resultError = err
					m.screen = screenResult
					return m, nil
				}
				m.generatedMnemonic = mn
				m.showDerivePrompt = false // Show mnemonic first
				return m, nil
			case 1: // Derive
			m.screen = screenDeriveInput
			m.mnemonicInput.Focus()
			return m, nil
		case 2: // Quit
			return m, tea.Quit
		}

case screenGenerate:
		if !m.showDerivePrompt {
			// First Enter press: show yes/no prompt
			m.showDerivePrompt = true
			return m, nil
		} else {
			// Second Enter press: default yes, transition to derive
			m.screen = screenDeriveInput
			m.mnemonicInput.SetValue(m.generatedMnemonic)
			m.mnemonic = m.generatedMnemonic
			m.mnemonicInput.Focus()
			return m, nil
		}

	case screenDeriveInput:
		m.mnemonic = strings.TrimSpace(m.mnemonicInput.Value())
		if m.mnemonic == "" {
			// Auto-generate if empty
			mn, err := mnemonic.Generate(24)
			if err != nil {
				m.resultError = err
				m.screen = screenResult
				return m, nil
			}
			m.mnemonic = mn
			m.generatedMnemonic = mn
		}
		if err := mnemonic.Validate(m.mnemonic); err != nil {
			m.resultError = fmt.Errorf("invalid mnemonic: %w", err)
			m.screen = screenResult
			return m, nil
		}
		m.screen = screenDerivePath
		m.pathInput.Focus()
		return m, nil

	case screenDerivePath:
		m.derivePath = m.pathInput.Value()
		if m.derivePath == "" {
			m.derivePath = path.DefaultPath
		}
		if _, err := path.Parse(m.derivePath); err != nil {
			m.resultError = fmt.Errorf("invalid path: %w", err)
			m.screen = screenResult
			return m, nil
		}
		m.screen = screenDeriveOutput
		m.outputInput.Focus()
		return m, nil

	case screenDeriveOutput:
		m.outputPath = m.outputInput.Value()
		if m.outputPath == "" {
			m.outputPath = "id_ed25519"
		}
		m.performDerivation()
		return m, nil

	case screenResult:
		m.screen = screenMenu
		m.resultError = nil
		m.resultMessage = ""
		m.generatedMnemonic = ""
		m.mnemonicInput.SetValue("")
		m.pathInput.SetValue(path.DefaultPath)
		m.outputInput.SetValue("id_ed25519")
		return m, nil
	}

	return m, nil
}

func (m *model) performDerivation() {
	seed := mnemonic.ToSeed(m.mnemonic, m.passphrase)

	indices, err := path.Parse(m.derivePath)
	if err != nil {
		m.resultError = err
		m.screen = screenResult
		return
	}

	key, err := slip10.NewMasterKey(seed)
	if err != nil {
		m.resultError = fmt.Errorf("failed to create master key: %w", err)
		m.screen = screenResult
		return
	}

	for _, index := range indices {
		key, err = key.DeriveChild(index)
		if err != nil {
			m.resultError = fmt.Errorf("failed to derive child key: %w", err)
			m.screen = screenResult
			return
		}
	}

	pair, err := keygen.Generate(key.PrivateKey, m.derivePath)
	if err != nil {
		m.resultError = fmt.Errorf("failed to generate SSH key: %w", err)
		m.screen = screenResult
		return
	}

	if err := keygen.WriteKeyPair(pair, m.outputPath, false); err != nil {
		m.resultError = err
		m.screen = screenResult
		return
	}

	m.fingerprint = pair.Fingerprint
	m.resultMessage = fmt.Sprintf("SSH key saved to %s", m.outputPath)
	m.screen = screenResult
}

func (m model) View() string {
	s := newStyles()

	var content string
	switch m.screen {
	case screenMenu:
		content = m.viewMenu(s)
	case screenGenerate:
		content = m.viewGenerate(s)
	case screenDeriveInput:
		content = m.viewDeriveInput(s)
	case screenDerivePath:
		content = m.viewDerivePath(s)
	case screenDeriveOutput:
		content = m.viewDeriveOutput(s)
	case screenResult:
		content = m.viewResult(s)
	}

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m model) viewMenu(s styles) string {
	var b strings.Builder

	b.WriteString(s.title.Render("🔐 bip32-ssh-keygen") + "\n")
	b.WriteString(s.subtitle.Render("Generate deterministic SSH keys from BIP39 mnemonics") + "\n\n")

	for i, item := range m.menuItems {
		cursor := "  "
		if m.cursor == i {
			cursor = s.menuCursor.String()
		}
		b.WriteString(cursor + s.menuItem.Render(item) + "\n")
	}

	b.WriteString("\n" + s.help.Render("↑/↓ navigate • enter select • esc quit"))

	return b.String()
}

func (m model) viewGenerate(s styles) string {
	var b strings.Builder

	b.WriteString(s.title.Render("📝 Generated Mnemonic") + "\n\n")

	// Split mnemonic into words and display in 4x6 grid
	words := strings.Split(m.generatedMnemonic, " ")
	for i := 0; i < len(words); i += 4 {
		end := i + 4
		if end > len(words) {
			end = len(words)
		}
		row := words[i:end]
		b.WriteString(strings.Join(row, " ") + "\n")
	}

	b.WriteString("\n" + s.subtitle.Render("Write down these words and store them securely.") + "\n")

	if m.showDerivePrompt {
		b.WriteString("\n" + s.help.Render("Generate SSH key from this mnemonic? [Y/n]: "))
	} else {
		b.WriteString("\n" + s.help.Render("press enter to continue"))
	}

	return b.String()
}

func (m model) viewDeriveInput(s styles) string {
	var b strings.Builder

	b.WriteString(s.title.Render("🔑 Enter Mnemonic") + "\n\n")
	b.WriteString("Enter your BIP39 mnemonic (12-24 words):\n")
	b.WriteString("(Leave empty to auto-generate a new mnemonic)\n\n")
	b.WriteString(m.mnemonicInput.View() + "\n\n")
	b.WriteString(s.help.Render("enter continue • esc cancel"))

	return b.String()
}

func (m model) viewDerivePath(s styles) string {
	var b strings.Builder

	b.WriteString(s.title.Render("📍 Derivation Path") + "\n\n")
	b.WriteString("Enter the derivation path (must use hardened indices):\n\n")
	b.WriteString(m.pathInput.View() + "\n\n")
	b.WriteString(s.help.Render("enter continue • esc cancel"))

	return b.String()
}

func (m model) viewDeriveOutput(s styles) string {
	var b strings.Builder

	b.WriteString(s.title.Render("💾 Output Location") + "\n\n")
	b.WriteString("Enter the output path for the SSH key:\n\n")
	b.WriteString(m.outputInput.View() + "\n\n")
	b.WriteString(s.help.Render("enter to derive key • esc cancel"))

	return b.String()
}

func (m model) viewResult(s styles) string {
	var b strings.Builder

	if m.resultError != nil {
		b.WriteString(s.title.Render("❌ Error") + "\n\n")
		b.WriteString(s.errorStyle.Render(m.resultError.Error()) + "\n\n")
	} else {
		b.WriteString(s.title.Render("✅ Success") + "\n\n")
		if m.generatedMnemonic != "" && m.screen == screenResult {
			b.WriteString("Generated mnemonic:\n")
			b.WriteString(s.mnemonic.Render(m.generatedMnemonic) + "\n\n")
		}
		b.WriteString(s.success.Render(m.resultMessage) + "\n")
		if m.fingerprint != "" {
			b.WriteString(fmt.Sprintf("\nFingerprint: %s\n", m.fingerprint))
		}
		b.WriteString(fmt.Sprintf("\nPrivate key: %s\n", m.outputPath))
		b.WriteString(fmt.Sprintf("Public key:  %s.pub\n", m.outputPath))
	}

	b.WriteString("\n" + s.help.Render("press enter to return to menu"))

	return b.String()
}

// Cmd returns the tui command.
// The tui command launches an interactive terminal UI for generating
// mnemonics and deriving SSH keys.
func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch an interactive terminal UI",
		Long: `Launch an interactive terminal UI for bip32-ssh-keygen.

The TUI provides a menu-driven interface for:
  • Generating new BIP39 mnemonics
  • Deriving SSH keys from mnemonics
  • Configuring derivation paths and output locations

This is useful for users who prefer an interactive experience over
command-line flags.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := tea.NewProgram(initialModel(), tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
				return err
			}
			return nil
		},
	}
}
