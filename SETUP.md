# Scholarweave Setup & Installation Guide

## Prerequisites & Installation

### Backend (Go API)

#### Required
- **Go 1.25+** — Download from [golang.org](https://golang.org/dl)
  - Verify: `go version`
  - On Windows, add Go `bin/` to PATH (usually done automatically by installer)

#### Installation Steps
1. Install Go 1.25 or newer
2. Clone/navigate to the project root: `cd n:\Projects\Scholarweave`
3. Download dependencies: `go mod download`
4. Build the binary: `go build -o scholarweave.exe`
5. Run the server: `.\scholarweave.exe`
   - Server starts on `http://localhost:3000`
   - Health check: `curl http://localhost:3000/api/v1/health`

#### Optional (Recommended for Development)
- **golangci-lint** — Linter for Go code quality checks
  - Install: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
  - Run: `golangci-lint run ./...`
- **Air** — Hot-reload development server
  - Install: `go install github.com/cosmtrek/air@latest`
  - Run from project root: `air`
  - Watches for file changes and auto-rebuilds

---

### Frontend (React + Vite + Tailwind)

#### Required
- **Node.js 16+** — Download from [nodejs.org](https://nodejs.org)
  - Verify: `node --version` and `npm --version`
  - On Windows, use the official installer or Chocolatey: `choco install nodejs`

#### Installation Steps
1. Navigate to frontend directory: `cd n:\Projects\Scholarweave\frontend`
2. Install dependencies: `npm install`
3. Start dev server: `npm run dev`
   - Frontend runs on `http://localhost:5173` (default Vite port)
   - Hot module reloading (HMR) enabled by default
4. Build for production: `npm run build`
   - Outputs to `frontend/dist/`

#### Optional (Recommended for Development)
- **VS Code Extension: ES7+ React/Redux/React-Native snippets** — Speed up component writing
- **VS Code Extension: Tailwind CSS IntelliSense** — Autocomplete for Tailwind classes
- **prettier** — Code formatter (should be in `devDependencies`)
  - Run: `npm run format` (if configured in `package.json`)

---

## Running the Full Stack (Local Development)

### Terminal 1: Backend (Go API)
```powershell
cd n:\Projects\Scholarweave
go run main.go
# or after building: .\scholarweave.exe
```

### Terminal 2: Frontend (React Dev Server)
```powershell
cd n:\Projects\Scholarweave\frontend
npm run dev
```

Then:
- Open browser to `http://localhost:5173`
- Frontend will proxy API calls to `http://localhost:3000/api/v1` (configured in `.env` or Vite proxy)

---

## Future Development: Tools & Libraries to Consider

### Backend Enhancements
- **Database**: `gorm` (ORM) or `sqlc` (SQL-first) for persistent storage (papers, users, searches)
- **Authentication**: `github.com/golang-jwt/jwt` for JWT tokens or `ory/kratos` for full auth service
- **Caching**: `go-redis` for Redis caching of OpenAlex results (avoid rate limits)
- **Rate Limiting**: `github.com/grpc-ecosystem/go-grpc-middleware` or `tollbooth` for HTTP rate limiter
- **Structured Logging**: `go.uber.org/zap` or `sirupsen/logrus` (replace simple `log` package)
- **Config Management**: `spf13/viper` for environment-based config files
- **Testing**: `testify` and `httptest` for unit and integration tests
- **Async Tasks**: `hibiken/asynq` for background job queues (e.g., periodic OpenAlex syncs)
- **GraphQL** (optional): `99designs/gqlgen` if you want a GraphQL layer instead of REST
- **Containerization**: Docker + `Dockerfile` to run backend in containers

### Frontend Enhancements
- **State Management**: `zustand` (lightweight) or `redux-toolkit` (larger apps) for global state
- **API Client**: `axios` or `react-query`/`@tanstack/react-query` for server state and caching
- **Routing**: `react-router-dom@latest` for multi-page navigation
- **UI Component Library** (optional): `shadcn/ui` (headless + Tailwind), `material-ui`, or `headless-ui` for pre-built components
- **Form Validation**: `react-hook-form` + `zod` for type-safe forms
- **Testing**: `vitest` (Vite-native), `@testing-library/react` for unit tests, `cypress` or `playwright` for E2E
- **Environment Config**: `.env.local`, `.env.development`, `.env.production` for API endpoints per environment
- **PWA / Offline**: `workbox` or `vite-plugin-pwa` for Progressive Web App support
- **Analytics**: `posthog` or `mixpanel` for user behavior tracking

### DevOps & CI/CD
- **GitHub Actions**: Automate `go build`, `go test`, `go vet`, `npm run build`, and `npm run test` on every push
- **Docker Compose**: Multi-service setup (backend, frontend, optional database) in one command
- **Vercel / Netlify**: Deploy frontend (automatic deployments on push to `main`)
- **Railway / Render**: Deploy backend (easy Go hosting)
- **Database Hosting**: AWS RDS, Heroku Postgres, or MongoDB Atlas for production
- **S3 / Cloud Storage**: For file uploads (if papers have PDFs or datasets)

### Monitoring & Observability
- **Prometheus**: Metrics collection from Go backend
- **Grafana**: Dashboard for visualizing metrics
- **Sentry**: Error tracking and crash reporting (backend + frontend)
- **ELK Stack** or **Loki**: Centralized logging

### Project Additions (High Priority)
1. **.github/workflows/ci.yml** — Automated tests and builds on every PR
2. **Dockerfile + docker-compose.yml** — Easy local and production deployment
3. **database schema** — Store papers, users, search history
4. **Authentication** — User login, saved searches, favorites
5. **Rate limiting** — Protect OpenAlex API from abuse
6. **.env files** — Config per environment (dev, staging, prod)
7. **Unit tests** — Backend handlers, frontend components
8. **README.md updates** — API documentation, deployment guide

---

## Environment Variables (Setup for Future)

Create `backend/.env` (or set via system environment):
```
PORT=3000
OPENALEX_BASEURL=https://api.openalex.org
LOG_LEVEL=info
ENVIRONMENT=development
```

Create `frontend/.env`:
```
VITE_API_BASE_URL=http://localhost:3000/api/v1
```

---

## Quick Checklist

- [ ] Install Go 1.25+
- [ ] Install Node.js 16+
- [ ] Clone/navigate to project
- [ ] `go mod download` (backend)
- [ ] `npm install` (frontend, from `frontend/` directory)
- [ ] Run `go run main.go` (backend on port 3000)
- [ ] Run `npm run dev` (frontend on port 5173)
- [ ] Open `http://localhost:5173` and test search

---

## Troubleshooting

### Backend Won't Start
- Port 3000 in use? Change in `main.go`: `app.Listen(":3001")`
- Go not installed? Run `go version` to verify
- Missing dependencies? Run `go mod tidy`

### Frontend Won't Start
- Node/npm not installed? Run `node --version` and `npm --version`
- Port 5173 in use? Vite will auto-increment to 5174
- CORS error? Backend already has `AllowOrigins: ["*"]` in `main.go` — should work

### API Calls Fail
- Check backend is running on `http://localhost:3000`
- Check frontend `.env` has correct `VITE_API_BASE_URL`
- Check OpenAlex is reachable: `curl https://api.openalex.org/works?search=ai`

---

## Next Steps

1. **Immediate**: Set up both backend and frontend, confirm they run and communicate
2. **Short-term**: Add database (PostgreSQL + GORM) to persist papers and users
3. **Medium-term**: Add authentication (JWT), user accounts, saved searches
4. **Long-term**: CI/CD, production deployment, monitoring, multi-source paper APIs
