# Arcane - Docker Management UI

<p align="center">
  <img src=".github/assets/arcane.png" alt="Arcane Logo" width="200"/>
</p>

[![Status](https://img.shields.io/badge/status-pre--release-orange)](https://shields.io/)

**⚠️ Note:** Arcane is currently pre-release software. Features may change, and bugs are expected. Use with caution, especially in production environments.

Arcane is a web-based user interface built with SvelteKit designed to simplify interaction with your Docker environment. It provides a clean overview and management capabilities for your containers, images, volumes, and networks.

## Features

- **Containers:** List, inspect, start, stop, restart, and remove containers. View container logs. Create containers (basic implementation).
- **Images:** List, pull, and remove images. Prune unused images.
- **Volumes:** List, create, and remove volumes.
- **Networks:** List, inspect, create, and remove networks.
- **Stacks:** Manage Docker Compose stacks (via `docker-compose` CLI integration - requires `docker-compose` accessible in the container/host).
- **Settings:** Configure Docker connection details (Host/Socket Path) and application preferences like polling interval.
- **Responsive UI:** Built with shadcn-svelte and Tailwind CSS for a modern look and feel on various devices.

## Documentation

For detailed instructions on getting started, configuration, development, and more, please visit the **[official documentation site](https://docs.example.com/arcane)** (Link placeholder).

## Important Notes

- **Security:** Mounting the Docker socket (`/var/run/docker.sock`) into any container grants it root-level access to your Docker host. Understand the security implications before running Arcane or any container with socket access.
- **Pre-release:** This software is under active development. Expect breaking changes and bugs.
- **Data:** Stack definitions and application settings are stored in the mapped data volume (`./arcane-data` by default when using Docker Compose). Back up this directory if needed.
