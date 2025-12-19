
# URL Shortener Service - Finan Backend Challenge

A robust and feature-rich URL shortening service built with Golang, featuring custom aliases, real-time expiration tracking, and integrated QR code management.


## 🚀 Key Features

Custom Aliases: Users can define personalized short codes (e.g., /summer-sale) instead of random strings.

Real-time Expiration Timer:

 - Set expiration dates for shortened links to manage link lifecycle.

 - Dynamic countdown timer on the dashboard updates every second     without page refreshes.

QR Code Integration:

- Automatically generates a unique QR code for every shortened link.

- Download Support: Save QR codes as .png files directly to your device.

Copy to Clipboard: One-click action to copy QR images for quick sharing.

Input Validation: Strict frontend and backend checks to ensure valid URL formats (http/https) and prevent empty submissions.

Click Analytics: Real-time tracking of total visits per link.


## 🛠 Tech Stack

**Backend:** Golang with Gin Gonic (High-performance HTTP web framework).

**Database:** SQLite via GORM (Portable, file-based database requiring zero setup).

**Frontend:** HTML5, Vanilla JavaScript, and Bootstrap 5 for a responsive UI.

**Third-party API:** QR Server API for dynamic QR code generation.


## 🏗 Project Structure

```
├── handlers/     # API logic (Shortening, Redirecting, Fetching)
├── models/       # Data schemas (GORM Link model)
├── store/        # Database connection & migration setup
├── static/       # Frontend assets (index.html)
├── main.go       # Application entry point
└── gorm.db       # SQLite database file (Auto-generated)
```


## 📝 Technical Decisions

1. Zero-Configuration DB: SQLite was chosen to ensure the project is "Plug and Play" for recruiters. No Docker or external SQL server is required to review the code.

2. State Management: JavaScript's window.onload and setInterval ensure that the data is synchronized and the UI remains interactive (Live timers) from the moment the page loads.

3. Scalability Consideration: The backend is structured with separate handlers and models, making it easy to swap SQLite for PostgreSQL or MySQL as traffic grows.
## 🏃 How to Run

1. Ensure you have Go 1.18+ installed.

2. Clone this repository.

3. Run the application:
```bash
  go run main.go
```
4. Open your browser and navigate to: http://localhost:8080

```This project is part of a technical assessment. Future updates may include User Authentication and Advanced Analytics.```

