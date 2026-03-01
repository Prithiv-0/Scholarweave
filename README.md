# ScholarWeave

<div align="center">

**A modern, full-stack academic paper search platform powered by OpenAlex API**

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react)](https://reactjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.3+-3178C6?logo=typescript)](https://www.typescriptlang.org)

[Features](#-features) • [Quick Start](#-quick-start) • [Architecture](#-architecture) • [API Documentation](#-api-documentation) • [Contributing](#-contributing)

</div>

---

## Overview

**ScholarWeave** is a comprehensive academic research platform that enables researchers, students, and academics to search, discover, and explore scholarly papers from across the world. Built with a modern tech stack, it provides a fast, intuitive, and responsive interface for academic literature discovery.

The platform integrates with the [OpenAlex](https://openalex.org) API, a free and open catalog of the world's scholarly papers, authors, institutions, and more, offering access to millions of research papers with comprehensive metadata.

### Use Cases

- **Researchers**: Quickly find relevant papers, track citations, and discover related work
- **Students**: Access academic literature for research papers and assignments
- **Academics**: Stay updated with the latest publications in their field
- **Librarians**: Help patrons discover scholarly resources efficiently

---

## Features

### Current Features

- **Full-Text Search**: Search across millions of academic papers by title, author, keywords, or topics
- **Citation Metrics**: View citation counts and paper impact at a glance
- **Author Information**: Display comprehensive author details for each paper
- **Paper Details**: Access abstracts, DOI links, and complete metadata
- **Health Monitoring**: Real-time API health status tracking
- **Responsive Design**: Fully optimized for mobile, tablet, and desktop devices
- **Fast Performance**: Built with Vite for lightning-fast development and production builds
- **Modern UI**: Clean, intuitive interface powered by Tailwind CSS
- **Direct Paper Access**: Navigate to detailed paper views with full information
- **Semantic Web Support**: JSON-LD structured data and RDF/XML export for machine readability

### Upcoming Features

- Save and organize favorite papers
- Create custom reading lists and collections
- Citation alerts and paper recommendations
- Advanced analytics and visualization
- Multi-source integration (Semantic Scholar, PubMed, arXiv)
- Export citations in multiple formats (BibTeX, RIS, EndNote)
- Dark mode support

---

## Architecture

ScholarWeave currently follows a modular **full-stack architecture** with clear separation between frontend and backend:

```
┌─────────────────────────────────────────────────────────────┐
│                         Client Layer                         │
│  (React + TypeScript + Vite + Tailwind CSS)                 │
│  - Search Interface                                          │
│  - Paper Display                                             │
│  - User Interactions                                         │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP/REST API
┌────────────────────────▼────────────────────────────────────┐
│                      API Layer (Go)                          │
│  - Fiber v3 Web Framework                                    │
│  - RESTful Endpoints                                         │
│  - Request Validation                                        │
│  - Error Handling                                            │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    Business Logic Layer                      │
│  - Search Service                                            │
│  - Data Normalization                                        │
│  - Metadata Enrichment                                       │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                  External API Integration                    │
│  - OpenAlex API Client                                       │
│  - HTTP Client with Error Handling                           │
│  - Response Parsing & Mapping                                │
└─────────────────────────────────────────────────────────────┘
```

### Tech Stack

#### Backend
- **Language**: Go 1.25+
- **Web Framework**: [Fiber v3](https://github.com/gofiber/fiber) - Express-inspired web framework
- **HTTP Client**: Native Go `net/http`
- **Middleware**: CORS, Logger, Recover (panic recovery)
- **Architecture Pattern**: MVC with clean separation of concerns

#### Frontend
- **Framework**: React 18
- **Language**: TypeScript 5.3+
- **Build Tool**: Vite 5
- **UI Library**: Tailwind CSS 3.4
- **Routing**: React Router DOM 6
- **HTTP Client**: Axios 1.6
- **Code Quality**: ESLint + Prettier

#### External APIs
- **OpenAlex API**: Primary data source for scholarly papers
- **API Documentation**: [OpenAlex Docs](https://docs.openalex.org)

---

## API Documentation

### Base URL

`http://localhost:3000/api/v1`

### Endpoints

#### `GET /health`
Returns API health information.

#### `GET /papers/search`
Searches papers from OpenAlex with pagination, sorting, and filters.

**Query parameters**
- `q` (required): search query
- `page` (optional, default `1`): page number
- `per_page` (optional, default `10`, max `50`): items per page
- `sort` (optional): `relevance` | `citations` | `date`
- `from_year` (optional): lower publication year bound
- `to_year` (optional): upper publication year bound
- `type` (optional): publication type (for example `journal-article`)
- `subject` (optional): subject keyword appended to search query

**Response shape**
```json
{
   "meta": {
      "count": 12345,
      "page": 1,
      "per_page": 10,
      "total_pages": 1235,
      "sort": "relevance"
   },
   "results": [
      {
         "id": "https://openalex.org/W123",
         "title": "Example paper",
         "abstract": "...",
         "doi": "https://doi.org/...",
         "authors": [{ "name": "Author Name" }],
         "cited_by_count": 42,
         "source": "OpenAlex"
      }
   ]
}
```

#### `GET /papers/:id`
Returns a normalized paper object by OpenAlex work ID (or encoded OpenAlex URL).

#### `GET /papers/:id/rdf`
Returns RDF/XML metadata for a paper.

#### `POST /auth/register`
Creates a user account and returns a JWT token with user profile.

#### `POST /auth/login`
Authenticates user credentials and returns a JWT token with user profile.

#### `GET /users/me`
Returns the authenticated user's profile. Requires `Authorization: Bearer <token>`.

#### `GET /users/me/favorites`
Lists the authenticated user's saved favorite papers.

#### `POST /users/me/favorites`
Adds a paper to favorites.

#### `DELETE /users/me/favorites/:paperId`
Removes a paper from favorites.

#### `GET /users/me/lists`
Lists the authenticated user's reading lists.

#### `POST /users/me/lists`
Creates a reading list.

#### `POST /users/me/lists/:listId/items`
Adds a paper to a reading list.

#### `DELETE /users/me/lists/:listId/items/:paperId`
Removes a paper from a reading list.

---

## Quick Start

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.25+** - [Download](https://golang.org/dl)
- **Node.js 16+** - [Download](https://nodejs.org)
- **npm** (comes with Node.js)
- **Git** - [Download](https://git-scm.com)

### Installation

#### 1. Clone the Repository

```bash
git clone https://github.com/Prithiv-0/Scholarweave.git
cd Scholarweave
```

#### 2. Backend Setup

```bash
# Download Go dependencies
go mod download

# Run the backend server
go run main.go
```

The backend API will start on `http://localhost:3000`

**Test the API:**
```bash
curl http://localhost:3000/api/v1/health
```

Expected response:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2024-12-24T09:17:10Z",
  "services": {
    "api": "healthy",
    "openalex": "connected"
  }
}
```

#### 3. Frontend Setup

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start the development server
npm run dev
```

The frontend will start on `http://localhost:5173`

#### 4. Access the Application

Open your browser and navigate to:
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:3000
- **API Health Check**: http://localhost:3000/api/v1/health

---

## Deployment to Render

ScholarWeave is optimized for deployment on Render using the included Web Service + PostgreSQL Database containerized workflow.

1. Fork or push this repository to your own GitHub account.
2. Log into [Render](https://render.com) and click **New > Blueprint**.
3. Connect your GitHub repository to Render.
4. Render will read the `render.yaml` blueprint and automatically create:
   - A PostgreSQL database (`scholarweave-db`)
   - A Docker web service (`scholarweave-api`)
5. The deployment process will automatically build the React frontend, compile the Go backend, run database migrations, and start the unified server.

---

## Development

### Project Structure

```
Scholarweave/
├── api/                          # API handlers
│   └── handlers/
│       └── openalex.go          # OpenAlex API handler
├── internal/                     # Internal packages
│   ├── models/
│   │   └── paper.go             # Data models
│   └── services/
│       └── search_service.go    # Business logic
├── frontend/                     # React frontend
│   ├── src/
│   │   ├── api/                 # API client
│   │   ├── components/          # React components
│   │   ├── pages/              # Page components
│   │   ├── App.tsx             # Root component
│   │   └── main.tsx            # Entry point
│   ├── package.json
│   └── vite.config.ts
├── main.go                      # Backend entry point
├── go.mod                       # Go dependencies
├── go.sum                       # Go checksums
├── .env.example                 # Environment variables template
├── SETUP.md                     # Detailed setup guide
├── LICENSE                      # GPL v3 License
└── README.md                    # This file
```

## 🤝 Contributing

We welcome contributions from the community! Here's how you can help:

### Getting Started

1. **Fork the repository**
2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/Scholarweave.git
   cd Scholarweave
   ```
3. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
4. **Make your changes**
5. **Test your changes**
   - Backend: `go test ./...`
   - Frontend: `npm test` (when tests are implemented)
6. **Commit your changes**
   ```bash
   git commit -m "Add amazing feature"
   ```
7. **Push to your fork**
   ```bash
   git push origin feature/amazing-feature
   ```
8. **Open a Pull Request**

#### Pull Request Process
1. Update the README.md with details of changes if needed
2. Update the SETUP.md if installation steps change
3. Ensure all tests pass

#### Reporting Issues
When reporting issues, please include:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Screenshots (if applicable)
- Environment details (OS, Go version, Node version)

---

## 📋 Roadmap

### Phase 1: Core Features (Current)
- [x] OpenAlex API integration
- [x] Paper search functionality
- [x] Paper detail views
- [x] Responsive UI
- [x] Health monitoring

### Phase 2: Enhanced Search
- [x] Advanced search filters (date range, publication type, subject area)
- [x] Sort by relevance, citations, date
- [x] Pagination for large result sets
- [ ] Search suggestions and autocomplete

### Phase 3: User Features
- [x] User authentication (JWT)
- [x] User profiles
- [x] Save favorite papers
- [x] Create reading lists
- [ ] Share collections

### Phase 4: Data & Analytics
- [ ] Citation network visualization
- [ ] Author collaboration graphs
- [ ] Topic trending analysis
- [ ] Paper recommendations

### Phase 5: Multi-Source Integration
- [ ] Semantic Scholar integration
- [ ] PubMed/MEDLINE integration
- [ ] arXiv integration
- [ ] CrossRef metadata

### Phase 6: Advanced Features
- [ ] PDF viewer integration
- [ ] Citation export (BibTeX, RIS, EndNote)
- [ ] Browser extension
- [ ] Mobile apps (React Native)

---

## 🙏 Acknowledgments

- **OpenAlex** - For providing free, open access to scholarly data
- **Fiber Framework** - For the fast and elegant Go web framework
- **React Team** - For the powerful frontend library
- **Vite** - For the lightning-fast build tool
- **Tailwind CSS** - For the utility-first CSS framework
- **All contributors** - For making this project better

---

## 🌟 Star History

If you find this project useful, please consider giving it a ⭐ on GitHub!

---

<div align="center">

**Built with ❤️ by the ScholarWeave Team**

[⬆ Back to Top](#scholarweave-)

</div>
