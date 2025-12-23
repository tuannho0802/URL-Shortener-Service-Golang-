# URL Shortener Service - Golang

A robust and feature-rich URL shortening service built with Golang, featuring custom aliases, real-time expiration tracking, and integrated QR code management.

## 🚀 Key Features

- **Core Shortening:** Converts long URLs into manageable short codes with automatic redirection.
- **User Authentication:** Secure user registration and login with JWT-based authentication.
- **Custom Aliases:** Users can define personalized short codes (e.g., /my-custom-link) instead of random strings.
- **Real-Time Dashboard:** Uses WebSockets to update the link list and click counts instantly across all connected clients without page refreshes.
- **Expiration Management:** Supports Minute/Hour/Day expiration presets with a dynamic JavaScript countdown timer.
- **Background Worker:** A background worker automatically cleans up expired links to keep the database clean.
- **QR Code Integration:** Automatically generates unique QR codes via API, with support for instant preview and .png downloads.
- **Advanced UI:** Features full pagination, sorting (by clicks or creation date), and responsive design using Bootstrap 5.
- **Click Analytics:** Real-time tracking of total visits per link.

## 🛠 Tech Stack

- **Backend:** Golang (Gin Gonic) - Chosen for its superior execution speed and native support for high concurrency.
- **Database:** PostgreSQL (GORM) - A powerful, open-source object-relational database system.
- **Real-time Communication:** Gorilla WebSocket - Ensures live data synchronization between the server and the dashboard.
- **Authentication:** JWT (JSON Web Tokens) - For secure user authentication.
- **Testing:** k6 - Used for load testing to ensure the service remains stable under stress.
- **Third-party API:** QR Server API for dynamic QR code generation.

## 🏗 Project Structure

```
├── handlers/    # API & WebSocket logic (Shorten, Redirect, Real-time Hub)
├── middleware/  # Authentication and other middleware
├── models/      # GORM database schemas
├── store/       # Database connection & auto-migration setup
├── static/      # Frontend assets (HTML, Vanilla JS, CSS)
├── templates/   # HTML templates
├── .env.example # Example environment variables
├── main.go      # Application entry point
├── Dockerfile   # Docker configuration
└── README.md    # Project documentation
```

## 📝 Technical Decisions & Trade-offs

### 1. Generation Algorithm & Conflict Handling
- **Method:** The service generates a random short code and verifies its uniqueness in the database.
- **Decision:** I chose a random string over Base62 ID encoding to prevent competitors from "scraping" links sequentially.
- **Trade-off:** This requires a database check before insertion, but it's mitigated by indexing the short_code column for O(1) lookup performance.

### 2. Concurrency & Performance
- **Stress Test Results:** The system was tested using k6, handling ~450 requests/sec with 150 Virtual Users.
- **WebSocket Throttling:** To prevent UI crashes during high-traffic bursts, I implemented a 2-second notification batching mechanism in the NotifyDataChange function.

### 3. Database Choice: Why PostgreSQL?
- **Why:** PostgreSQL is a powerful, open-source object-relational database system with a strong reputation for reliability, feature robustness, and performance.
- **Optimization:** Using GORM for object-relational mapping simplifies database interactions and improves developer productivity.

## 🔒 Security & Limitations

- **Validation:** Strict URL validation on both frontend and backend to prevent XSS or invalid protocol submissions.
- **Authentication:** User authentication is implemented using JWT, ensuring that only authorized users can create and manage links.
- **Future Scalability:** To handle 100x traffic, the architecture allows for an easy swap to a different database and the addition of Redis for caching hot URLs.

## 🏃 How to Run

1.  **Prerequisites:**
    *   Go 1.18+ installed.
    *   PostgreSQL running on your machine or a remote server.
    *   Docker (optional, for running the application in a container).

2.  **Clone the repository:**
    ```bash
    git clone https://github.com/tuannho0802/URL-Shortener-Service-Golang-.git
    cd URL-Shortener-Service-Golang-
    ```

3.  **Set up environment variables:**
    *   Create a `.env` file in the root of the project.
    *   Copy the contents of `.env.example` into `.env`.
    *   Update the `DATABASE_URL` with your PostgreSQL connection string.
    *   Update the `JWT_SECRET` with a secret key for signing JWTs.

4.  **Run the application:**
    ```bash
    go mod tidy
    go run main.go
    ```

5.  **Open your browser and navigate to:** `http://localhost:8080`

## 🐳 Running with Docker

1.  **Build the Docker image:**
    ```bash
    docker build -t url-shortener .
    ```

2.  **Run the Docker container:**
    ```bash
    docker run -p 8080:8080 -e DATABASE_URL="your_database_url" -e JWT_SECRET="your_jwt_secret" url-shortener
    ```

## 💡 Challenges and Learnings
During the development of this project, I encountered several technical hurdles that required research and iterative problem-solving:

### 1. Real-time Synchronization vs. Performance

- **Challenge:** Initially, the dashboard required manual refreshes to see new links or updated click counts.
- **Solution:** I implemented WebSockets to push updates to all connected clients instantly.
- **Optimization:** During stress testing, I realized that frequent updates (400+ per second) could overwhelm the browser's main thread. I solved this by implementing a Throttling/Batching mechanism in the Go backend, ensuring notifications are sent at most once every 2 seconds.

### 2. High Concurrency and Database Locking
- **Challenge:** Using a database under high-stress loads can lead to locking issues.
- **Solution:** I chose PostgreSQL for its robust concurrency control, which handles multiple simultaneous connections efficiently.

### 3. Frontend State & Pagination Logic
- **Challenge:** Maintaining a consistent UI state while pages were being updated via WebSockets was difficult. If a user was on "Page 2" and a new link was added, they would often get pushed back to "Page 1."
- **Solution:** I refactored the frontend logic to pass the `currentPage` variable into the `loadLinks()` function during WebSocket updates. This ensures the user stays on their current view while the table content refreshes dynamically.

## 🎓 What I Learned

**Go Concurrency:** I gained hands-on experience with Goroutines and Channels to manage the WebSocket Hub and background tasks.
**Database Management:** I learned how to work with PostgreSQL and GORM, including setting up connections, performing migrations, and handling transactions.
**Authentication:** I implemented JWT-based authentication to secure the application and protect user data.
**Load Testing:** Using k6 taught me how to identify bottlenecks in my code and how to interpret performance metrics like p(95) latency.
**System Design:** I learned the importance of Trade-offs—such as choosing a database that fits the project's needs while considering its scalability and performance characteristics.