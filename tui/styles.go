package tui

import "github.com/charmbracelet/lipgloss"

var (
	// ── Core Text Styles ──
	TitleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	HeaderStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	MutedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	HelpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	WarningStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("202"))
	SelectedMenuStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	MenuItemStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	BoxStyle          = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)

	// ── Dashboard KPI Boxes ──
	IncomeBoxStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Foreground(lipgloss.Color("42")).Bold(true)
	ExpenseBoxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Foreground(lipgloss.Color("196")).Bold(true)
	BalanceBoxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Foreground(lipgloss.Color("39")).Bold(true)
	TableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69")).Background(lipgloss.Color("237"))
	TableRowStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	// ── Symbol Styles ──
	IncomeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	ExpenseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	// ── Pane-Based Layout Styles ──
	SidebarStyle = lipgloss.NewStyle().
			Width(22).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 1).
			MarginRight(1)

	MainPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			Bold(true)

	SearchBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	// ── Anomaly & Prediction Styles ──
	AnomalyBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("202")).
			Padding(0, 1).
			Foreground(lipgloss.Color("202"))

	PredictionBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("177")).
				Padding(0, 1).
				Foreground(lipgloss.Color("177"))

	// ── KPI Compact Styles (for horizontal row) ──
	KPILabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Bold(false)

	KPIValueStyle = lipgloss.NewStyle().
			Bold(true)

	KPITrendUpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	KPITrendDownStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	// ── Active/Inactive Sidebar Items ──
	SidebarActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("62")).
				Padding(0, 1).
				Width(18)

	SidebarInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Padding(0, 1).
				Width(18)
)
