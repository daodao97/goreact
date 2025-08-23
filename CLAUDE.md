# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoReact is a server-side rendering (SSR) framework that combines a Go backend with React frontend, designed for high performance and maintainable full-stack development. The framework provides both development-time hot module replacement (HMR) and production-optimized server-side rendering.

## Development Commands

### Essential Build Commands
- `go run cmd/goreact/main.go` - Builds JavaScript/CSS assets and starts development mode
- `go mod tidy` - Install Go dependencies
- `npm install` - Install Node.js dependencies (automatically run during JS build if package.json changes)

### Frontend Build Process
The build system automatically:
1. Copies `frontend/` to a temporary directory
2. Generates entry files for React components in `pages/` directory  
3. Builds client-side bundles with esbuild (ES modules, code splitting)
4. Builds server-side bundles with V8Go polyfills
5. Compiles TailwindCSS using `npx @tailwindcss/cli`
6. Copies `frontend/public/` to `build/` directory

### Development Mode Features
- **Hot Module Replacement**: File changes in `frontend/`, `locales/`, or `package.json` trigger automatic rebuilds
- **Intelligent Caching**: Only rebuilds when content actually changes, not just timestamps
- **Live Reload**: Browser automatically refreshes via Server-Sent Events at `/hmr`

### Lint and Type Checking
The project uses Go's standard tooling. Before committing changes, run:
- `gofmt -w .` - Format Go code
- `go vet ./...` - Static analysis
- `go build ./...` - Verify compilation

## Architecture Overview

### Core Architecture Pattern
GoReact implements a hybrid SSR architecture:

1. **Request Handling**: Gin web server receives HTTP requests and routes to template renderer
2. **Server Rendering**: V8Go engine executes React components server-side using generated JavaScript bundles
3. **Client Hydration**: Browser receives pre-rendered HTML, then React "hydrates" for interactivity
4. **Development Flow**: File watchers trigger esbuild compilation and HMR updates

### Directory Structure and Key Components

#### `/server/` - Core Server Logic
- `server.go`: Gin server setup, CORS, security headers, static asset serving
- `react_render.go`: Server-side React rendering using V8Go engine
- `build.go`: Build orchestration, file watching, cache management  
- `esbuild.go`: JavaScript/TypeScript compilation with alias resolution
- `hmr.go`: Hot module replacement system with WebSocket-like SSE
- `template.go`: HTML template rendering integration

#### `/base/` - Application Features
- `login/`: Authentication system with OAuth providers (Google, GitHub)
- `invite/`: User invitation system
- `privacy.go`, `upload.go`: Core application features

#### `/auth/`, `/dao/`, `/model/` - Data Layer
- JWT-based authentication with cookie storage
- Database access objects for users, API tokens, balances
- Data models and schema definitions

#### `/conf/` - Configuration Management
- `conf.go`: Application configuration structure
- `website.go`: Website-specific settings, i18n, analytics integration

#### `/i18n/` - Internationalization
- Multi-language support with dynamic translation loading
- Integrated with React rendering for localized SSR

### Build System Details

#### Frontend Asset Pipeline
The build system generates entry files dynamically:
- Scans `frontend/pages/` for `.jsx/.tsx` components
- Creates client entry: imports component + `renderPage` wrapper
- Creates server entry: imports component + `createServerRenderer` wrapper
- Uses path aliases: `@/` → `frontend/`, `#/` → `node_modules/goreact/ui/`

#### JavaScript Engine Integration  
Server-side rendering uses V8Go with polyfills:
- `TextEncoder/TextDecoder` for text manipulation
- `MessageChannel` for React concurrent features
- `process.env` for Node.js compatibility
- Injects global context: `window.location`, `INITIAL_PROPS`, `TRANSLATIONS`, etc.

#### File Watching and Caching
Intelligent build caching based on SHA256 content hashes:
- Tracks `frontend/` directory changes recursively
- Monitors `package.json` for dependency updates
- Debounced file system events (300ms throttling)
- Cache invalidation on build failures

### Development Workflow Integration

#### HMR Implementation
- SSE endpoint at `/hmr` for real-time updates
- Multi-client support with throttled event broadcasting
- Watches `frontend/`, `locales/`, `package.json` for changes
- Automatic page reload on successful rebuilds

#### Error Handling and Recovery
- Build failures clear caches to allow rebuild attempts
- Graceful degradation when V8Go rendering fails
- Detailed logging for debugging build and render issues

## Important Development Notes

### Code Architecture Constraints
- Go files should not exceed 250 lines (current violations in `build.go:538`, `hmr.go:443`)
- Each directory should contain max 8 files (some directories exceed this)
- Avoid circular dependencies between modules

### Security Considerations  
- JWT tokens stored as HTTP-only cookies
- CORS configured for development (allows all origins)
- Security headers: XSS protection, content sniffing prevention, frame denial
- Authentication providers configured in `conf.Website.AuthProvider`

### Performance Optimizations
- Code splitting enabled for client bundles
- Asset hashing for cache busting
- Gzip compression for static assets
- Build-time CSS optimization with TailwindCSS

### Frontend Integration Points
The React components must follow these conventions:
- Export default component from `frontend/pages/ComponentName.jsx`
- Use `@/` alias for internal imports, `#/` for UI library
- Access server data via `window.INITIAL_PROPS`
- Use `window.TRANSLATIONS` for i18n strings
- Check `window.ssr` to detect server-side rendering context