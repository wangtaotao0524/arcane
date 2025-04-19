#!/bin/sh

# Exit immediately if a command exits with a non-zero status.
set -e

# If we aren't running as root, just exec the CMD
if [ "$(id -u)" -ne 0 ]; then
    exec "$@"
fi

echo "Entrypoint: Running as root. Setting up user and permissions..."

# Default PUID/PGID if not provided
PUID=${PUID:-1000}
PGID=${PGID:-1000}
DOCKER_GID=${DOCKER_GID:-998}
APP_USER="arcane"
APP_GROUP="arcane"
APP_DIR="/app"

echo "Entrypoint: Using PUID=${PUID}, PGID=${PGID}, DOCKER_GID=${DOCKER_GID}"

# Create the arcane group if it doesn't exist
if ! getent group "$PGID" > /dev/null 2>&1; then
    echo "Entrypoint: Creating group ${APP_GROUP} with GID ${PGID}..."
    addgroup -g "$PGID" "$APP_GROUP"
else
    # If group with PGID exists, find its name
    GROUP_NAME=$(getent group "$PGID" | cut -d: -f1)
    echo "Entrypoint: Group with GID ${PGID} already exists: ${GROUP_NAME}"
fi

# Ensure the docker group exists with correct GID
if ! getent group docker > /dev/null 2>&1; then
    echo "Entrypoint: Creating docker group with GID ${DOCKER_GID}..."
    addgroup -g "$DOCKER_GID" docker
else
    CURRENT_DOCKER_GID=$(getent group docker | cut -d: -f3)
    if [ "$CURRENT_DOCKER_GID" != "$DOCKER_GID" ]; then
        echo "Entrypoint: Updating docker group from GID ${CURRENT_DOCKER_GID} to ${DOCKER_GID}..."
        groupmod -g "$DOCKER_GID" docker
    fi
fi

# Create the arcane user if it doesn't exist
if ! getent passwd "$PUID" > /dev/null 2>&1; then
    echo "Entrypoint: Creating user ${APP_USER} with UID ${PUID}..."
    adduser -D -u "$PUID" -G "$APP_GROUP" "$APP_USER"
else
    # If user with PUID exists, find their name
    USERNAME=$(getent passwd "$PUID" | cut -d: -f1)
    echo "Entrypoint: User with UID ${PUID} already exists: ${USERNAME}"
    # Rename user if it's not the expected name
    if [ "$USERNAME" != "$APP_USER" ]; then
        echo "Entrypoint: Renaming user from ${USERNAME} to ${APP_USER}..."
        usermod -l "$APP_USER" "$USERNAME"
    fi
fi

# Ensure arcane user is in the docker group
if ! id -nG "$APP_USER" | grep -qw "docker"; then
    echo "Entrypoint: Adding ${APP_USER} to the docker group..."
    addgroup "$APP_USER" docker
fi

# Fix permissions for the Docker socket if it exists
if [ -e "/var/run/docker.sock" ]; then
    echo "Entrypoint: Setting permissions for Docker socket..."
    chmod 666 /var/run/docker.sock
fi

# Change ownership of application directories
echo "Entrypoint: Ensuring ownership of ${APP_DIR} for ${PUID}:${PGID}..."
chown -R "$PUID":"$PGID" "$APP_DIR"

# Execute the command passed to the script (CMD) as the specified user
echo "Entrypoint: Switching to user ${APP_USER} (${PUID}:${PGID}) and executing command: $@"
exec su-exec "$APP_USER" "$@"
