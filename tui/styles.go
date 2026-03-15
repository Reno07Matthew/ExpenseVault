package tui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	HeaderStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	MutedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	HelpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	WarningStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("202"))
	SelectedMenuStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	MenuItemStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	BoxStyle          = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)

	// Dashboard box styles
	IncomeBoxStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Foreground(lipgloss.Color("42")).Bold(true)
	ExpenseBoxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Foreground(lipgloss.Color("196")).Bold(true)
	BalanceBoxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Foreground(lipgloss.Color("39")).Bold(true)
	TableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69")).Background(lipgloss.Color("237"))
	TableRowStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	// Symbol styles
	IncomeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	ExpenseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
)
