package tui

import (
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestInitialModel(t *testing.T) {
	m := initialModel()

	assert.Equal(t, screenMenu, m.screen)
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, 3, len(m.menuItems))
	assert.Equal(t, 24, m.wordCount)
	assert.NotNil(t, m.mnemonicInput)
	assert.NotNil(t, m.pathInput)
	assert.NotNil(t, m.outputInput)
	assert.NotNil(t, m.passphraseInput)
}

func TestModelInit(t *testing.T) {
	m := initialModel()
	cmd := m.Init()
	assert.Nil(t, cmd)
}

func TestCmdFunction(t *testing.T) {
	cmd := Cmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "tui", cmd.Name())
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestHandleEnter_Menu(t *testing.T) {
	m := initialModel()

	// Test selecting Generate (cursor = 0)
	m.cursor = 0
	newM, cmd := m.handleEnter()
	assert.Equal(t, screenGenerate, newM.(model).screen)
	assert.NotEmpty(t, newM.(model).generatedMnemonic)

	// Reset and test Derive (cursor = 1)
	m = initialModel()
	m.cursor = 1
	newM, cmd = m.handleEnter()
	assert.Equal(t, screenDeriveInput, newM.(model).screen)
	assert.Nil(t, cmd)

	// Reset and test Quit (cursor = 2)
	m = initialModel()
	m.cursor = 2
	newM, cmd = m.handleEnter()
		assert.NotNil(t, cmd)
}

func TestHandleEnter_Generate(t *testing.T) {
	m := initialModel()
	m.screen = screenGenerate
	newM, _ := m.handleEnter()
	assert.Equal(t, screenMenu, newM.(model).screen)
}

func TestHandleEnter_DeriveInput_Empty(t *testing.T) {
	m := initialModel()
	m.screen = screenDeriveInput
	m.mnemonicInput.SetValue("")
	newM, _ := m.handleEnter()
	// Should auto-generate mnemonic
	assert.Equal(t, screenDerivePath, newM.(model).screen)
	assert.NotEmpty(t, newM.(model).mnemonic)
	assert.NotEmpty(t, newM.(model).generatedMnemonic)
}

func TestHandleEnter_DeriveInput_WithValue(t *testing.T) {
	m := initialModel()
	m.screen = screenDeriveInput
	m.mnemonicInput.SetValue("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about")
	newM, _ := m.handleEnter()
	assert.Equal(t, screenDerivePath, newM.(model).screen)
	assert.Equal(t, "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about", newM.(model).mnemonic)
}

func TestHandleEnter_DeriveInput_InvalidMnemonic(t *testing.T) {
	m := initialModel()
	m.screen = screenDeriveInput
	m.mnemonicInput.SetValue("invalid mnemonic words")
	newM, _ := m.handleEnter()
	assert.Equal(t, screenResult, newM.(model).screen)
	assert.Error(t, newM.(model).resultError)
}

func TestHandleEnter_DerivePath(t *testing.T) {
	m := initialModel()
	m.screen = screenDerivePath
	m.mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	m.pathInput.SetValue("m/44'/22'/0'/0'")
	newM, _ := m.handleEnter()
	assert.Equal(t, screenDeriveOutput, newM.(model).screen)
	assert.Equal(t, "m/44'/22'/0'/0'", newM.(model).derivePath)
}

func TestHandleEnter_DerivePath_Invalid(t *testing.T) {
	m := initialModel()
	m.screen = screenDerivePath
	m.pathInput.SetValue("m/44/22/0/0")
	newM, _ := m.handleEnter()
	assert.Equal(t, screenResult, newM.(model).screen)
	assert.Error(t, newM.(model).resultError)
}

func TestHandleEnter_DeriveOutput(t *testing.T) {
	m := initialModel()
	m.screen = screenDeriveOutput
	m.mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	m.derivePath = "m/44'/22'/0'/0'"
	m.outputInput.SetValue("test_key")
	newM, _ := m.handleEnter()
	// Should perform derivation and go to result
	assert.Equal(t, screenResult, newM.(model).screen)
	assert.NoError(t, newM.(model).resultError)
	assert.NotEmpty(t, newM.(model).resultMessage)
}

func TestHandleEnter_Result(t *testing.T) {
	m := initialModel()
	m.screen = screenResult
	m.resultError = nil
	m.resultMessage = "test"
	newM, _ := m.handleEnter()
	assert.Equal(t, screenMenu, newM.(model).screen)
	assert.Empty(t, newM.(model).resultError)
	assert.Empty(t, newM.(model).resultMessage)
}

func TestUpdate_WindowSize(t *testing.T) {
	m := initialModel()
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newM, _ := m.Update(msg)
	assert.Equal(t, 100, newM.(model).width)
	assert.Equal(t, 50, newM.(model).height)
}

func TestUpdate_KeyUp(t *testing.T) {
	m := initialModel()
	m.cursor = 1
	msg := tea.KeyMsg{Type: tea.KeyUp}
	newM, _ := m.Update(msg)
	assert.Equal(t, 0, newM.(model).cursor)
}

func TestUpdate_KeyDown(t *testing.T) {
	m := initialModel()
	m.cursor = 1
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newM, _ := m.Update(msg)
	assert.Equal(t, 2, newM.(model).cursor)
}

func TestUpdate_KeyDown_Wrap(t *testing.T) {
	m := initialModel()
	m.cursor = 2 // Last item
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newM, _ := m.Update(msg)
	assert.Equal(t, 0, newM.(model).cursor)
}

func TestUpdate_KeyUp_Wrap(t *testing.T) {
	m := initialModel()
	m.cursor = 0 // First item
	msg := tea.KeyMsg{Type: tea.KeyUp}
	newM, _ := m.Update(msg)
	assert.Equal(t, 2, newM.(model).cursor)
}

func TestUpdate_KeyEsc_FromMenu(t *testing.T) {
	m := initialModel()
	m.screen = screenMenu
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := m.Update(msg)
		assert.NotNil(t, cmd)
}

func TestUpdate_KeyEsc_FromOtherScreen(t *testing.T) {
	m := initialModel()
	m.screen = screenGenerate
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newM, _ := m.Update(msg)
	assert.Equal(t, screenMenu, newM.(model).screen)
}

func TestNewStyles(t *testing.T) {
	s := newStyles()
	assert.NotZero(t, s.title)
	assert.NotZero(t, s.subtitle)
	assert.NotZero(t, s.menuItem)
	assert.NotZero(t, s.menuCursor)
	assert.NotZero(t, s.mnemonic)
	assert.NotZero(t, s.errorStyle)
	assert.NotZero(t, s.success)
	assert.NotZero(t, s.help)
}
