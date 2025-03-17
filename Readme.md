# Uptime Monitor

## Overview

Uptime Monitor is a Golang-based project designed to track website availability and performance. It provides uptime statistics, response time metrics, and alerting mechanisms.

## Timeline

- **CRUD APIs for Uptime Monitoring**
- **Basic HTMX UI**
- **List Integration** for managing multiple URLs
- **Search List**
- **Summary Statistics**
- **Uptime Trend with Color Codes**
- **Performance Chart** displaying TTFB vs Total Response Time
- **Batching of Scheduler Operations** for efficiency
- **Service Downtime Detection** using multiple failed checks
- **Email Alerts** for downtime notifications
- **Graphs:**
  - **TTFB vs Response Time** (combined view)
  - **Average Response Time Graph**
  - **Uptime Bars** (last 30 days)
- **Date Range Filter** for graph insights
- **Pause Resume Montior processing**
- **Session based Auth Flow**
- **Performance improvement** using go concurrency
- **Dockerization**

## TODO List

- **Input Validation using zod like lib**
- **Default Values**
- **Precalculated Stats**
- **Multi-URL Stats Comparison**
- **Cloud Deployment**
- **Refactoring**
- **SSE/Websocket**

## Tech Stack

- **Backend:** Golang (net/http), PostgreSQL
- **Frontend:** HTMX, Tailwind CSS
- **Charting:** go-echarts

## Installation & Setup

1. Clone the repository:
   ```sh
   git clone https://github.com/NitinJuyal1610/uptime_monit_go
   cd uptime-monitor
   ```
2. Install dependencies:
   ```sh
   go mod tidy
   ```
3. Set up environment variables in `.env` file:
   ```env
    PORT=8022
    DB_HOST=db
    DB_PORT=5432
    DB_USER=postgres_user
    DB_PASSWORD=postgres_pass
    DB_NAME=db_postgres
    GMAIL_USER=temp@gmail.com
    GMAIL_PASS=xxxx xxxx xxxx xxxx
    SESSION_SECRET=session_secret
   ```
4. Run the application using Docker Compose:
   ```sh
   docker-compose up --build
   ```
5. Open the frontend UI in a browser.
