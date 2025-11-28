# Personal Website

Personal portfolio website built with PocketBase, Go, htmx, Alpine.js, and Templ.

## Tech Stack

**Backend:**
- PocketBase v0.30+ (app framework, database, admin UI)
- Go 1.25+ (handlers and business logic)
- Templ (type-safe Go templates)
- SQLite (database via PocketBase)

**Frontend:**
- htmx 2.0 (dynamic updates)
- Alpine.js 3.14 (client-side interactivity)
- TailwindCSS 4.x (styling)

**Development:**
- Air (live reload)
- Docker Compose (local environment)
- Make (task automation)

## Quick Start

### Prerequisites

- [asdf](https://asdf-vm.com/) for version management
- Docker and Docker Compose

### Setup

```bash
# Install dependencies
make asdf-install

# Start development server
make dev
```

The application will be available at:
- **App**: http://localhost:8090
- **Admin UI**: http://localhost:8090/_/

### Available Commands

```bash
make help              # Show all available commands
make dev               # Start development server
make dev-build         # Rebuild Docker images
make docker-logs       # View logs
make docker-clean      # Clean up volumes
make tidy              # Tidy Go modules
make test              # Run tests
make fmt               # Format code
```

## Project Structure

```
personal-website/
├── main.go                      # PocketBase app initialization
├── internal/
│   ├── models/                  # Data models
│   ├── services/                # Business logic
│   └── handlers/                # HTTP handlers
├── pb_migrations/               # Database migrations
├── web/
│   ├── templates/               # Templ files
│   │   ├── base.templ
│   │   └── components/
│   └── static/
│       ├── css/
│       ├── js/
│       └── images/
└── pb_data/                     # PocketBase data (gitignored)
```

## Architecture

See [PERSONAL_WEBSITE_ARCHITECTURE.md](/Users/damien/Projects/planning-poker/PERSONAL_WEBSITE_ARCHITECTURE.md) for detailed architecture documentation.

## Development Workflow

1. **Local Development**: `make dev` starts Air with hot reload
2. **Database Admin**: Access PocketBase admin at `http://localhost:8090/_/`
3. **Edit Templates**: Templ files auto-recompile on save
4. **Style Changes**: Tailwind watches `input.css` and templates

## License

MIT
