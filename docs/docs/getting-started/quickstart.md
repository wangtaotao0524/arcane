---
sidebar_position: 1
title: Quick Start
---

# Quick Start Guide

This guide provides the fastest way to get Arcane up and running using Docker Compose, the recommended method.

## Prerequisites

- `Docker and Docker Compose`

## Steps

1.  **Create `docker-compose.yml`:**
    Create a `docker-compose.yml` file with the following content:

```yaml
services:
  arcane:
    image: ghcr.io/ofkm/arcane:latest
    container_name: arcane
    ports:
      - '3000:3000'
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - arcane-data:/app/data
    environment:
      - PUID=1000
      - PGID=1000
      - DOCKER_GID=998 # getent group docker | cut -d: -f3
    restart: unless-stopped

volumes:
  arcane-data:
  driver: local
```

You may need to modify the environment variables to fit your setup. Mainly the `DOCKER_GID` variable.

2.  **Review Configuration & Permissions:**
    Before starting, review the `docker-compose.yml` file:

    - **Docker Socket:** It mounts `/var/run/docker.sock` (read-only) to allow Arcane to manage Docker.

    - **Data Persistence:**: You can mount a volume or local mount to `/app/data` inside the container. This will store Arcane's settings, stacks, users, sessions, and encryption keys.

    - **Permissions (Important):**

      - The `PUID` and `PGID` (default `1000`) control the ownership of the `/app/data` volume inside the container. Adjust if your user ID/group ID is different.

      - You **must** set `DOCKER_GID` to match the group ID of the Docker socket (`/var/run/docker.sock`) on your host machine. This allows Arcane to communicate with Docker. Find your Docker group ID using one of these commands in your terminal:
        - Linux: `getent group docker | cut -d: -f3`
        - Linux (alternative): `stat -c '%g' /var/run/docker.sock`
        - macOS (if Docker group exists): `dscl . -read /Groups/docker PrimaryGroupID | awk '{print $2}'` (Often not needed on standard Docker Desktop for Mac setups).
      - Update the `DOCKER_GID=999` line in the `docker-compose.yml` file with the correct ID.

3.  **Start Arcane:**
    Open your terminal, navigate to the directory where you saved `docker-compose.yml`, and run:

    ```bash
    docker-compose up -d
    ```

4.  **Access Arcane:**
    Open your web browser and navigate to:

    `http://localhost:3000`

    You should now see the Arcane UI, connected to your local Docker environment.

## Next Steps

- Explore the Arcane interface to manage your containers, images, volumes, and networks.
- Learn how to customize Arcane's behavior in the **[Configuration](./configuration.md)** guide.
