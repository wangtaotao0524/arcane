---
sidebar_position: 2
title: Building from Source
---

# Building Arcane from Source

This guide explains how to build the Arcane application from its source code. This is useful if you want to contribute to development, test unreleased features, or create custom builds.

## Prerequisites

- Node.js: Version 22 or higher.
- npm, yarn, pnpm, or bun: A Node.js package manager. Examples will use `npm`.
- Git: Required to clone the repository.
- Docker Engine: Required if you intend to build the Docker image, and test functionality locally.

## Steps

1.  **Clone the Repository:**
    Open your terminal and clone the Arcane repository:

    ```bash
    git clone https://github.com/ofkm/arcane
    cd arcane
    ```

2.  **Install Dependencies:**
    Install the necessary project dependencies using your preferred package manager:

    ```bash
    npm install
    ```

3.  **Run the Development Server (Optional):**
    To run Arcane in development mode with hot-reloading:

    ```bash
    npm run dev
    ```

    This will start a local development server, typically accessible at `http://localhost:3000` .

4.  **Linting and Formatting:**
    Before building, you might want to check for code style issues:

    ```bash
    npm run lint
    npm run format
    ```

5.  **Create a Production Build:**
    To build the optimized production version of Arcane:

    ```bash
    npm run build
    ```

    This command uses SvelteKit's build process (powered by Vite) and the `@sveltejs/adapter-node` adapter. The output will be placed in the `build/` directory by default. This directory contains the standalone Node.js server and static assets needed to run Arcane.

6.  **Run the Production Build:**
    After a successful build, you can run the production server:

    ```bash
    node build/index.js
    ```

    Arcane should now be running using the built artifacts, typically accessible at `http://localhost:3000` (or the port configured for the production environment).

## Building the Docker Image

The repository includes a `Dockerfile` to containerize the application.

1.  **Ensure Docker is Running:** Make sure your Docker daemon is active.
2.  **Build the Image:** From the root of the project directory, run:

    ```bash
    docker build -t arcane-local .
    ```

    You can replace `arcane-local` with your preferred image tag.

3.  **Run the Docker Container:**

    See the [Quickstart](/docs/getting-started/quickstart) guide on how to run the docker container.

## Summary

This covers the essential steps for building and running Arcane from its source code, both as a standalone Node.js application and as a Docker image.
