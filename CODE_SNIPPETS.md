# ExpenseVault — Code Snippets Reference

---

## 1. HTTP Server with Efficient Routing

### Server Setup with Middleware Chain
> File: `api/server.go`

```go
func StartServer(addr string) error {
    mux := http.NewServeMux()

    // JWT secret for auth middleware.
    jwtSecret := []byte("your-secret-key-change-in-production")

    // Rate limiter: 10 requests per second, burst of 20.
    limiter := NewRateLimiter(10, 20, time.Second)

    // /health — public (logging only)
    mux.Handle("/health", LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })))

    // /sync — protected (logging + rate limit + JWT auth)
    syncHandler := JWTAuthMiddleware(jwtSecret)(http.HandlerFunc(handleSync))
    syncHandler = RateLimitMiddleware(limiter)(syncHandler)
    mux.Handle("/sync", LoggingMiddleware(syncHandler))

    Logger.Info("Server starting", "addr", addr)
    return http.ListenAndServe(addr, mux)
}
```

### Structured Logging Middleware (slog)
> File: `api/middleware.go`

```go
type responseRecorder struct {
    http.ResponseWriter
    statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
    rr.statusCode = code
    rr.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

        next.ServeHTTP(recorder, r)

        Logger.Info("HTTP Request",
            slog.String("method", r.Method),
            slog.String("path", r.URL.Path),
            slog.Int("status", recorder.statusCode),
            slog.Duration("latency", time.Since(start)),
            slog.String("remote", r.RemoteAddr),
            slog.String("user", getUserFromContext(r.Context())),
        )
    })
}
```

### Token-Bucket Rate Limiter
> File: `api/middleware.go`

```go
type RateLimiter struct {
    mu       sync.Mutex
    buckets  map[string]*tokenBucket
    rate     int
    burst    int
    interval time.Duration
}

type tokenBucket struct {
    tokens   int
    lastTime time.Time
}

func NewRateLimiter(rate, burst int, interval time.Duration) *RateLimiter {
    rl := &RateLimiter{
        buckets:  make(map[string]*tokenBucket),
        rate:     rate,
        burst:    burst,
        interval: interval,
    }

    // Background goroutine to clean up stale buckets every minute.
    go func() {
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()
        for range ticker.C {
            rl.cleanup()
        }
    }()

    return rl
}

func (rl *RateLimiter) allow(ip string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    bucket, exists := rl.buckets[ip]
    if !exists {
        rl.buckets[ip] = &tokenBucket{tokens: rl.burst - 1, lastTime: time.Now()}
        return true
    }

    // Refill tokens based on elapsed time.
    elapsed := time.Since(bucket.lastTime)
    refill := int(elapsed / rl.interval) * rl.rate
    bucket.tokens += refill
    if bucket.tokens > rl.burst {
        bucket.tokens = rl.burst
    }
    bucket.lastTime = time.Now()

    if bucket.tokens <= 0 {
        return false
    }
    bucket.tokens--
    return true
}

func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := r.RemoteAddr
            if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
                ip = strings.Split(fwd, ",")[0]
            }

            if !limiter.allow(ip) {
                http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### JWT Authentication Middleware
> File: `api/middleware.go`

```go
type contextKey string

const userContextKey contextKey = "username"

func JWTAuthMiddleware(secret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, `{"error":"authorization header required"}`, http.StatusUnauthorized)
                return
            }

            parts := strings.SplitN(authHeader, " ", 2)
            if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
                http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
                return
            }

            token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, jwt.ErrSignatureInvalid
                }
                return secret, nil
            })

            if err != nil || !token.Valid {
                http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
                return
            }

            claims, ok := token.Claims.(jwt.MapClaims)
            if !ok {
                http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
                return
            }

            username, _ := claims["sub"].(string)
            ctx := context.WithValue(r.Context(), userContextKey, username)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

---

## 2. Security using Bcrypt

### User Signup with Bcrypt Hashing
> File: `cmd/signup.go`

```go
var signupCmd = &cobra.Command{
    Use:   "signup",
    Short: "Create a new user account",
    RunE: func(cmd *cobra.Command, args []string) error {
        username, _ := cmd.Flags().GetString("username")
        if username == "" {
            return fmt.Errorf("username is required")
        }

        fmt.Print("Password: ")
        passBytes, err := term.ReadPassword(int(syscall.Stdin))
        fmt.Println("")
        if err != nil {
            return err
        }

        hash, err := bcrypt.GenerateFromPassword(passBytes, bcrypt.DefaultCost)
        if err != nil {
            return err
        }

        _, err = store.CreateUser(username, string(hash))
        if err != nil {
            return err
        }

        fmt.Println("Signup successful.")
        return nil
    },
}
```

### TUI Login with Bcrypt Verification
> File: `tui/app.go`

```go
func (m Model) submitLogin() (tea.Model, tea.Cmd) {
    username := m.authUserInput.Value()
    password := m.authPassInput.Value()

    user, err := m.store.GetUserByUsername(username)
    if err != nil {
        m.authMessage = "User not found."
        return m, nil
    }

    // Bcrypt comparison — constant-time check prevents timing attacks
    if err := bcrypt.CompareHashAndPassword(
        []byte(user.PasswordHash), []byte(password),
    ); err != nil {
        m.authMessage = "Incorrect password."
        return m, nil
    }

    return m, func() tea.Msg {
        return userLoggedInMsg{user: user}
    }
}
```

---

## 3. Concurrency — Optimizing Asynchronous Workflows

### Worker Pool (Fan-out/Fan-in Pattern)
> File: `services/workerpool.go`

```go
type Job struct {
    ID      int
    Name    string
    Execute func() (interface{}, error)
}

type Result struct {
    JobID  int
    Name   string
    Output interface{}
    Err    error
}

type WorkerPool struct {
    workerCount int
    jobs        chan Job
    results     chan Result
    ctx         context.Context
    cancel      context.CancelFunc
    wg          sync.WaitGroup
}

func NewWorkerPool(workerCount, jobBufferSize int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    return &WorkerPool{
        workerCount: workerCount,
        jobs:        make(chan Job, jobBufferSize),
        results:     make(chan Result, jobBufferSize),
        ctx:         ctx,
        cancel:      cancel,
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workerCount; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
    go func() {
        wp.wg.Wait()
        close(wp.results)
    }()
}

func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()
    for {
        select {
        case <-wp.ctx.Done():
            return
        case job, ok := <-wp.jobs:
            if !ok {
                return
            }
            output, err := job.Execute()
            wp.results <- Result{
                JobID: job.ID, Name: job.Name,
                Output: output, Err: err,
            }
        }
    }
}

// ProcessBatch — full fan-out/fan-in in one call
func ProcessBatch(workerCount int, jobs []Job) []Result {
    pool := NewWorkerPool(workerCount, len(jobs))
    pool.Start()

    // Fan-out: submit all jobs
    go func() {
        for _, job := range jobs {
            pool.Submit(job)
        }
        pool.Close()
    }()

    // Fan-in: collect all results
    var results []Result
    for result := range pool.Results() {
        results = append(results, result)
    }
    return results
}
```

### Recurring Transaction Scheduler (Background Goroutine)
> File: `services/scheduler.go`

```go
type Scheduler struct {
    rules    []RecurringRule
    store    TransactionAdder
    stopChan chan struct{}
}

func (s *Scheduler) StartBackground() {
    go func() {
        // Process on startup.
        created := s.ProcessDueTransactions()
        if created > 0 {
            schedLogger.Info("Startup: created recurring transactions",
                slog.Int("count", created))
        }

        ticker := time.NewTicker(1 * time.Hour)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                created := s.ProcessDueTransactions()
                if created > 0 {
                    schedLogger.Info("Periodic: created recurring transactions",
                        slog.Int("count", created))
                }
            case <-s.stopChan:
                schedLogger.Info("Scheduler stopped")
                return
            }
        }
    }()
}

func (s *Scheduler) ProcessDueTransactions() int {
    now := time.Now()
    created := 0

    for i := range s.rules {
        rule := &s.rules[i]
        if !rule.Active {
            continue
        }

        for rule.NextDue.Before(now) || rule.NextDue.Equal(now) {
            tx := models.NewTransaction(
                rule.UserID, rule.Type, rule.Amount.ToFloat64(),
                rule.Category, rule.Description,
                rule.NextDue.Format("2006-01-02"),
            )
            id, err := s.store.AddTransaction(tx)
            if err != nil {
                break
            }
            created++

            // Advance to the next due date.
            switch rule.Frequency {
            case FrequencyDaily:
                rule.NextDue = rule.NextDue.AddDate(0, 0, 1)
            case FrequencyWeekly:
                rule.NextDue = rule.NextDue.AddDate(0, 0, 7)
            case FrequencyMonthly:
                rule.NextDue = rule.NextDue.AddDate(0, 1, 0)
            }
        }
    }
    return created
}

func (s *Scheduler) Stop() {
    close(s.stopChan)
}
```

---

## 4. UI with Backend (TUI — Bubble Tea)

### Pane-Based Layout — Sidebar + Main Content
> File: `tui/app.go`

```go
func (m Model) View() string {
    // ...

    // Pane-Based Layout for Post-Login Views
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

    return sb.String()
}
```

### Sidebar Navigation Pane
> File: `tui/app.go`

```go
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

    if m.currentUser != nil && len(m.transactions) > 0 {
        sb.WriteString(MutedStyle.Render(fmt.Sprintf(" 📝 %d txns", len(m.transactions))))
        sb.WriteString("\n")
    }
    return SidebarStyle.Render(sb.String())
}
```

### Status Bar with Search / Query Modes
> File: `tui/app.go`

```go
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
```

### Lipgloss Style Definitions
> File: `tui/styles.go`

```go
var (
    SidebarStyle = lipgloss.NewStyle().
        Width(22).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62")).
        Padding(1, 1).MarginRight(1)

    MainPaneStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62")).
        Padding(1, 2)

    StatusBarStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("229")).
        Background(lipgloss.Color("236")).
        Padding(0, 1).Bold(true)

    SidebarActiveStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("229")).
        Background(lipgloss.Color("62")).
        Padding(0, 1).Width(18)

    SidebarInactiveStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("252")).
        Padding(0, 1).Width(18)

    AnomalyBoxStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("202")).
        Padding(0, 1).Foreground(lipgloss.Color("202"))

    PredictionBoxStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("177")).
        Padding(0, 1).Foreground(lipgloss.Color("177"))
)
```

---

## 5. Anomaly Detection (Z-Score Statistical Analysis)

> File: `services/anomaly.go`

```go
type AnomalyAlert struct {
    Transaction models.Transaction
    ZScore      float64
    Message     string
}

func DetectAnomalies(txs []models.Transaction) []AnomalyAlert {
    categoryTxs := make(map[models.Category][]models.Transaction)
    for _, tx := range txs {
        if tx.IsExpense() {
            categoryTxs[tx.Category] = append(categoryTxs[tx.Category], tx)
        }
    }

    var alerts []AnomalyAlert

    for cat, catTxs := range categoryTxs {
        if len(catTxs) < 3 {
            continue
        }

        // Calculate mean.
        var sum float64
        for _, tx := range catTxs {
            sum += tx.Amount.ToFloat64()
        }
        mean := sum / float64(len(catTxs))

        // Calculate standard deviation.
        var varianceSum float64
        for _, tx := range catTxs {
            diff := tx.Amount.ToFloat64() - mean
            varianceSum += diff * diff
        }
        stdDev := math.Sqrt(varianceSum / float64(len(catTxs)))

        if stdDev == 0 {
            continue
        }

        // Flag transactions > 2 standard deviations above mean.
        threshold := 2.0
        for _, tx := range catTxs {
            zScore := (tx.Amount.ToFloat64() - mean) / stdDev
            if zScore > threshold {
                alerts = append(alerts, AnomalyAlert{
                    Transaction: tx,
                    ZScore:      math.Round(zScore*100) / 100,
                    Message: fmt.Sprintf(
                        "⚠️ Unusual %s expense: ₹%.0f (%.1fx above average ₹%.0f)",
                        cat, tx.Amount.ToFloat64(), zScore, mean,
                    ),
                })
            }
        }
    }

    sort.Slice(alerts, func(i, j int) bool {
        return alerts[i].ZScore > alerts[j].ZScore
    })
    return alerts
}
```

---

## 6. Predictive Budgeting (End-of-Month Forecast)

> File: `services/anomaly.go`

```go
type PredictionData struct {
    DailyBurnRate     float64
    DaysElapsed       int
    DaysRemaining     int
    ProjectedExpenses float64
    ProjectedSavings  float64
    Confidence        string // "High", "Medium", "Low"
}

func PredictEndOfMonth(txs []models.Transaction, monthlyIncome float64) PredictionData {
    now := time.Now()
    daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
    daysElapsed := now.Day()
    daysRemaining := daysInMonth - daysElapsed

    currentMonth := now.Format("2006-01")
    var monthExpenses float64
    var txCount int
    for _, tx := range txs {
        if tx.IsExpense() && len(tx.Date) >= 7 && tx.Date[:7] == currentMonth {
            monthExpenses += tx.Amount.ToFloat64()
            txCount++
        }
    }

    pred := PredictionData{
        DaysElapsed:   daysElapsed,
        DaysRemaining: daysRemaining,
    }

    if daysElapsed == 0 {
        pred.Confidence = "Low"
        pred.ProjectedSavings = monthlyIncome
        return pred
    }

    pred.DailyBurnRate = monthExpenses / float64(daysElapsed)
    pred.ProjectedExpenses = pred.DailyBurnRate * float64(daysInMonth)
    pred.ProjectedSavings = monthlyIncome - pred.ProjectedExpenses

    switch {
    case txCount >= 15 && daysElapsed >= 15:
        pred.Confidence = "High"
    case txCount >= 7 && daysElapsed >= 7:
        pred.Confidence = "Medium"
    default:
        pred.Confidence = "Low"
    }

    return pred
}

func FormatPrediction(pred PredictionData) string {
    if pred.Confidence == "Low" {
        return "🔮 Not enough data to predict — keep logging!"
    }
    icon := "🟢"
    if pred.ProjectedSavings < 0 {
        icon = "🔴"
    } else if pred.ProjectedSavings < pred.ProjectedExpenses*0.1 {
        icon = "🟡"
    }
    return fmt.Sprintf(
        "%s Predicted EOM Savings: ₹%.0f | Burn Rate: ₹%.0f/day | %d days left [%s confidence]",
        icon, pred.ProjectedSavings, pred.DailyBurnRate, pred.DaysRemaining, pred.Confidence,
    )
}
```

---

## 7. Custom Query Language (CQL) & Fuzzy Search

> File: `services/queryparser.go`

```go
type QueryFilter struct {
    Category  string
    AmountOp  string  // ">", "<", "="
    AmountVal float64
    DateRange string  // "last-week", "last-month", "YYYY-MM-DD"
    FuzzyText string  // free text for fuzzy matching
}

func ParseQuery(query string) QueryFilter {
    var filter QueryFilter
    tokens := strings.Fields(query)
    var freeText []string

    for _, token := range tokens {
        lower := strings.ToLower(token)

        switch {
        case strings.HasPrefix(lower, "cat:"):
            filter.Category = strings.TrimPrefix(lower, "cat:")
        case strings.HasPrefix(lower, "amt:"):
            amtStr := strings.TrimPrefix(lower, "amt:")
            if strings.HasPrefix(amtStr, ">") {
                filter.AmountOp = ">"
                fmt.Sscanf(amtStr[1:], "%f", &filter.AmountVal)
            } else if strings.HasPrefix(amtStr, "<") {
                filter.AmountOp = "<"
                fmt.Sscanf(amtStr[1:], "%f", &filter.AmountVal)
            } else {
                filter.AmountOp = "="
                fmt.Sscanf(amtStr, "%f", &filter.AmountVal)
            }
        case strings.HasPrefix(lower, "date:"):
            filter.DateRange = strings.TrimPrefix(lower, "date:")
        default:
            freeText = append(freeText, token)
        }
    }

    filter.FuzzyText = strings.Join(freeText, " ")
    return filter
}

func ApplyQueryFilter(txs []models.Transaction, filter QueryFilter) []models.Transaction {
    var results []models.Transaction
    for _, tx := range txs {
        if !matchesFilter(tx, filter) {
            continue
        }
        results = append(results, tx)
    }
    return results
}
```

---

## 8. Encrypted Backup / Restore (AES-256-GCM)

> File: `services/crypto.go`

```go
const (
    saltSize = 32    // 256-bit salt
    keySize  = 32    // AES-256
    scryptN  = 32768
    scryptR  = 8
    scryptP  = 1
)

func Encrypt(plaintext []byte, password string) ([]byte, error) {
    // Generate random salt for key derivation.
    salt := make([]byte, saltSize)
    if _, err := io.ReadFull(rand.Reader, salt); err != nil {
        return nil, err
    }

    // Derive key from password using scrypt.
    key, err := scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, keySize)
    if err != nil {
        return nil, err
    }

    // Create AES cipher.
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    // Create GCM for authenticated encryption.
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    // Generate random nonce.
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    // Encrypt and authenticate.
    ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

    // Combine: salt + nonce + ciphertext
    result := make([]byte, 0, saltSize+gcm.NonceSize()+len(ciphertext))
    result = append(result, salt...)
    result = append(result, nonce...)
    result = append(result, ciphertext...)

    return result, nil
}

func Decrypt(data []byte, password string) ([]byte, error) {
    if len(data) < saltSize+12 {
        return nil, errors.New("ciphertext too short")
    }

    salt := data[:saltSize]
    rest := data[saltSize:]

    key, err := scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, keySize)
    if err != nil {
        return nil, err
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    nonce := rest[:nonceSize]
    ciphertext := rest[nonceSize:]

    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, errors.New("decryption failed: wrong password or corrupted data")
    }

    return plaintext, nil
}
```

---

## 9. AI-Powered Natural Language Queries (Gemini API)

> File: `services/llm.go`

```go
func GenerateSQL(query string, userID int64) (string, error) {
    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey == "" {
        return "", fmt.Errorf("GEMINI_API_KEY environment variable is not set")
    }

    prompt := fmt.Sprintf("%s\n\nUSER QUERY: %s\n\nNote: The user's user_id is %d.",
        dbSchemaContext, query, userID)

    return callGeminiAPI(prompt, apiKey)
}

func SummarizeData(query string, data []map[string]interface{}) (string, error) {
    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey == "" {
        return "", fmt.Errorf("GEMINI_API_KEY environment variable is not set")
    }

    dataJSON, _ := json.MarshalIndent(data, "", "  ")

    prompt := fmt.Sprintf(`
You are an AI assistant for a CLI expense tracker app.
USER'S ORIGINAL QUESTION: "%s"
DATABASE RESULTS (JSON format):
%s
INSTRUCTIONS:
1. Based only on the database results, answer the user's question.
2. Provide a concise, friendly, conversational response.
3. Format currency amounts cleanly.
`, query, string(dataJSON))

    return callGeminiAPI(prompt, apiKey)
}
```

---

## 10. Spending Trend Analysis

> File: `services/anomaly.go`

```go
func GetSpendingTrend(txs []models.Transaction) string {
    expenses := make([]models.Transaction, 0)
    for _, tx := range txs {
        if tx.IsExpense() {
            expenses = append(expenses, tx)
        }
    }

    if len(expenses) < 6 {
        return "📊 Not enough data for trend analysis yet."
    }

    sort.Slice(expenses, func(i, j int) bool {
        return expenses[i].Date < expenses[j].Date
    })

    mid := len(expenses) / 2
    olderHalf := expenses[:mid]
    recentHalf := expenses[mid:]

    var olderSum, recentSum float64
    for _, tx := range olderHalf {
        olderSum += tx.Amount.ToFloat64()
    }
    for _, tx := range recentHalf {
        recentSum += tx.Amount.ToFloat64()
    }

    olderAvg := olderSum / float64(len(olderHalf))
    recentAvg := recentSum / float64(len(recentHalf))

    changePercent := ((recentAvg - olderAvg) / olderAvg) * 100

    if changePercent > 15 {
        return fmt.Sprintf("📈 Spending is UP %.0f%% — watch your expenses!", changePercent)
    } else if changePercent < -15 {
        return fmt.Sprintf("📉 Spending is DOWN %.0f%% — great discipline!", math.Abs(changePercent))
    }
    return "📊 Spending is stable. Keep it up!"
}
```
