# ExpenseVault — Architecture Diagrams

## 1. System Architecture

```mermaid
flowchart TB
    subgraph UI["User Interface Layer"]
        CLI["CLI Commands - Cobra Framework"]
        TUI["TUI Dashboard - Bubble Tea + Lipgloss"]
        API["HTTP API Server - net/http"]
    end

    subgraph MW["Middleware Chain"]
        MW1["Structured Logger - slog JSON"]
        MW2["Rate Limiter - Token Bucket"]
        MW3["JWT Auth - Bearer Validation"]
    end

    subgraph SVC["Service Layer"]
        AUTH["Auth Service - bcrypt"]
        INSIGHTS["Insights Engine"]
        ANOMALY["Anomaly Detection - Z-Score"]
        PREDICT["Predictive Engine - EOM Forecast"]
        REPORTS["Report Generator"]
        LLM["LLM Service - Gemini API"]
        SCHED["Scheduler - Recurring Txns"]
        POOL["Worker Pool - Fan-out/Fan-in"]
        CQL["Query Parser - CQL + Fuzzy"]
        CRYPTO["Crypto Service - AES-256-GCM"]
    end

    subgraph DATA["Data Layer"]
        STORE["Unified Store - db.Store"]
        EXPORT["Export Engine - CSV / JSON"]
    end

    subgraph DBS["Database Backends"]
        SQLITE[("SQLite")]
        MYSQL[("MySQL")]
        SUPA[("Supabase PostgreSQL")]
    end

    subgraph EXT["External Services"]
        GEMINI["Google Gemini API"]
    end

    CLI --> AUTH
    CLI --> STORE
    CLI --> REPORTS
    CLI --> EXPORT
    CLI --> CRYPTO
    CLI --> LLM

    TUI --> AUTH
    TUI --> STORE
    TUI --> INSIGHTS
    TUI --> ANOMALY
    TUI --> PREDICT
    TUI --> REPORTS
    TUI --> CQL
    TUI --> LLM

    API --> MW1 --> MW2 --> MW3
    MW3 --> STORE

    SCHED --> STORE
    POOL --> REPORTS

    LLM --> GEMINI

    STORE --> SQLITE
    STORE --> MYSQL
    STORE --> SUPA

    EXPORT --> STORE
    CRYPTO --> EXPORT

    style CLI fill:#4a9eff,color:#fff
    style TUI fill:#9b59b6,color:#fff
    style API fill:#2ecc71,color:#fff
    style SQLITE fill:#3498db,color:#fff
    style MYSQL fill:#e67e22,color:#fff
    style SUPA fill:#27ae60,color:#fff
    style GEMINI fill:#ea4335,color:#fff
    style MW1 fill:#1abc9c,color:#fff
    style MW2 fill:#f39c12,color:#fff
    style MW3 fill:#e74c3c,color:#fff
```

---

## 2. Request Flow — HTTP Middleware Chain

```mermaid
sequenceDiagram
    participant C as Client
    participant L as LoggingMiddleware
    participant R as RateLimiter
    participant J as JWTAuth
    participant H as Handler
    participant DB as Database

    C->>L: POST /sync
    L->>L: Record start time
    L->>R: Forward request
    R->>R: Check token bucket for IP
    alt Rate limit exceeded
        R-->>C: 429 Too Many Requests
    else Allowed
        R->>J: Forward request
        J->>J: Parse Bearer token
        alt Invalid or Missing JWT
            J-->>C: 401 Unauthorized
        else Valid JWT
            J->>J: Inject user into context
            J->>H: Forward request
            H->>DB: Execute query
            DB-->>H: Results
            H-->>C: 200 OK + JSON
        end
    end
    L->>L: Log method, path, status, latency
```

---

## 3. Worker Pool — Fan-out/Fan-in Pattern

```mermaid
flowchart LR
    subgraph Producer
        SUBMIT["Submit Jobs"]
    end

    subgraph JobChannel
        JC["chan Job"]
    end

    subgraph Workers["Workers - Fan-out"]
        W1["Worker 0"]
        W2["Worker 1"]
        W3["Worker 2"]
        W4["Worker N"]
    end

    subgraph ResultChannel
        RC["chan Result"]
    end

    subgraph Consumer["Consumer - Fan-in"]
        COLLECT["Collect Results"]
    end

    subgraph Control
        CTX["context.Context - Graceful Shutdown"]
    end

    SUBMIT --> JC
    JC --> W1
    JC --> W2
    JC --> W3
    JC --> W4
    W1 --> RC
    W2 --> RC
    W3 --> RC
    W4 --> RC
    RC --> COLLECT
    CTX -.->|cancel| W1
    CTX -.->|cancel| W2
    CTX -.->|cancel| W3
    CTX -.->|cancel| W4

    style SUBMIT fill:#3498db,color:#fff
    style COLLECT fill:#2ecc71,color:#fff
    style CTX fill:#e74c3c,color:#fff
    style W1 fill:#9b59b6,color:#fff
    style W2 fill:#9b59b6,color:#fff
    style W3 fill:#9b59b6,color:#fff
    style W4 fill:#9b59b6,color:#fff
```

---

## 4. Anomaly Detection and Prediction Pipeline

```mermaid
flowchart TD
    TX["All Transactions"] --> GROUP["Group by Category"]
    GROUP --> STATS["Calculate Mean and Std Dev per Category"]
    STATS --> ZSCORE["Compute Z-Score for Each Transaction"]
    ZSCORE --> FLAG{"Z-Score > 2 sigma?"}
    FLAG -->|Yes| ALERT["AnomalyAlert with Message and Severity"]
    FLAG -->|No| SAFE["Normal"]
    ALERT --> SORT["Sort by Severity - Highest First"]
    SORT --> DASH["Dashboard Display - Top 3 Alerts"]

    TX --> TREND["Split Older vs Recent - Compare Averages"]
    TREND --> TDIR{"Change exceeds 15%?"}
    TDIR -->|Up| UP["Spending UP"]
    TDIR -->|Down| DOWN["Spending DOWN"]
    TDIR -->|Stable| STABLE["Stable"]

    TX --> BURN["Daily Burn Rate = Month Expenses / Days Elapsed"]
    BURN --> PROJECT["Projected Expenses = Burn Rate x Days in Month"]
    PROJECT --> SAVINGS["Projected Savings = Income minus Projected"]
    SAVINGS --> CONF{"Data Points?"}
    CONF -->|"15+ txns, 15+ days"| HIGH["High Confidence"]
    CONF -->|"7+ txns, 7+ days"| MED["Medium Confidence"]
    CONF -->|"Less than 7"| LOW["Low Confidence"]

    style ALERT fill:#e74c3c,color:#fff
    style HIGH fill:#27ae60,color:#fff
    style MED fill:#f39c12,color:#fff
    style LOW fill:#e74c3c,color:#fff
    style DASH fill:#9b59b6,color:#fff
```

---

## 5. Encrypted Backup and Restore Flow

```mermaid
flowchart LR
    subgraph Backup["Backup with --encrypt"]
        B1["Fetch All Transactions"] --> B2["Serialize to JSON"]
        B2 --> B3["Prompt Password"]
        B3 --> B4["Generate Random Salt"]
        B4 --> B5["Derive Key via scrypt"]
        B5 --> B6["AES-256-GCM Encrypt"]
        B6 --> B7["Save salt + nonce + ciphertext"]
    end

    subgraph Restore["Restore with --decrypt"]
        R1["Read Encrypted File"] --> R2["Prompt Password"]
        R2 --> R3["Extract Salt from Header"]
        R3 --> R4["Derive Key via scrypt"]
        R4 --> R5["AES-256-GCM Decrypt"]
        R5 --> R6["Parse JSON"]
        R6 --> R7["BulkInsert to Database"]
    end

    style B6 fill:#e74c3c,color:#fff
    style R5 fill:#27ae60,color:#fff
    style B7 fill:#3498db,color:#fff
    style R7 fill:#3498db,color:#fff
```

---

## 6. TUI Pane Layout Architecture

```mermaid
flowchart TB
    subgraph Layout["View Function - Master Layout"]
        TITLE["Title Bar - ExpenseVault + DB Info + Username"]

        subgraph HorizontalJoin["Horizontal Join"]
            SIDEBAR["Sidebar - Navigation Pane - 22px fixed width"]
            MAIN["Main Content Pane - Dynamic width"]
        end

        STATUS["Status Bar - Hotkeys or Search or Query Input"]
    end

    TITLE --> SIDEBAR
    TITLE --> MAIN
    SIDEBAR --> STATUS
    MAIN --> STATUS

    MAIN --> D["ViewDashboard - KPIs + Bars + Anomalies + Prediction"]
    MAIN --> T["ViewTransactions - Table + Filtered Results"]
    MAIN --> A["ViewAddForm - 6-Field Input"]
    MAIN --> R["ViewReports - Monthly or Category or Yearly"]
    MAIN --> AI["ViewAsk - AI Query + Response"]

    style SIDEBAR fill:#2c3e50,color:#fff
    style MAIN fill:#34495e,color:#fff
    style STATUS fill:#f39c12,color:#000
    style TITLE fill:#8e44ad,color:#fff
```

---

## 7. CQL Query Pipeline

```mermaid
flowchart LR
    INPUT["User Input: cat:food amt:>500 date:last-week"] --> TOKENIZE["Tokenize via strings.Fields"]
    TOKENIZE --> PARSE["Parse Tokens"]
    PARSE --> CAT["cat: Category Filter"]
    PARSE --> AMT["amt:> Amount Filter"]
    PARSE --> DATE["date: Date Range Filter"]
    PARSE --> FREE["Free Text Fuzzy Match"]
    CAT --> QF["QueryFilter Struct"]
    AMT --> QF
    DATE --> QF
    FREE --> QF
    QF --> APPLY["ApplyQueryFilter - Filter Transactions"]
    APPLY --> RESULT["Filtered Results - Live Updated in TUI"]

    style INPUT fill:#3498db,color:#fff
    style QF fill:#9b59b6,color:#fff
    style RESULT fill:#27ae60,color:#fff
```
