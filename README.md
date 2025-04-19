# Arcane - Docker Management UI

Arcane is a web-based user interface built with SvelteKit to interact with and manage your Docker environment. It provides a clean overview of your containers, images, volumes, and networks, along with controls for common Docker operations.

## Features

- **Dashboard:** Overview of your Docker environment (coming soon).
- **Containers:** List, inspect, start, stop, restart, and remove containers. View container logs.
- **Images:** List and remove images.
- **Volumes:** List and manage Docker volumes.
- **Networks:** List and manage Docker networks.
- **Settings:** Configure Docker connection details (Host/Socket Path) and application preferences.
- **Responsive UI:** Built with shadcn-svelte and Tailwind CSS for a modern look and feel.

## Technology Stack

- **Frontend:** SvelteKit, TypeScript
- **UI Components:** shadcn-svelte
- **Styling:** Tailwind CSS
- **Docker Interaction:** Dockerode (Node.js client for the Docker Engine API)
- **Deployment:** Docker Compose

## Prerequisites

- **Docker:** The Docker Engine must be installed and running.
- **Docker Compose:** Required for running the application via the provided configuration.
- **Node.js & npm/pnpm/yarn:** Required only for local development.

## Getting Started (Docker Compose - Recommended)

This is the easiest way to run Arcane, as it bundles the application and its dependencies.

1.  **Clone the repository:**

    ```bash
    git clone <your-repository-url> arcane
    cd arcane
    ```

2.  **Docker Socket Access:**
    The provided `docker-compose.yml` file mounts the Docker socket (`/var/run/docker.sock`) into the Arcane container. This allows Arcane to communicate with your local Docker Engine. Ensure the path to the Docker socket is correct for your system if it differs.

3.  **Build and Run:**

    ```bash
    docker-compose up -d --build
    ```

    - `--build`: Builds the Docker image for Arcane based on the `Dockerfile`.
    - `-d`: Runs the containers in detached mode (in the background).

4.  **Access Arcane:**
    Open your web browser and navigate to `http://localhost:5173` (or the port specified in your `docker-compose.yml` if you changed it).

## Development (Local)

If you want to contribute or run the application locally for development:

1.  **Clone the repository:**

    ```bash
    git clone <your-repository-url> arcane
    cd arcane
    ```

2.  **Install dependencies:**

    ```bash
    npm install
    # or pnpm install / yarn install
    ```

3.  **Configure Docker Connection:**
    By default, Arcane will try to connect to the Docker socket at `/var/run/docker.sock`. If your Docker setup is different, you can modify the default in `src/lib/services/docker-service.ts` or configure it via the Settings page once the application is running (settings are saved to `app-settings.json` in the project root).

4.  **Start the development server:**
    ```bash
    npm run dev -- --open
    ```
    This will start the SvelteKit development server, typically on `http://localhost:5173`.

## Configuration

Application settings, including the Docker host connection string, can be configured via the **Settings** page within the Arcane UI. These settings are persisted in the `app-settings.json` file in the project's root directory. When running via Docker Compose, this file will exist _inside_ the container unless you map a volume to persist it externally.

## Building for Production (Standalone)

While Docker Compose is recommended, you can build a production version of the SvelteKit app:

```bash
npm run build
```

This requires a Node.js environment to run. You'll need an adapter (like `adapter-node`) configured in `svelte.config.js` for standalone deployment. See the [SvelteKit documentation on adapters](https://kit.svelte.dev/docs/adapters) for more details. You can preview the build locally with `npm run preview`.
