# ScholarWeave - Detailed Project Report

**Date:** December 10, 2025  
**Status:** Functional MVP (Minimum Viable Product)  
**Repository:** [Prithiv-0/Scholarweave](https://github.com/Prithiv-0/Scholarweave)

---

## Executive Summary

**ScholarWeave** is a full-stack web application for searching and discovering academic papers powered by the **OpenAlex API**. It combines a Go REST API backend with a modern React + Vite + Tailwind CSS frontend, enabling users to search millions of academic papers with a responsive, intuitive interface.

### Key Features
- ✅ Real-time paper search across OpenAlex database (~150M+ papers)
- ✅ Rich paper detail pages with abstracts, authors, DOI links, and citation counts
- ✅ Health monitoring and API status dashboard
- ✅ Responsive design (mobile, tablet, desktop)
- ✅ Fast dev experience with Vite HMR
- ✅ Type-safe frontend (TypeScript/React)
- ✅ Clean REST API with error handling

---

## Project Architecture

```
ScholarWeave/
├── backend/                    # Go API server
│   ├── main.go                # Entry point (Fiber setup, routes)
│   ├── go.mod / go.sum        # Go dependencies
│   ├── api/
│   │   └── handlers/
│   │       └── openalex.go    # OpenAlex client & handlers
│   └── internal/
│       ├── models/
│       │   └── paper.go       # Paper & Author data models
│       └── services/
│           └── search_service.go  # Paper normalization logic
│
├── frontend/                   # React + Vite + Tailwind
│   ├── package.json           # Dependencies
│   ├── vite.config.ts         # Vite config with API proxy
│   ├── tailwind.config.cjs    # Tailwind CSS config
│   ├── tsconfig.json          # TypeScript config
│   ├── index.html             # HTML entry point
│   └── src/
│       ├── main.tsx           # React root
│       ├── App.tsx            # Route setup (React Router)
│       ├── index.css          # Global styles
│       ├── api/
│       │   └── client.ts      # Axios HTTP client & API methods
│       ├── components/
│       │   ├── SearchBox.tsx  # Search input form
│       │   ├── PaperCard.tsx  # Paper result card component
│       │   └── Health.tsx     # API health status widget
│       └── pages/
│           ├── Search.tsx     # Main search page (route: /)
│           └── PaperDetail.tsx # Paper detail page (route: /papers/:id)
│
├── SETUP.md                   # Installation & development guide
├── .env.example               # Environment variable template
├── npm.bat / npm.ps1          # Windows PATH helpers for Node.js
└── PROJECT_REPORT.md          # This file
```

---

## Technology Stack

### Backend (Go)
| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Framework** | Fiber | v3.0.0-rc.1 | HTTP web framework (Express.js-like) |
| **Language** | Go | 1.25 | System language |
| **HTTP Client** | net/http | stdlib | Calls OpenAlex API |
| **JSON** | encoding/json | stdlib | Request/response serialization |
| **Middleware** | Fiber built-in | - | CORS, logging, recovery, error handling |
| **External API** | OpenAlex | - | Academic paper metadata source |

### Frontend (React)
| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Framework** | React | 18.2.0 | UI component library |
| **Build Tool** | Vite | 5.0.8 | Fast bundler with HMR |
| **Language** | TypeScript | 5.3.3 | Type-safe JavaScript |
| **Styling** | Tailwind CSS | 3.4.1 | Utility-first CSS framework |
| **Routing** | React Router | 6.14.1 | Client-side navigation |
| **HTTP Client** | Axios | 1.6.0 | Promise-based HTTP requests |
| **CSS Processing** | PostCSS + Autoprefixer | - | CSS vendor prefixing |

### DevOps & Tools
- **Git** — Version control
- **GitHub** — Repository hosting
- **Node.js** — JavaScript runtime (v24.11.1 in dev environment)
- **npm** — Package manager (v11.6.2 in dev environment)

---

## API Specification

### Backend Endpoints

#### 1. Health Check
```
GET /api/v1/health
```
**Response (200 OK):**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2025-12-10T01:00:00+00:00",
  "services": {
    "api": "healthy",
    "openalex": "connected"
  }
}
```

#### 2. Search Papers
```
GET /api/v1/papers/search?q={query}
```
**Query Parameters:**
- `q` (required) — Search query (title, author, keywords, etc.)

**Response (200 OK):**
```json
{
  "meta": { "count": 2545637 },
  "results": [
    {
      "id": "https://openalex.org/W2101234009",
      "title": "Scikit-learn: Machine Learning in Python",
      "abstract": "Scikit-learn is a Python module...",
      "doi": "https://doi.org/10.48550/arxiv.1201.0490",
      "authors": [
        { "name": "Fabián Pedregosa" },
        { "name": "Gaël Varoquaux" }
      ],
      "cited_by_count": 62986,
      "source": "OpenAlex"
    }
  ]
}
```

#### 3. Get Paper Details
```
GET /api/v1/papers/{id}
```
**Path Parameters:**
- `id` — Paper OpenAlex ID or full URL (e.g., `https://openalex.org/W2101234009`); URL-encoded

**Response (200 OK):**
```json
{
  "id": "https://openalex.org/W2101234009",
  "title": "Scikit-learn: Machine Learning in Python",
  "abstract": "Full abstract text (reconstructed from inverted index if needed)...",
  "doi": "https://doi.org/10.48550/arxiv.1201.0490",
  "authors": [...],
  "cited_by_count": 62986,
  "source": "OpenAlex"
}
```

---

## Frontend Pages & Components

### Page: Search (`/`)
**Purpose:** Main landing page; search and browse papers.

**Features:**
- Search input box with "Search" button
- Real-time health status indicator (green = OK, red = error)
- Paper result cards in responsive grid (3 columns on desktop, 1 on mobile)
- Loading skeleton UI while fetching results
- Error messages if search fails
- Empty state prompts when no results

**Components Used:**
- `SearchBox` — Controlled input + submit handler
- `Health` — Auto-refreshing health status
- `PaperCard` × N — Clickable result cards

### Page: Paper Detail (`/papers/:id`)
**Purpose:** View full metadata for a single paper.

**Features:**
- Back link to search
- Full paper title, authors, abstract
- DOI as clickable external link (opens in new tab)
- Citation count and source badge
- "Open DOI" button (links to external DOI or arXiv)
- Error handling if paper not found or API fails

**Components Used:**
- React Router's `useParams` hook to extract `:id`
- `useEffect` to fetch paper on mount

### Component: SearchBox
**Props:**
- `onSearch: (query: string) => void` — Callback when form submitted
- `loading?: boolean` — Disable button while loading

**Behavior:**
- Text input with placeholder
- Form submission on Enter or button click
- Disabled state during loading

### Component: PaperCard
**Props:**
- `paper: Paper` — Paper object
- `onClick?: () => void` — Callback when card clicked

**Display:**
- Paper title (clipped to 2 lines)
- Abstract excerpt (first ~160 characters + "...")
- Authors (first 2 + count of rest)
- DOI (clickable link if present)
- Citation count
- Source badge (OpenAlex, etc.)

### Component: Health
**Props:** None

**Behavior:**
- Fetches `/api/v1/health` on mount
- Auto-refreshes every 30 seconds
- Displays green/red status indicator
- Shows version, API status, OpenAlex status
- Error state with message if API unreachable

---

## Data Models

### Paper (Frontend & Backend)
```typescript
interface Paper {
  id: string                  // OpenAlex work URL
  title: string              // Paper title
  abstract: string           // Full abstract text
  doi: string                // DOI URL (if available)
  authors: Author[]          // List of authors
  cited_by_count: number     // Citation count
  source: string             // "OpenAlex", etc.
}

interface Author {
  name: string               // Author display name
  id?: string                // OpenAlex author ID
  orcid?: string             // ORCID (optional)
}
```

### OpenAlex Work (from API)
```json
{
  "id": "https://openalex.org/W...",
  "title": "...",
  "authorships": [
    {
      "author": {
        "id": "https://openalex.org/A...",
        "display_name": "...",
        "orcid": "..."
      }
    }
  ],
  "abstract": "...",
  "abstract_inverted_index": {
    "word1": [0, 5],
    "word2": [1, 3]
  },
  "doi": "https://doi.org/...",
  "cited_by_count": 12345
}
```

---

## Current Implementation Status

### ✅ Completed Features
1. **Backend API** (Go + Fiber)
   - ✅ HTTP server on `:3000`
   - ✅ CORS enabled for all origins
   - ✅ Health endpoint returning service status
   - ✅ Search endpoint proxying to OpenAlex
   - ✅ Paper detail endpoint with single work lookup
   - ✅ Error handling with structured JSON responses
   - ✅ Abstract reconstruction from `abstract_inverted_index` (when `abstract` field is empty)
   - ✅ URL decoding and ID extraction (handles full OpenAlex URLs)
   - ✅ Request logging with timestamps and latency

2. **Frontend UI** (React + Vite + Tailwind)
   - ✅ Search page with live results
   - ✅ Paper detail page with routing
   - ✅ Responsive grid layout (mobile-first)
   - ✅ Paper card component showing title, authors, abstract, DOI, citations
   - ✅ Health status widget with auto-refresh
   - ✅ Error boundaries and fallback messages
   - ✅ Loading states with skeleton animations
   - ✅ React Router for SPA navigation
   - ✅ TypeScript for type safety

3. **Data Processing**
   - ✅ Normalization of OpenAlex responses into internal `Paper` shape
   - ✅ Author extraction and transformation
   - ✅ Abstract reconstruction from inverted index
   - ✅ URL encoding/decoding for safe API calls

### 🔄 In Progress / Known Limitations
1. **No Persistence** — All data from OpenAlex; no database
2. **No Authentication** — Public API, no user accounts
3. **Limited Search** — OpenAlex search limited to 10 results per query
4. **No Caching** — Every search hits OpenAlex; potential rate limiting on high traffic
5. **CORS Wide Open** — `AllowOrigins: ["*"]` acceptable for MVP, but not production-safe
6. **No Pagination** — Fixed 10 results per search

### 🚀 Recommended Next Steps
1. **Database Integration** — PostgreSQL + GORM to store papers and user searches
2. **User Accounts** — JWT authentication for saved searches, favorites
3. **Caching Layer** — Redis to cache popular searches and paper details
4. **Rate Limiting** — Protect backend from abuse (both OpenAlex and user-facing)
5. **Tests** — Unit tests for backend handlers, frontend components
6. **CI/CD** — GitHub Actions to run tests, build, and deploy
7. **Pagination** — Implement offset/limit or cursor-based pagination for large result sets
8. **Advanced Search** — Filters by author, year, citation count, field
9. **Export Options** — CSV, JSON, BibTeX export of search results
10. **Analytics** — Track popular searches, user engagement

---

## Running the Project

### Prerequisites
- **Go 1.25+** (backend)
- **Node.js 16+** (frontend)
- **npm** (Node package manager)
- **Git** (version control)

### Quick Start

#### Terminal 1: Start Backend
```powershell
cd n:\Projects\Scholarweave
go run main.go
```
Output should show:
```
2025/12/10 12:00:00 Starting ScholarWeave API on http://localhost:3000
```

#### Terminal 2: Start Frontend
```powershell
$env:PATH = "N:\Apps\node;$env:PATH"
cd n:\Projects\Scholarweave\frontend
npm run dev
```
Output should show:
```
VITE v5.4.21 ready in XXX ms
➜ Local: http://localhost:5173/
```

#### Open Browser
Navigate to `http://localhost:5173` and start searching!

---

## Development Workflow

### Adding a New Feature (Example: Favorite Papers)

1. **Backend (Go):**
   - Add `favorite: bool` field to `Paper` model
   - Create `PATCH /api/v1/papers/:id/favorite` endpoint
   - Implement toggle logic (database required first)

2. **Frontend (React):**
   - Add favorite icon/button to `PaperCard` component
   - Call backend `PATCH` endpoint on click
   - Update local state to reflect favorite status
   - Persist to localStorage or backend

3. **Test:**
   - Search for a paper, click favorite, verify UI updates
   - Refresh page, confirm favorite persists (if using backend)

### Debugging

**Backend:**
- Check Go server logs in terminal 1
- Use Postman/curl to test endpoints:
  ```bash
  curl http://localhost:3000/api/v1/health
  curl "http://localhost:3000/api/v1/papers/search?q=machine%20learning"
  ```

**Frontend:**
- Open browser DevTools (F12)
- Check Network tab for API calls and status codes
- Check Console for JavaScript errors
- Use React DevTools extension to inspect component state

---

## File Structure & Responsibilities

| File | Responsibility | LOC |
|------|---|---|
| `main.go` | Server setup, middleware, route definitions | ~70 |
| `api/handlers/openalex.go` | OpenAlex API client, paper fetching, normalization | ~180 |
| `internal/models/paper.go` | Data structures (`Paper`, `Author`) | ~20 |
| `internal/services/search_service.go` | Data normalization, validation | ~35 |
| `frontend/src/App.tsx` | Route setup (React Router) | ~20 |
| `frontend/src/pages/Search.tsx` | Search page logic, state management | ~130 |
| `frontend/src/pages/PaperDetail.tsx` | Paper detail page, fetching | ~60 |
| `frontend/src/components/SearchBox.tsx` | Search input form | ~30 |
| `frontend/src/components/PaperCard.tsx` | Paper result card UI | ~40 |
| `frontend/src/components/Health.tsx` | Health status widget | ~50 |
| `frontend/src/api/client.ts` | Axios HTTP client, API methods | ~45 |

**Total:** ~620 lines of application code (excluding config, tests, dependencies)

---

## Performance & Scalability Notes

### Current Bottlenecks
1. **No Caching** — Every search queries OpenAlex (API latency ~1-3s)
2. **Single Server** — No horizontal scaling
3. **In-Memory** — No persistence; data lost on restart
4. **No CDN** — Frontend assets not cached globally

### Optimization Opportunities
1. **Redis Cache** — Store frequently searched terms and paper details
2. **Database** — Move search history and favorites to PostgreSQL
3. **Load Balancer** — Distribute traffic across multiple backend instances
4. **Static Site Generation** — Pre-cache popular paper pages
5. **API Rate Limiting** — Throttle requests, recover from OpenAlex rate limits gracefully

---

## Security Considerations

### Current State
- ✅ Error responses don't leak internal info (sanitized)
- ✅ HTTPS ready (no hardcoded passwords or secrets)
- ✅ Input validation on query parameters

### Recommendations
- ⚠️ **CORS** — Restrict to known domains in production
- ⚠️ **Rate Limiting** — Add user-based or IP-based limits
- ⚠️ **Authentication** — Require login for saving/exporting data
- ⚠️ **HTTPS/TLS** — Use in production
- ⚠️ **Environment Variables** — Move config to `.env` files (not committed)

---

## Deployment Options

### Option 1: Docker (Recommended)
```dockerfile
# Dockerfile (backend)
FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN go build -o scholarweave main.go

FROM alpine:latest
COPY --from=builder /app/scholarweave /usr/local/bin/
EXPOSE 3000
CMD ["scholarweave"]
```

### Option 2: Heroku / Railway
1. Create `Procfile`: `web: go run main.go`
2. Deploy: `git push heroku main`

### Option 3: AWS/GCP/Azure
- Deploy backend as serverless function or container
- Deploy frontend to CDN (S3 + CloudFront, or Vercel/Netlify)

---

## Maintenance & Support

### Regular Tasks
- Monitor OpenAlex API changes (breaking changes, rate limits)
- Update Go and Node.js dependencies monthly
- Review error logs for patterns
- Test on multiple browsers/devices

### Common Issues & Fixes

| Issue | Cause | Fix |
|-------|-------|-----|
| 404 on paper detail | Paper ID not URL-encoded | Frontend now encodes with `encodeURIComponent()` |
| No abstract shown | OpenAlex `abstract` field empty | Backend reconstructs from `abstract_inverted_index` |
| CORS errors | Frontend calls wrong API domain | Check `vite.config.ts` proxy setting |
| Port already in use | Previous server still running | Kill process: `Get-Process -Name main \| Stop-Process` |
| npm install fails | Node PATH not set (Windows) | Run: `$env:PATH = "N:\Apps\node;$env:PATH"` |

---

## Conclusion

**ScholarWeave** is a functional MVP demonstrating modern full-stack development:

- **Clean Architecture** — Separation of concerns (handlers, services, models)
- **Type Safety** — Go on backend, TypeScript on frontend
- **User Experience** — Fast, responsive, intuitive interface
- **Scalability** — Ready for database, caching, and authentication layers
- **Developer Experience** — Hot reload, clear error messages, comprehensive setup guide

**Next Phase:** Add persistence layer (database) and authentication to unlock saved searches, favorites, and multi-user capabilities.

---

**Repository:** [Prithiv-0/Scholarweave](https://github.com/Prithiv-0/Scholarweave)  
**Last Updated:** December 10, 2025  
**Status:** Active Development
