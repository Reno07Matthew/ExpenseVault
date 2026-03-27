package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"expenseVault/db"
	"expenseVault/models"
	"expenseVault/services"
	"expenseVault/utils"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/crypto/bcrypt"
)

type View int

const (
	ViewAuthMenu View = iota
	ViewSignup
	ViewLogin
	ViewDashboard
	ViewTransactions
	ViewAddForm
	ViewReports
	ViewAsk
)

const (
	fieldType = iota
	fieldAmount
	fieldCategory
	fieldDescription
	fieldDate
	fieldNotes
	fieldCount
)

type Model struct {
	store        *db.Store
	config       *utils.Config
	view         View
	cursor       int
	transactions []models.Transaction
	menuItems    []string
	message      string
	width        int
	height       int
	reportType   int
	quitting     bool

	currentUser *models.User

	authMenuItems    []string
	authUserInput    textinput.Model
	authPassInput    textinput.Model
	authConfirmInput textinput.Model
	authFocusIndex   int
	authMessage      string

	inputs      []textinput.Model
	focusIndex  int
	formMessage string

	dashData models.DashboardData

	anomalyAlerts []services.AnomalyAlert
	spendingTrend string
	prediction    string

	searchInput textinput.Model
	searchMode  bool   // true when / search is active
	queryMode   bool   // true when : CQL query is active
	searchQuery string
	filteredTxs []models.Transaction

	askInput  textinput.Model
	askAnswer string
	isAsking  bool
}

func NewModel(store *db.Store, config *utils.Config) Model {
	inputs := make([]textinput.Model, fieldCount)

	inputs[fieldType] = textinput.New()
	inputs[fieldType].Placeholder = "income or expense"
	inputs[fieldType].CharLimit = 10
	inputs[fieldType].Width = 30
	inputs[fieldType].Prompt = "Type: "

	inputs[fieldAmount] = textinput.New()
	inputs[fieldAmount].Placeholder = "e.g. 250.00"
	inputs[fieldAmount].CharLimit = 15
	inputs[fieldAmount].Width = 30
	inputs[fieldAmount].Prompt = "Amount: "

	inputs[fieldCategory] = textinput.New()
	inputs[fieldCategory].Placeholder = "Food, Travel, Bills..."
	inputs[fieldCategory].CharLimit = 20
	inputs[fieldCategory].Width = 30
	inputs[fieldCategory].Prompt = "Category: "

	inputs[fieldDescription] = textinput.New()
	inputs[fieldDescription].Placeholder = "Description"
	inputs[fieldDescription].CharLimit = 50
	inputs[fieldDescription].Width = 40
	inputs[fieldDescription].Prompt = "Description: "

	inputs[fieldDate] = textinput.New()
	inputs[fieldDate].Placeholder = time.Now().Format("2006-01-02")
	inputs[fieldDate].CharLimit = 10
	inputs[fieldDate].Width = 30
	inputs[fieldDate].Prompt = "Date: "

	inputs[fieldNotes] = textinput.New()
	inputs[fieldNotes].Placeholder = "(optional)"
	inputs[fieldNotes].CharLimit = 50
	inputs[fieldNotes].Width = 40
	inputs[fieldNotes].Prompt = "Notes: "

	inputs[fieldType].Focus()

	authUser := textinput.New()
	authUser.Placeholder = "Username"
	authUser.CharLimit = 50
	authUser.Width = 30
	authUser.Prompt = "Username: "

	authPass := textinput.New()
	authPass.Placeholder = "Password"
	authPass.CharLimit = 72
	authPass.Width = 30
	authPass.Prompt = "Password: "
	authPass.EchoMode = textinput.EchoPassword
	authPass.EchoCharacter = '*'

	authConfirm := textinput.New()
	authConfirm.Placeholder = "Confirm Password"
	authConfirm.CharLimit = 72
	authConfirm.Width = 30
	authConfirm.Prompt = "Confirm: "
	authConfirm.EchoMode = textinput.EchoPassword
	authConfirm.EchoCharacter = '*'

	authUser.Focus()

	askIn := textinput.New()
	askIn.Placeholder = "Ask a question..."
	askIn.CharLimit = 255
	askIn.Width = 60
	askIn.Prompt = "🤖 ? "

	searchIn := textinput.New()
	searchIn.Placeholder = "Search transactions..."
	searchIn.CharLimit = 100
	searchIn.Width = 50
	searchIn.Prompt = "🔍 "

	return Model{
		store:  store,
		config: config,
		view:   ViewAuthMenu,
		authMenuItems: []string{
			"Sign up",
			"Log in",
			"Exit",
		},
		authUserInput:    authUser,
		authPassInput:    authPass,
		authConfirmInput: authConfirm,
		menuItems: []string{
			"Dashboard",
			"Transactions",
			"Add Transaction",
			"Reports",
			"Ask AI",
			"Exit",
		},
		inputs:      inputs,
		focusIndex:  0,
		askInput:    askIn,
		searchInput: searchIn,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) loadTransactions() tea.Msg {
	if m.currentUser == nil {
		return txsLoadedMsg{[]models.Transaction{}, models.DashboardData{}, nil, "", ""}
	}
	txs, err := m.store.GetAllTransactions(m.currentUser.ID)
	if err != nil {
		return errMsg{err}
	}
	
	month := time.Now().Format("2006-01")
	budgets, err := m.store.GetBudgets(m.currentUser.ID, month)
	if err != nil {
		// If budgets fail, we still want to show transactions
		budgets = make(map[models.Category]models.Rupees)
	}

	dashData := services.CalculateDashboardData(txs, budgets)
	anomalies := services.DetectAnomalies(txs)
	trend := services.GetSpendingTrend(txs)

	// Predictive budgeting.
	pred := services.PredictEndOfMonth(txs, dashData.MonthlyIncome.ToFloat64())
	predStr := services.FormatPrediction(pred)

	return txsLoadedMsg{txs, dashData, anomalies, trend, predStr}
}

type txsLoadedMsg struct {
	transactions  []models.Transaction
	dashData      models.DashboardData
	anomalyAlerts []services.AnomalyAlert
	spendingTrend string
	prediction    string
}
type errMsg struct{ err error }
type txAddedMsg struct{ id int64 }
type signupSuccessMsg struct{ username string }
type userLoggedInMsg struct{ user *models.User }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.view == ViewAuthMenu || m.view == ViewSignup || m.view == ViewLogin {
		return m.updateAuth(msg)
	}
	if m.view == ViewAddForm {
		return m.updateForm(msg)
	}
	if m.view == ViewAsk {
		return m.updateAsk(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case txsLoadedMsg:
		m.transactions = msg.transactions
		m.dashData = msg.dashData
		m.anomalyAlerts = msg.anomalyAlerts
		m.spendingTrend = msg.spendingTrend
		m.prediction = msg.prediction
		m.message = ""
		return m, nil
	case txAddedMsg:
		m.message = fmt.Sprintf("Transaction #%d added", msg.id)
		m.view = ViewDashboard
		m.cursor = 0
		return m, m.loadTransactions
	case errMsg:
		m.message = fmt.Sprintf("Error: %v", msg.err)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			max := len(m.menuItems) - 1
			if m.view == ViewTransactions {
				max = len(m.transactions) - 1
			}
			if m.cursor < max {
				m.cursor++
			}
		case "enter":
			return m.handleEnter()
		case "1":
			m.view = ViewDashboard
			m.cursor = 0
		case "2":
			m.view = ViewTransactions
			m.cursor = 0
		case "3":
			m.view = ViewAddForm
			m.resetForm()
			return m, m.inputs[0].Focus()
		case "4":
			m.view = ViewReports
			m.cursor = 0
		case "5":
			m.view = ViewAsk
			m.askInput.Focus()
			m.askAnswer = ""
			m.isAsking = false
		case "/":
			if !m.searchMode && !m.queryMode {
				m.searchMode = true
				m.searchInput.SetValue("")
				m.searchInput.Focus()
				return m, nil
			}
		case ":":
			if !m.searchMode && !m.queryMode {
				m.queryMode = true
				m.searchInput.SetValue("")
				m.searchInput.Prompt = "🔎 "
				m.searchInput.Placeholder = "cat:food amt:>500 date:last-week"
				m.searchInput.Focus()
				return m, nil
			}
		case "esc":
			if m.searchMode || m.queryMode {
				m.searchMode = false
				m.queryMode = false
				m.searchQuery = ""
				m.filteredTxs = nil
				m.searchInput.Blur()
				m.searchInput.Prompt = "🔍 "
				m.searchInput.Placeholder = "Search transactions..."
				return m, nil
			}
			m.view = ViewDashboard
			m.cursor = 0
			return m, nil
		}
	}

	// Handle search input updates.
	if m.searchMode || m.queryMode {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		// Live-filter on every keystroke.
		query := m.searchInput.Value()
		if query != "" {
			if m.queryMode {
				filter := services.ParseQuery(query)
				m.filteredTxs = services.ApplyQueryFilter(m.transactions, filter)
			} else {
				// Simple fuzzy search.
				filter := services.QueryFilter{FuzzyText: query}
				m.filteredTxs = services.ApplyQueryFilter(m.transactions, filter)
			}
			m.searchQuery = query
		} else {
			m.filteredTxs = nil
			m.searchQuery = ""
		}
		return m, cmd
	}

	return m, nil
}

func (m Model) updateAuth(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case signupSuccessMsg:
		m.authMessage = "Signup successful. Please log in."
		m.view = ViewLogin
		m.authPassInput.SetValue("")
		m.authConfirmInput.SetValue("")
		m.authFocusIndex = 1
		return m, m.updateAuthFocus()
	case userLoggedInMsg:
		m.currentUser = msg.user
		m.view = ViewDashboard
		m.cursor = 0
		m.message = fmt.Sprintf("Welcome, %s", msg.user.Username)
		return m, m.loadTransactions
	case errMsg:
		m.authMessage = fmt.Sprintf("Error: %v", msg.err)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			if m.view == ViewSignup || m.view == ViewLogin {
				m.view = ViewAuthMenu
				m.authMessage = ""
				m.cursor = 0
				m.resetAuthInputs()
				return m, nil
			}
			return m, nil
		}

		if m.view == ViewAuthMenu {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.authMenuItems)-1 {
					m.cursor++
				}
			case "enter":
				switch m.cursor {
				case 0:
					m.view = ViewSignup
					m.authMessage = ""
					m.resetAuthInputs()
					return m, m.updateAuthFocus()
				case 1:
					m.view = ViewLogin
					m.authMessage = ""
					m.resetAuthInputs()
					return m, m.updateAuthFocus()
				case 2:
					m.quitting = true
					return m, tea.Quit
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "tab", "down":
			m.authFocusIndex = (m.authFocusIndex + 1) % m.authFieldCount()
			return m, m.updateAuthFocus()
		case "shift+tab", "up":
			m.authFocusIndex = (m.authFocusIndex - 1 + m.authFieldCount()) % m.authFieldCount()
			return m, m.updateAuthFocus()
		case "enter":
			if m.view == ViewSignup {
				if m.authFocusIndex == m.authFieldCount()-1 {
					return m.submitSignup()
				}
				m.authFocusIndex = (m.authFocusIndex + 1) % m.authFieldCount()
				return m, m.updateAuthFocus()
			}
			if m.view == ViewLogin {
				if m.authFocusIndex == m.authFieldCount()-1 {
					return m.submitLogin()
				}
				m.authFocusIndex = (m.authFocusIndex + 1) % m.authFieldCount()
				return m, m.updateAuthFocus()
			}
		}
	}

	if m.view == ViewSignup || m.view == ViewLogin {
		var cmd tea.Cmd
		switch m.authFocusIndex {
		case 0:
			m.authUserInput, cmd = m.authUserInput.Update(msg)
		case 1:
			m.authPassInput, cmd = m.authPassInput.Update(msg)
		case 2:
			m.authConfirmInput, cmd = m.authConfirmInput.Update(msg)
		}
		return m, cmd
	}

	return m, nil
}

func (m *Model) resetAuthInputs() {
	m.authUserInput.SetValue("")
	m.authPassInput.SetValue("")
	m.authConfirmInput.SetValue("")
	m.authFocusIndex = 0
	m.updateAuthFocus()
}

func (m *Model) updateAuthFocus() tea.Cmd {
	count := m.authFieldCount()
	cmds := make([]tea.Cmd, 0, 3)

	m.authUserInput.Blur()
	m.authPassInput.Blur()
	m.authConfirmInput.Blur()

	if m.authFocusIndex == 0 {
		cmds = append(cmds, m.authUserInput.Focus())
	}
	if m.authFocusIndex == 1 && count >= 2 {
		cmds = append(cmds, m.authPassInput.Focus())
	}
	if m.authFocusIndex == 2 && count == 3 {
		cmds = append(cmds, m.authConfirmInput.Focus())
	}

	return tea.Batch(cmds...)
}

func (m Model) authFieldCount() int {
	if m.view == ViewSignup {
		return 3
	}
	return 2
}

func (m Model) submitSignup() (tea.Model, tea.Cmd) {
	username := strings.TrimSpace(m.authUserInput.Value())
	password := m.authPassInput.Value()
	confirm := m.authConfirmInput.Value()

	if username == "" || password == "" || confirm == "" {
		m.authMessage = "Username and passwords are required"
		return m, nil
	}
	if password != confirm {
		m.authMessage = "Passwords do not match"
		return m, nil
	}

	store := m.store
	return m, func() tea.Msg {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return errMsg{err}
		}
		_, err = store.CreateUser(username, string(hash))
		if err != nil {
			return errMsg{err}
		}
		return signupSuccessMsg{username: username}
	}
}

func (m Model) submitLogin() (tea.Model, tea.Cmd) {
	username := strings.TrimSpace(m.authUserInput.Value())
	password := m.authPassInput.Value()

	if username == "" || password == "" {
		m.authMessage = "Username and password are required"
		return m, nil
	}

	store := m.store
	return m, func() tea.Msg {
		user, err := store.GetUserByUsername(username)
		if err != nil {
			return errMsg{err}
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			return errMsg{fmt.Errorf("invalid username or password")}
		}
		_ = store.UpdateLastLogin(user.ID)
		return userLoggedInMsg{user: user}
	}
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case txsLoadedMsg:
		m.transactions = msg.transactions
		return m, nil
	case txAddedMsg:
		m.message = fmt.Sprintf("Transaction #%d added", msg.id)
		m.view = ViewDashboard
		m.cursor = 0
		return m, m.loadTransactions
	case errMsg:
		m.formMessage = fmt.Sprintf("Error: %v", msg.err)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			m.view = ViewDashboard
			m.cursor = 0
			m.formMessage = ""
			return m, nil
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % fieldCount
			return m, m.updateFocus()
		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + fieldCount) % fieldCount
			return m, m.updateFocus()
		case "enter":
			if m.focusIndex == fieldCount-1 {
				return m.submitForm()
			}
			m.focusIndex = (m.focusIndex + 1) % fieldCount
			return m, m.updateFocus()
		case "ctrl+s":
			return m.submitForm()
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *Model) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, fieldCount)
	for i := range m.inputs {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
		m.inputs[i].Blur()
	}
	m.focusIndex = 0
	m.formMessage = ""
	m.inputs[fieldDate].SetValue(time.Now().Format("2006-01-02"))
}

func (m Model) submitForm() (tea.Model, tea.Cmd) {
	txType := strings.TrimSpace(strings.ToLower(m.inputs[fieldType].Value()))
	amountStr := strings.TrimSpace(m.inputs[fieldAmount].Value())
	category := strings.TrimSpace(m.inputs[fieldCategory].Value())
	desc := strings.TrimSpace(m.inputs[fieldDescription].Value())
	date := strings.TrimSpace(m.inputs[fieldDate].Value())
	notes := strings.TrimSpace(m.inputs[fieldNotes].Value())

	if txType != "income" && txType != "expense" {
		m.formMessage = "Type must be income or expense"
		return m, nil
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		m.formMessage = "Amount must be a positive number"
		return m, nil
	}

	if desc == "" {
		m.formMessage = "Description is required"
		return m, nil
	}

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	cat := models.Category(category)
	if category == "" {
		cat = services.NewCategorizer().AutoCategorize(desc)
	}

	// LAB 4: Factory function returns *Transaction (pointer).
	tx := models.NewTransaction(
		m.currentUser.ID,
		models.TransactionType(txType),
		amount,
		cat,
		desc,
		date,
	)
	// LAB 4: Pointer receiver — mutates the transaction in-place.
	if notes != "" {
		tx.SetNotes(notes)
	}

	store := m.store
	return m, func() tea.Msg {
		// LAB 4: Dereference pointer to pass value to AddTransaction.
		id, err := store.AddTransaction(tx)
		if err != nil {
			return errMsg{err}
		}
		return txAddedMsg{id}
	}
}

func (m Model) updateAsk(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			m.view = ViewDashboard
			m.cursor = 0
			return m, nil
		case "enter":
			if m.isAsking {
				return m, nil
			}
			query := strings.TrimSpace(m.askInput.Value())
			if query == "" {
				return m, nil
			}
			
			m.isAsking = true
			m.askAnswer = "Thinking... ⏳"
			
			// Capture variables for the async command
			store := m.store
			userID := m.currentUser.ID
			
			return m, func() tea.Msg {
				sqlQuery, err := services.GenerateSQL(query, userID)
				if err != nil {
					return errMsg{fmt.Errorf("LLM Error: %w", err)}
				}
				
				results, err := store.ExecuteReadQuery(sqlQuery)
				if err != nil {
					return errMsg{fmt.Errorf("DB Error: %w", err)}
				}
				
				summary, err := services.SummarizeData(query, results)
				if err != nil {
					return errMsg{fmt.Errorf("Summarize Error: %w", err)}
				}
				
				return askResponseMsg{answer: summary}
			}
		}
	case askResponseMsg:
		m.isAsking = false
		m.askAnswer = msg.answer
		return m, nil
	case errMsg:
		m.isAsking = false
		m.askAnswer = fmt.Sprintf("Error: %v", msg.err)
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	var cmd tea.Cmd
	m.askInput, cmd = m.askInput.Update(msg)
	return m, cmd
}

type askResponseMsg struct {
	answer string
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	if m.view == ViewDashboard {
		switch m.cursor {
		case 0:
			return m, nil
		case 1:
			m.view = ViewTransactions
			m.cursor = 0
		case 2:
			m.view = ViewAddForm
			m.resetForm()
			return m, m.inputs[0].Focus()
		case 3:
			m.view = ViewReports
			m.cursor = 0
		case 4:
			m.view = ViewAsk
			m.askInput.Focus()
			m.askAnswer = ""
			m.isAsking = false
		case 5:
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye.\n"
	}

	var sb strings.Builder

	dbInfo := ""
	if m.config != nil {
		if m.config.DBType == "supabase" || m.config.DBType == "postgres" {
			dbInfo = "[Supabase]"
		} else if m.config.DBType == "mysql" {
			dbInfo = fmt.Sprintf("[MySQL %s:%d/%s]", m.config.MySQLHost, m.config.MySQLPort, m.config.MySQLDatabase)
		} else {
			dbInfo = "[SQLite]"
		}
	}

	// Auth views render fullscreen (no sidebar).
	if m.view == ViewAuthMenu || m.view == ViewSignup || m.view == ViewLogin {
		sb.WriteString(TitleStyle.Render("ExpenseVault " + dbInfo))
		sb.WriteString("\n\n")

		switch m.view {
		case ViewAuthMenu:
			sb.WriteString(m.renderAuthMenu())
		case ViewSignup:
			sb.WriteString(m.renderAuthForm("Sign up", true))
		case ViewLogin:
			sb.WriteString(m.renderAuthForm("Log in", false))
		}

		sb.WriteString("\n")
		if m.view == ViewAuthMenu {
			sb.WriteString(HelpStyle.Render("[Enter] Select  [Up/Down] Move  [q] Quit"))
		} else {
			sb.WriteString(HelpStyle.Render("[Tab] Next field  [Shift+Tab] Prev field  [Enter] Submit  [Esc] Back"))
		}

		if m.message != "" {
			sb.WriteString("\n")
			sb.WriteString(WarningStyle.Render(m.message))
		}
		return sb.String()
	}

	// ── Pane-Based Layout for Post-Login Views ──
	sb.WriteString(TitleStyle.Render("ExpenseVault " + dbInfo))
	if m.currentUser != nil {
		sb.WriteString(MutedStyle.Render("  👤 " + m.currentUser.Username))
	}
	sb.WriteString("\n\n")

	// Build sidebar
	sidebar := m.renderSidebar()

	// Build main content
	var mainContent string
	switch m.view {
	case ViewDashboard:
		mainContent = m.renderDashboard()
	case ViewTransactions:
		mainContent = m.renderTransactions()
	case ViewAddForm:
		mainContent = m.renderAddForm()
	case ViewReports:
		mainContent = m.renderReports()
	case ViewAsk:
		mainContent = m.renderAsk()
	}

	mainWidth := m.width - 28
	if mainWidth < 40 {
		mainWidth = 40
	}
	mainPane := MainPaneStyle.Width(mainWidth).Render(mainContent)

	// Join sidebar + main horizontally
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainPane))
	sb.WriteString("\n")

	// Status bar
	sb.WriteString(m.renderStatusBar())

	if m.message != "" {
		sb.WriteString("\n")
		sb.WriteString(WarningStyle.Render(m.message))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderSidebar renders the fixed left navigation pane.
func (m Model) renderSidebar() string {
	var sb strings.Builder

	sb.WriteString(HeaderStyle.Render("📂 NAVIGATION"))
	sb.WriteString("\n\n")

	icons := []string{"📊", "📋", "➕", "📈", "🤖", "🚪"}

	for i, item := range m.menuItems {
		label := fmt.Sprintf(" %s %s", icons[i], item)
		if m.view == ViewDashboard+View(i) {
			sb.WriteString(SidebarActiveStyle.Render(label))
		} else {
			sb.WriteString(SidebarInactiveStyle.Render(label))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(MutedStyle.Render("─────────────────"))
	sb.WriteString("\n\n")

	// Quick stats
	if m.currentUser != nil && len(m.transactions) > 0 {
		sb.WriteString(MutedStyle.Render(fmt.Sprintf(" 📝 %d txns", len(m.transactions))))
		sb.WriteString("\n")
	}

	return SidebarStyle.Render(sb.String())
}

// renderStatusBar renders the bottom bar with hotkeys & search.
func (m Model) renderStatusBar() string {
	if m.searchMode || m.queryMode {
		mode := "SEARCH"
		if m.queryMode {
			mode = "QUERY"
		}
		modeTag := StatusBarStyle.Render(fmt.Sprintf(" %s ", mode))
		searchView := SearchBarStyle.Width(m.width - 15).Render(m.searchInput.View())
		return lipgloss.JoinHorizontal(lipgloss.Center, modeTag, searchView)
	}

	hotkeys := " [1]Dash [2]Txns [3]Add [4]Reports [5]AI  [/]Search [:]Query [q]Quit "
	return StatusBarStyle.Width(m.width).Render(hotkeys)
}

func (m Model) renderAuthMenu() string {
	var sb strings.Builder
	if m.authMessage != "" {
		sb.WriteString(WarningStyle.Render(m.authMessage))
		sb.WriteString("\n\n")
	}
	for i, item := range m.authMenuItems {
		if i == m.cursor {
			sb.WriteString(SelectedMenuStyle.Render("> " + item))
		} else {
			sb.WriteString(MenuItemStyle.Render("  " + item))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (m Model) renderAuthForm(title string, includeConfirm bool) string {
	var sb strings.Builder
	sb.WriteString(HeaderStyle.Render(title))
	sb.WriteString("\n\n")

	fields := []string{
		m.authUserInput.View(),
		m.authPassInput.View(),
	}
	if includeConfirm {
		fields = append(fields, m.authConfirmInput.View())
	}

	sb.WriteString(BoxStyle.Render(strings.Join(fields, "\n\n")))

	if m.authMessage != "" {
		sb.WriteString("\n\n")
		sb.WriteString(WarningStyle.Render(m.authMessage))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderAddForm() string {
	var sb strings.Builder

	sb.WriteString(HeaderStyle.Render("  ➕ Add New Transaction"))
	sb.WriteString("\n\n")

	sb.WriteString(BoxStyle.Render(
		fmt.Sprintf(
			"Fill in the details below and press Ctrl+S to save.\n\n"+
				"%s\n\n"+
				"%s\n\n"+
				"%s\n\n"+
				"%s\n\n"+
				"%s\n\n"+
				"%s",
			m.fieldView(fieldType, "Type", "income / expense"),
			m.fieldView(fieldAmount, "Amount", "in ₹"),
			m.fieldView(fieldCategory, "Category", "leave empty for auto-detect"),
			m.fieldView(fieldDescription, "Description", "what was it for?"),
			m.fieldView(fieldDate, "Date", "YYYY-MM-DD"),
			m.fieldView(fieldNotes, "Notes", "optional extra info"),
		),
	))

	// Available categories hint
	sb.WriteString("\n\n")
	sb.WriteString(MutedStyle.Render("  📂 Categories: Food, Travel, Shopping, Bills, Health, Education, Entertainment, Salary, Freelance, Other"))
	sb.WriteString("\n")

	if m.formMessage != "" {
		sb.WriteString("\n")
		sb.WriteString(WarningStyle.Render("  " + m.formMessage))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderAsk() string {
	var sb strings.Builder

	sb.WriteString(HeaderStyle.Render("  🤖 Ask AI"))
	sb.WriteString("\n\n")

	sb.WriteString("  " + MutedStyle.Render("Query your finances using natural language (e.g., 'What did I spend the most on last month?')"))
	sb.WriteString("\n\n")

	sb.WriteString("  " + m.askInput.View())
	sb.WriteString("\n\n")

	if m.askAnswer != "" {
		ansBox := BoxStyle.Width(m.width - 10).Render(m.askAnswer)
		sb.WriteString(ansBox)
		sb.WriteString("\n")
	}

	return sb.String()
}

// fieldView renders a single form field with its label.
func (m Model) fieldView(index int, label string, hint string) string {
	focusIndicator := "  "
	style := MutedStyle
	if m.focusIndex == index {
		focusIndicator = "▸ "
		style = HeaderStyle
	}

	return fmt.Sprintf("%s%s %s\n%s",
		focusIndicator,
		style.Render(label),
		MutedStyle.Render("("+hint+")"),
		m.inputs[index].View(),
	)
}

func (m Model) renderDashboard() string {
	var sb strings.Builder

	// Summary boxes
	label := "💰 Income"
	if m.dashData.UsingBudget {
		label = "🎯 Budgeted"
	}
	incomeBox := IncomeBoxStyle.Render(fmt.Sprintf("%s\n%s", label, m.dashData.MonthlyIncome))
	expenseBox := ExpenseBoxStyle.Render(fmt.Sprintf("💸 Expenses\n%s", m.dashData.TotalExpenses))
	balanceBox := BalanceBoxStyle.Render(fmt.Sprintf("📊 Savings\n%s", m.dashData.Savings))

	sb.WriteString("  " + lipgloss.JoinHorizontal(lipgloss.Top, incomeBox, "  ", expenseBox, "  ", balanceBox))
	sb.WriteString("\n\n")

	// Progress Bars
	title := "  📊 Budget Overview"
	if m.dashData.UsingBudget {
		title = "  📊 Spending relative to Budget"
	}
	sb.WriteString(HeaderStyle.Render(title))
	sb.WriteString("\n\n")
	
	if !m.dashData.UsingBudget {
		sb.WriteString(m.renderProgressBar("Income  ", 1.0, 30) + " " + IncomeStyle.Render("100%"))
		sb.WriteString("\n")
	}

	sb.WriteString(m.renderProgressBar("Expenses", m.dashData.ExpenseRatio, 30) + " " + ExpenseStyle.Render(fmt.Sprintf("%.0f%%", m.dashData.ExpenseRatio*100)))
	sb.WriteString("\n")
	
	if !m.dashData.UsingBudget {
		sb.WriteString(m.renderProgressBar("Savings ", m.dashData.SavingsRatio, 30) + " " + BalanceBoxStyle.Copy().Border(lipgloss.HiddenBorder()).Padding(0).Render(fmt.Sprintf("%.0f%%", m.dashData.SavingsRatio*100)))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Category Breakdown
	if len(m.dashData.Breakdown) > 0 {
		sb.WriteString(HeaderStyle.Render("  📂 Category Targets"))
		sb.WriteString("\n\n")
		for _, b := range m.dashData.Breakdown {
			targetInfo := ""
			if b.Target > 0 {
				targetInfo = fmt.Sprintf(" / %s", b.Target)
			}
			sb.WriteString(fmt.Sprintf("  %-15s %s%s (%2.0f%%)\n", b.Category, b.Amount, targetInfo, b.Percent*100))
		}
		sb.WriteString("\n")
	}

	// Smart Insights & Tips
	if m.dashData.SmartInsight != "" {
		sb.WriteString(HeaderStyle.Render("  💡 Budget Insight"))
		sb.WriteString("\n\n")
		sb.WriteString(BoxStyle.Width(m.width - 10).Render(m.dashData.SmartInsight))
		sb.WriteString("\n")
	} else if m.dashData.DailyTip != "" {
		sb.WriteString(HeaderStyle.Render("  💡 Daily Budget Tip"))
		sb.WriteString("\n\n")
		sb.WriteString(BoxStyle.Width(m.width - 10).Render(m.dashData.DailyTip))
		sb.WriteString("\n")
	}

	// Anomaly Alerts
	if len(m.anomalyAlerts) > 0 {
		sb.WriteString("\n")
		sb.WriteString(HeaderStyle.Render("  🔍 Anomaly Detection"))
		sb.WriteString("\n\n")
		maxAlerts := 3
		if len(m.anomalyAlerts) < maxAlerts {
			maxAlerts = len(m.anomalyAlerts)
		}
		for _, alert := range m.anomalyAlerts[:maxAlerts] {
			sb.WriteString(WarningStyle.Render("  " + alert.Message))
			sb.WriteString("\n")
		}
	}

	// Spending Trend
	if m.spendingTrend != "" {
		sb.WriteString("\n")
		sb.WriteString(BoxStyle.Render(m.spendingTrend))
		sb.WriteString("\n")
	}

	// Predictive Budgeting
	if m.prediction != "" {
		sb.WriteString("\n")
		sb.WriteString(PredictionBoxStyle.Render("🔮 " + m.prediction))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderProgressBar(label string, percent float64, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}

	filledLength := int(percent * float64(width))
	if filledLength > width {
		filledLength = width
	}
	emptyLength := width - filledLength

	filled := strings.Repeat("█", filledLength)
	empty := strings.Repeat("░", emptyLength)

	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	if label == "Expenses" {
		if percent > 0.8 {
			barStyle = barStyle.Foreground(lipgloss.Color("196"))
		} else if percent > 0.5 {
			barStyle = barStyle.Foreground(lipgloss.Color("214"))
		}
	} else if label == "Savings " {
		barStyle = barStyle.Foreground(lipgloss.Color("42"))
	}

	return fmt.Sprintf("  %s %s", MutedStyle.Render(label), barStyle.Render(filled+empty))
}

func (m Model) renderTransactions() string {
	var sb strings.Builder

	// Determine which transactions to show.
	txs := m.transactions
	if m.filteredTxs != nil {
		txs = m.filteredTxs
		sb.WriteString(HeaderStyle.Render(fmt.Sprintf("  📋 Filtered: %d results", len(txs))))
	} else {
		sb.WriteString(HeaderStyle.Render("  📋 All Transactions"))
	}
	sb.WriteString("\n\n")

	if len(txs) == 0 {
		sb.WriteString(MutedStyle.Render("  No transactions found."))
		sb.WriteString("\n")
		return sb.String()
	}

	// Header row
	sb.WriteString(TableHeaderStyle.Render(
		fmt.Sprintf("  %-4s %-8s %-12s %-15s %-25s %-12s", "ID", "Type", "Amount", "Category", "Description", "Date"),
	))
	sb.WriteString("\n")

	displayCount := len(txs)
	if displayCount > 20 {
		displayCount = 20
	}

	for i := 0; i < displayCount; i++ {
		tx := txs[i]
		style := TableRowStyle
		if i == m.cursor {
			style = style.Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
		}

		symbol := "💰"
		if tx.IsExpense() {
			symbol = "💸"
		}

		row := fmt.Sprintf("  %-4d %s%-6s %-12s %-15s %-25s %-12s",
			tx.ID, symbol, tx.Type, tx.Amount, tx.Category, tx.Description, tx.Date)
		sb.WriteString(style.Render(row))
		sb.WriteString("\n")
	}

	if len(txs) > 20 {
		sb.WriteString(MutedStyle.Render(fmt.Sprintf("\n  ... and %d more transactions", len(txs)-20)))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderReports() string {
	var sb strings.Builder

	reportTypes := []string{"Monthly", "Category", "Yearly"}
	sb.WriteString(HeaderStyle.Render("  📈 Reports"))
	sb.WriteString("  ")
	for i, rt := range reportTypes {
		if i == m.reportType {
			sb.WriteString(SelectedMenuStyle.Render("[" + rt + "]"))
		} else {
			sb.WriteString(MutedStyle.Render(" " + rt + " "))
		}
		sb.WriteString("  ")
	}
	sb.WriteString(MutedStyle.Render("(press Tab to switch)"))
	sb.WriteString("\n\n")

	if len(m.transactions) == 0 {
		sb.WriteString(MutedStyle.Render("  No data for reports."))
		sb.WriteString("\n")
		return sb.String()
	}

	var reporter services.Reporter
	switch m.reportType {
	case 0:
		reporter = &services.MonthlyReporter{}
	case 1:
		reporter = &services.CategoryReporter{}
	case 2:
		reporter = &services.YearlyReporter{}
	}

	sb.WriteString(reporter.Generate(m.transactions))
	return sb.String()
}

// RunTUI starts the BubbleTea TUI application.
func RunTUI(store *db.Store, config *utils.Config) error {
	model := NewModel(store, config)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
