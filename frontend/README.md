# ScholarWeave Frontend

React + Vite + Tailwind CSS application for searching and discovering academic papers powered by OpenAlex.

## Features

- 🔍 Full-text search for academic papers
- 📊 Real-time results with citation counts
- 🏥 API health status monitoring
- 📱 Responsive design (mobile, tablet, desktop)
- ⚡ Fast dev server with HMR (Hot Module Reloading)

## Prerequisites

- **Node.js 16+** — [Download](https://nodejs.org)
- **npm** — Bundled with Node.js
- Backend API running on `http://localhost:3000` — See root `SETUP.md`

## Installation & Setup

```bash
# Install dependencies
npm install

# Start dev server (runs on http://localhost:5173)
npm run dev

# Build for production
npm run build

# Preview production build locally
npm run preview
```

## Environment Variables

Create a `.env.local` file (optional, defaults are set in code):

```
VITE_API_BASE_URL=http://localhost:3000/api/v1
```

## Project Structure

```
frontend/
├── src/
│   ├── api/
│   │   └── client.ts          # Axios client & API methods
│   ├── components/
│   │   ├── SearchBox.tsx      # Search input component
│   │   ├── PaperCard.tsx      # Single paper card display
│   │   └── Health.tsx         # API health status
│   ├── pages/
│   │   └── Search.tsx         # Main search page
│   ├── App.tsx                # Root app component
│   ├── main.tsx               # Entry point
│   └── index.css              # Global styles + Tailwind
├── index.html                 # HTML template
├── package.json               # Dependencies & scripts
├── tsconfig.json              # TypeScript config
├── vite.config.ts             # Vite config + API proxy
├── tailwind.config.cjs        # Tailwind CSS config
└── postcss.config.cjs         # PostCSS config
```

## Available Scripts

- `npm run dev` — Start development server with HMR
- `npm run build` — Build optimized production bundle
- `npm run preview` — Preview production build
- `npm run lint` — Run ESLint (if configured)
- `npm run format` — Format code with Prettier

## API Integration

The frontend communicates with the backend API at `http://localhost:3000/api/v1`:

- **GET** `/health` — Check API status
- **GET** `/papers/search?q=<query>` — Search papers
- **GET** `/papers/:id` — Get paper details

All calls are handled by `src/api/client.ts` using Axios.

## Troubleshooting

### Port 5173 already in use?
Vite will automatically try the next available port (5174, 5175, etc.).

### CORS errors?
Make sure the backend is running and has CORS enabled (default in `main.go`).

### API calls failing?
Check that:
1. Backend is running: `go run main.go` (should print "Starting ScholarWeave API on http://localhost:3000")
2. `VITE_API_BASE_URL` in `.env.local` matches your backend URL
3. Network connectivity: `curl http://localhost:3000/api/v1/health`

### Dependencies not installing?
```bash
rm -r node_modules package-lock.json
npm install
```

## Future Enhancements

- [ ] React Router for multi-page navigation
- [ ] Zustand for global state management
- [ ] React Query for server state caching
- [ ] Unit tests with Vitest
- [ ] E2E tests with Playwright
- [ ] Dark mode toggle
- [ ] Paper detail modal or dedicated page
- [ ] Save favorites to local storage
- [ ] Export search results (CSV, JSON, BibTeX)

## Development Tips

- Use **TypeScript** for type safety — catch errors at build time
- Leverage **Tailwind classes** for rapid UI development — avoid custom CSS
- Use **React DevTools** browser extension for debugging component state
- Enable **Vite's network sharing** in dev: `vite --host` to test on other devices

## Deployment

### Build for production
```bash
npm run build
```

Outputs to `dist/` — ready to deploy to any static host:
- Vercel, Netlify (recommended for React)
- AWS S3 + CloudFront
- GitHub Pages
- Traditional web server (nginx, Apache)

**Note**: Update `VITE_API_BASE_URL` in build environment for production API endpoint.

## License

See root `LICENSE` file.
