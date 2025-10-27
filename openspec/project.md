# Project Context

## Purpose
Arcane is a modern Docker management platform designed to be easy-to-use and built for everyone. The goal is to provide a community-driven Docker management solution that makes container management accessible to users with various levels of expertise.

## Tech Stack
- **Frontend**: Svelte 5, TypeScript, Vite, Tailwind CSS
- **Package Manager**: pnpm with workspace support
- **Build Tool**: Vite with SvelteKit
- **UI Components**: Custom UI library with Bits UI integration
- **State Management**: Svelte stores and runed utilities
- **HTTP Client**: Axios with custom interceptors
- **Code Editor**: CodeMirror integration
- **Testing**: Playwright for end-to-end testing
- **Containerization**: Docker and Docker Compose

## Project Conventions

### Code Style
- **TypeScript**: Strict mode enabled with module resolution "bundler"
- **File Naming**: Kebab-case for component files, camelCase for utilities
- **Import Organization**: Grouped by external libraries, then internal modules
- **Component Structure**: Svelte 5 syntax with runes for reactivity
- **Formatting**: Prettier with Svelte and Tailwind CSS plugins

### Architecture Patterns
- **Service Layer Pattern**: Abstract BaseAPIService class for HTTP operations
- **Component-Based Architecture**: Reusable UI components in lib/components
- **Store Pattern**: Centralized state management for user and configuration
- **API-First Design**: Frontend proxies to backend via /api endpoints
- **Modular Design**: Separated by domains (volumes, networks, containers, etc.)

### Testing Strategy
- **End-to-End Testing**: Playwright for integration testing
- **Test Environment**: Docker Compose setup for isolated test environments
- **Component Testing**: Svelte testing utilities (where applicable)
- **Automated Setup**: npm scripts for test environment management

### Git Workflow
- **Conventional Commits**: Using commitizen for standardized messages
- **Branch Strategy**: Feature branches with semantic versioning
- **Version Management**: Git cliff for changelog generation
- **Collaboration**: Crowdin integration for translations

## Domain Context
- **Docker Management**: Focus on container lifecycle management
- **Multi-User Support**: User authentication and role-based access
- **Project Organization**: Resources grouped by projects and environments
- **Real-time Monitoring**: WebSocket connections for live status updates
- **Template System**: Customizable deployment templates

## Important Constraints
- **Browser Compatibility**: Modern browsers with ES2020+ support
- **Performance**: Optimized for large Docker deployments
- **Security**: Authentication required for most operations
- **Accessibility**: WCAG compliance for inclusive design

## External Dependencies
- **Docker Engine**: Primary container runtime
- **Translation Platform**: Crowdin for internationalization
- **Icon Library**: Lucide icons for UI elements
- **UI Framework**: Tailwind CSS with custom variants
- **Date Handling**: date-fns for date manipulation
