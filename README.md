# Distributed Search Engine

This project is a distributed search engine built with Go and React, designed to index and search through product data efficiently. The backend is architected to run in both a standalone mode and a distributed mode with multiple indexing workers. The frontend provides a clean, modern user interface for searching and filtering results.

## Architecture

The project is divided into three main parts: a Go-based backend, a React-based frontend, and the data processing scripts and files.

### Backend (Go)

The backend is responsible for indexing data and serving search requests.

-   **API Server**: A web server built with the **Gin** framework that exposes a search endpoint.
-   **Indexing**: The engine can build a search index from CSV files. It uses an inverted index to provide fast full-text search capabilities.
-   **Storage**: It uses **BoltDB**, an embedded key/value database, for storing the index locally.
-   **Distributed System**: In distributed mode, the system consists of:
    -   A **Web Server** that acts as a gateway, forwarding search requests to index workers.
    -   Multiple **gRPC Index Servers** (workers) that each hold a partition of the index and perform the actual search.
    -   **etcd** is used for service discovery.

### Frontend (React)

The frontend is a single-page application (SPA) that provides the user interface for the search engine.

-   **Framework**: Built with **React** and **Vite** for a fast development experience.
-   **UI Components**: Uses **Material-UI** for a rich set of pre-built and customizable components.
-   **Routing**: **React Router** is used for navigation between the home and search results pages.
-   **Communication**: Interacts with the backend's `/search` API endpoint to fetch results.

## Dataset

The search index is built using the **Amazon Products Dataset** available on Kaggle. This dataset contains a large number of products with details like name, category, price, and image URLs.

-   **Link**: [https://www.kaggle.com/datasets/lokeshparab/amazon-products-dataset](https://www.kaggle.com/datasets/lokeshparab/amazon-products-dataset)
-   The CSV files from this dataset are located in the `/data/archive` directory.

## Getting Started

### Prerequisites

-   [Go](https://go.dev/doc/install) (version 1.18 or later)
-   [Node.js](https://nodejs.org/en/download/) (version 16 or later) and npm

### Setup

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd <repository-directory>
    ```

2.  **Install backend dependencies:**
    ```bash
    go mod tidy
    ```

3.  **Install frontend dependencies:**
    ```bash
    cd frontend
    npm install
    cd ..
    ```

## How to Run

You can run the search engine in either standalone or distributed mode.

### 1. Backend

#### Standalone Mode

In this mode, a single server handles both web requests and searching.

1.  **Build the Index:**
    First, run the server with the `-index` flag to build the search index from the CSV files.
    ```bash
    go run ./cmd/server -mode=1 -index=true -port=5678 -dbPath=./data/local_db/standalone_bolt
    ```
    *Wait for the indexing process to complete. The server will then be ready.*

2.  **Run the Server:**
    For subsequent runs, you can start the server without the `-index` flag to load the previously built index.
    ```bash
    go run ./cmd/server -mode=1 -port=5678 -dbPath=./data/local_db/standalone_bolt
    ```

#### Distributed Mode

In this mode, one web server distributes search queries to multiple gRPC indexing workers.

1.  **Start Indexing Workers:**
    Launch two or more indexer workers. Each worker needs a unique `workerIndex` and `port`, and will manage a separate part of the index.

    *Terminal 1 (Worker 0):*
    ```bash
    go run ./cmd/server -mode=2 -index=true -port=5600 -dbPath=./data/local_db/worker0_bolt -totalWorkers=2 -workerIndex=0
    ```

    *Terminal 2 (Worker 1):*
    ```bash
    go run ./cmd/server -mode=2 -index=true -port=5601 -dbPath=./data/local_db/worker1_bolt -totalWorkers=2 -workerIndex=1
    ```
    *Wait for both workers to finish building their index partitions.*

2.  **Start the Web Server:**
    Once the workers are running, start the main web server which will act as the entry point.

    *Terminal 3 (Web Server):*
    ```bash
    go run ./cmd/server -mode=3 -port=5678
    ```

### 2. Frontend

1.  **Navigate to the frontend directory:**
    ```bash
    cd frontend
    ```

2.  **Start the development server:**
    ```bash
    npm run dev
    ```
    The frontend will be available at `http://localhost:5173`.

## API Endpoints

### Search

-   **URL**: `/search`
-   **Method**: `POST`
-   **Body (JSON)**:
    ```json
    {
      "Query": "your search query",
      "Classes": ["Optional", "Category", "Filters"]
    }
    ```
