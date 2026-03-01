# ScholarWeave — Semantic Knowledge Graph Implementation Plan

> **Created:** 2026-02-28  
> **Last Updated:** 2026-03-01  
> **Status:** Active (Phase F/G completed, code audit pass completed)

---

## 1) Progress Snapshot

### ✅ Implemented

- **Graph domain model is in place**
  - `internal/models/graph.go`
  - Node types: `paper`, `author`, `concept`, `institution`
  - Edge types: `authored`, `has_concept`, `cites`, `related`, `affiliated`

- **Paper model enriched for semantic graphing**
  - `internal/models/paper.go`
  - Added: concepts, institutions, referenced works, related works

- **OpenAlex normalization upgraded**
  - `api/handlers/openalex.go`
  - Parses concepts, referenced/related works, institutions into normalized paper output

- **Graph service implemented with multiple entry points**
  - `internal/services/graph_service.go`
  - `BuildPaperGraph(ctx, paperID, depth)`
  - `BuildAuthorGraph(ctx, authorID)`
  - `BuildConceptGraph(ctx, conceptID)`
  - `BuildSearchGraph(ctx, paperIDs)`

- **Graph API endpoints live**
  - `api/handlers/graph.go`, `main.go`
  - `GET /api/v1/graph/paper/:id?depth=1|2`
  - `GET /api/v1/graph/author/:id`
  - `GET /api/v1/graph/concept/:id`
  - `POST /api/v1/graph/search` with `{ "paper_ids": [...] }`

- **Frontend graph foundation implemented**
  - `frontend/src/components/KnowledgeGraph.tsx`
  - `frontend/src/pages/KnowledgeGraph.tsx`
  - `frontend/src/App.tsx`
  - `frontend/src/components/AppShell.tsx`
  - `frontend/src/pages/PaperDetail.tsx`
  - `frontend/src/pages/Search.tsx`

- **Graph UI features shipped**
  - Interactive force graph rendering (`react-force-graph-2d`)
  - Node-type filters and legend
  - Export actions: JSON + PNG
  - Graph explorer route: `/graph`
  - Share-link copy action in graph explorer
  - Graph history trail (pivot breadcrumbs)
  - Paper detail shortcut: “Explore Knowledge Graph”
  - Search page Graph View (builds graph from result paper IDs)
  - Reusable node detail panel component
  - Node inspector actions with pivots:
    - Open paper
    - Pivot to paper graph
    - Pivot to author graph
    - Pivot to concept graph

- **Backend graph quality improvements**
  - Stub paper node labels are lazily hydrated to real titles (bounded fetches)
  - Improves readability in citation/related subgraphs

- **Frontend client graph APIs complete**
  - `frontend/src/api/client.ts`
  - `getPaperGraph`, `getAuthorGraph`, `getConceptGraph`, `getSearchGraph`

- **Tests and build status**
  - Backend graph tests expanded:
    - `TestBuildPaperGraphDepthOne`
    - `TestBuildPaperGraphDepthValidation`
    - `TestBuildAuthorGraph`
    - `TestBuildConceptGraph`
    - `TestBuildSearchGraph`
  - `go test ./...` passing
  - `npm run test` passing
  - `npm run build` passing

### ✅ Code Audit & Hardening Pass (2026-03-01)

- **Security**
  - Error handler no longer leaks internal error messages (main.go)
  - JWT secret now refuses to start in production without `JWT_SECRET` set (config.go)
  - OpenAlex ID validation via strict regex — prevents path traversal (openalex.go, graph_service.go)
  - bcrypt password length capped at 72 chars (auth_service.go)
  - Password whitespace no longer silently trimmed (auth_service.go)
  - JSON-LD XSS vulnerability fixed with proper escaping (PaperDetail.tsx)

- **Bug Fixes**
  - Dual localStorage token key desync fixed — single source of truth via `authStorage` (AuthContext.tsx)
  - `rows.Err()` now checked after all `pgx` scan loops (library_service.go)
  - Fire-and-forget DB updates now log errors (library_service.go)
  - Health component now renders yellow/warning styling for degraded status (Health.tsx)
  - PaperCard DOI link no longer triggers parent card click (PaperCard.tsx)
  - `favoriteMessage` auto-clears after 3 seconds (PaperDetail.tsx)
  - Clipboard `writeText` wrapped in try/catch with fallback (KnowledgeGraph page)

- **Deterministic Output**
  - Non-deterministic map iteration replaced with sorted node/edge output in all graph builders (graph_service.go)
  - Edge sorting now uses source/target/relationship/weight tie-breakers for full determinism (graph_service.go)
  - `hydratePaperNodeLabels` now hydrates in sorted ID order (graph_service.go)

- **Context Propagation**
  - `fetchJSON` now accepts `context.Context` — requests cancel with HTTP handler (openalex.go)
  - Graph handlers pass Fiber context to service layer instead of `context.Background()` (graph.go)

- **Code Quality**
  - `catch (err: any)` replaced with `catch (err: unknown)` across all frontend files
  - Duplicated encode/decode ID pattern extracted to `normalizeId()` helper (client.ts)
  - Redundant CSS reset removed — Tailwind Preflight already covers it (index.css)
  - Redundant JS abstract truncation removed — CSS `line-clamp-3` suffices (PaperCard.tsx)
  - Search and graph pages now guard against stale async responses (request race safety)

- **Accessibility**
  - PaperCard: added `role="button"`, `tabIndex={0}`, keyboard handler
  - AppShell: navigation wrapped in `<nav aria-label="Main navigation">`
  - SearchBox: added `aria-label` to search input
  - Graph canvas: added `aria-label` to graph container
  - Search filters: added `aria-label` to year/subject filter inputs

---

## 2) Current Active Routes

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/v1/graph/paper/:id` | Paper-centric knowledge graph (depth optional) |
| GET | `/api/v1/graph/author/:id` | Author-centric graph |
| GET | `/api/v1/graph/concept/:id` | Concept-centric graph |
| POST | `/api/v1/graph/search` | Graph from selected paper IDs |

---

## 3) Remaining Work (Updated)

### Phase E — Graph Caching

- [x] **Skipped by product decision** (user requested no caching)

### Phase F — Export & Sharing

- [x] PNG export from graph canvas (`frontend/src/components/KnowledgeGraph.tsx`)
- [x] JSON export for current graph state
- [x] Shareable deep-link copy action in `/graph` page
- [ ] Optional backend graph RDF/Turtle export endpoint

### Phase G — UI/UX Hardening

- [x] Extracted reusable node detail panel component (`GraphNodeDetailPanel.tsx`)
- [x] Better labels for stub paper nodes (lazy title hydration)
- [x] Breadcrumb/history trail for graph pivots
- [x] Loading/error/empty graph states in graph components/pages

### Phase H — Additional Test Coverage

- [ ] Frontend component tests for `KnowledgeGraph.tsx`
- [ ] Frontend page tests for `KnowledgeGraph.tsx` route interactions
- [ ] API client tests for new graph methods

---

## 4) Revised Implementation Order

1. **Optional RDF/Turtle graph export endpoint**
2. **Frontend test expansion** (graph component/page/client)
3. **Minor UX polish** (if any issues surface in usage)

---

## 5) Notes on Current Design

- Author and concept graph builders are currently implemented via OpenAlex filtered works queries, which is adequate for MVP and keeps integration simple.
- Search Graph mode in `Search.tsx` uses live search result IDs and calls `/graph/search` directly, enabling immediate graph exploration from normal discovery workflows.
- The graph explorer page now supports query params for all contexts:
  - `?paper=<id>`
  - `?author=<id>`
  - `?concept=<id>`

---

## 6) Completion Checklist

| Area | State |
|---|---|
| Graph Models | ✅ Complete |
| Paper Graph API | ✅ Complete |
| Author Graph API | ✅ Complete |
| Concept Graph API | ✅ Complete |
| Search Graph API | ✅ Complete |
| Frontend Graph Explorer | ✅ Complete |
| Search Graph View | ✅ Complete |
| Node Pivot Workflows | ✅ Complete |
| Graph Caching | ⛔ Skipped (by decision) |
| Graph Export (PNG/JSON) | ✅ Complete |
| Graph Export (RDF/Turtle endpoint) | ⏳ Optional Pending |
| Graph UX Hardening (Phase G) | ✅ Complete |
| Code Audit & Hardening | ✅ Complete |
| Frontend Graph Tests | ⏳ Pending |
