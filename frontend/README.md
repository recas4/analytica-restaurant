# Frontend

Vue.js SPA für das analytica-restaurant System.

## Tech Stack

- Vue 3 + TypeScript
- Vite
- Tailwind CSS
- Pinia (State Management)
- Vue Router

## Projektstruktur

```
frontend/
├── src/
│   ├── api/            # API Client
│   ├── components/     # UI Components
│   ├── views/          # Page Views
│   ├── stores/         # Pinia Stores
│   ├── router/         # Vue Router Config
│   └── lib/            # Utilities
├── Dockerfile
└── nginx.conf
```

## Development

```bash
# Dependencies installieren
npm install

# Dev Server starten (http://localhost:5173)
npm run dev

# Production Build
npm run build

# Docker Build
docker build -t frontend .
```

## Deployment

Das Frontend wird als statische Files gebaut (`npm run build` → `dist/`) und per Nginx ausgeliefert. Die `nginx.conf` konfiguriert SPA-Routing (alle Routen → `index.html`) und Caching für statische Assets.

API-Requests (`/api/*`) werden vom separaten API-Gateway im Root-Verzeichnis (`/nginx.conf`) an die Backend-Services geroutet.

## Environment Variables

- `VITE_API_BASE_URL` - Backend API URL (default: `http://localhost/api`)

Für lokale Entwicklung `.env.local` erstellen:

```
VITE_API_BASE_URL=http://localhost/api
```
