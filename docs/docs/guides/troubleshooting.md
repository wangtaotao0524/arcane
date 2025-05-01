---
sidebar_position: 1
title: Troubleshooting
---

### Docker Socker Permission Issues

Make sure you run:

- Linux: `getent group docker | cut -d: -f3`
- macOS (if Docker group exists): `dscl . -read /Groups/docker PrimaryGroupID | awk '{print $2}'`

This will give you the GID of the local docker group.

Then pass that to the DOCKER_GID environment variable in your docker-compose.yml file.

For example: `DOCKER_GID=998`
