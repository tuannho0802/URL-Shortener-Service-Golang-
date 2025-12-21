
# URL Shortener Service - Finan Backend Challenge

A robust and feature-rich URL shortening service built with Golang, featuring custom aliases, real-time expiration tracking, and integrated QR code management.


## 🚀 Key Features

- **Core Shortening:** Converts long URLs into manageable short codes with automatic redirection.


- **Custom Aliases:** Users can define personalized short codes (e.g., /finan-test) instead of random strings.


- **Real-Time Dashboard:** Uses WebSockets to update the link list and click counts instantly across all connected clients without page refreshes.


- **Expiration Management:** Supports Minute/Hour/Day expiration presets with a dynamic JavaScript countdown timer.


- **QR Code Integration:** Automatically generates unique QR codes via API, with support for instant preview and .png downloads.


- **Advanced UI:** Features full pagination, sorting (by clicks or creation date), and responsive design using Bootstrap 5.

- **Click Analytics:** Real-time tracking of total visits per link.


## 🛠 Tech Stack

- **Backend:** Golang (Gin Gonic) - Chosen for its superior execution speed and native support for high concurrency.


- **Database:** SQLite (GORM) - Selected for its "Zero-Config" nature, making the project portable and easy for reviewers to run without external dependencies.

- **Real-time Communication:** Gorilla WebSocket - Ensures live data synchronization between the server and the dashboard.

- **Testing: k6:** Used for load testing to ensure the service remains stable under stress.

- **Third-party API:** QR Server API for dynamic QR code generation.


## 🏗 Project Structure

```
├── handlers/    # API & WebSocket logic (Shorten, Redirect, Real-time Hub)
├── models/      # GORM database schemas
├── store/       # Database connection & auto-migration setup
├── static/      # Frontend assets (HTML, Vanilla JS, CSS)
└── main.go      # Application entry point
└── gorm.db      # SQLite database file (Auto-generated)
```


## 📝 Technical Decisions & Trade-offs

### 1. Generation Algorithm & Conflict Handling
- **Method:** The service generates a random short code and verifies its uniqueness in the database.
- **Decision:** I chose a random string over Base62 ID encoding to prevent competitors from "scraping" links sequentially.
- **Trade-off:** This requires a database check before insertion, but it's mitigated by indexing the short_code column for $O(1)$ lookup performance.

### 2. Concurrency & Performance
- **Stress Test Results:** The system was tested using k6, handling ~450 requests/sec with 150 Virtual Users.
- **WebSocket Throttling:** To prevent UI crashes during high-traffic bursts, I implemented a 2-second notification batching mechanism in the NotifyDataChange function.

### 3. Database Choice: Why SQLite?
- **Why:** It is a file-based SQL database, making the project "Plug and Play".
- **Optimization:** Configured PRAGMA journal_mode=WAL to allow multiple concurrent readers and a single writer without blocking the system.

### 🔒 Security & Limitations
- **Validation:** Strict URL validation on both frontend and backend to prevent XSS or invalid protocol submissions.


- **Current Limitation:** The service is currently "No-User," meaning it is susceptible to storage spam (Improve later).


- **Future Scalability:** To handle 100x traffic, the architecture allows for an easy swap to PostgreSQL and the addition of Redis for caching hot URLs.
## 🏃 How to Run

1. Ensure you have Go 1.18+ installed.

2. Clone this repository.

3. Run the application:
```bash
go mod tidy
go run main.go
  
```
4. Open your browser and navigate to: *http://localhost:8080*

```This project is part of a technical assessment. Future updates may include User Authentication and Advanced Analytics.```


## 💡 Challenges and Learnings
During the development of this project, I encountered several technical hurdles that required research and iterative problem-solving:

### 1. Real-time Synchronization vs. Performance

- **Challenge:** Initially, the dashboard required manual refreshes to see new links or updated click counts.
- **Solution:** I implemented WebSockets to push updates to all connected clients instantly.
- **Optimization:** During stress testing, I realized that frequent updates (400+ per second) could overwhelm the browser's main thread. I solved this by implementing a Throttling/Batching mechanism in the Go backend, ensuring notifications are sent at most once every 2 seconds.

### 2. High Concurrency and Database Locking
- **Challenge:** Using SQLite under high-stress loads (150+ concurrent users) can lead to database is locked errors because SQLite normally allows only one writer at a time.
- **Solution:** I enabled WAL (Write-Ahead Logging) mode and set a busy_timeout in the GORM connection string. This allowed concurrent read/write operations and improved stability during the k6 load tests.

### 3. Frontend State & Pagination Logic
- **Challenge:** Maintaining a consistent UI state while pages were being updated via WebSockets was difficult. If a user was on "Page 2" and a new link was added, they would often get pushed back to "Page 1." 
- **Solution:** I refactored the frontend logic to pass the currentPage variable into the loadLinks() function during WebSocket updates. This ensures the user stays on their current view while the table content refreshes dynamically.


## 🎓 What I Learned

**Go Concurrency:** I gained hands-on experience with Goroutines and Channels to manage the WebSocket Hub and background tasks.


**Load Testing:** Using k6 taught me how to identify bottlenecks in my code and how to interpret performance metrics like p(95) latency.


**System Design:** I learned the importance of Trade-offs—such as choosing SQLite for its simplicity in a test environment while acknowledging its limitations for large-scale production.