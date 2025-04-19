#!/bin/bash

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
APP_USER="arcane"
APP_GROUP="arcane-group"
APP_DIR="/app"
# Optional: Define a data directory if you mount volumes for settings/data
# DATA_DIR="/app/data"

echo "Entrypoint: Using PUID=${PUID}, PGID=${PGID}"

# Create the group if it doesn't exist
if ! getent group "$PGID" > /dev/null 2>&1; then
    echo "Entrypoint: Creating group ${APP_GROUP} with GID ${PGID}..."
    addgroup -g "$PGID" "$APP_GROUP"
else
    # If group with PGID exists, find its name
    APP_GROUP=$(getent group "$PGID" | cut -d: -f1)
    echo "Entrypoint: Group with GID ${PGID} already exists: ${APP_GROUP}"
fi

# Create the user if it doesn't exist
if ! getent passwd "$PUID" > /dev/null 2>&1; then
    echo "Entrypoint: Creating user ${APP_USER} with UID ${PUID}..."
    adduser -D -u "$PUID" -G "$APP_GROUP" -h "$APP_DIR" -s /bin/bash "$APP_USER"
else
    # If user with PUID exists, find their name
    APP_USER=$(getent passwd "$PUID" | cut -d: -f1)
    echo "Entrypoint: User with UID ${PUID} already exists: ${APP_USER}"
    # Ensure the existing user is part of the target group
    if ! id -nG "$APP_USER" | grep -qw "$APP_GROUP"; then
        echo "Entrypoint: Adding existing user ${APP_USER} to group ${APP_GROUP}..."
        addgroup "$APP_USER" "$APP_GROUP"
    fi
fi

# Change ownership of application directories
echo "Entrypoint: Ensuring ownership of ${APP_DIR} for ${PUID}:${PGID}..."
# Use chown -R for simplicity on the main app directory
chown -R "$PUID":"$PGID" "$APP_DIR"

# Optional: Change ownership of a separate data directory if used
# if [ -d "$DATA_DIR" ]; then
#   echo "Entrypoint: Ensuring ownership of ${DATA_DIR} for ${PUID}:${PGID}..."
#   chown -R "$PUID":"$PGID" "$DATA_DIR"
# fi

# Execute the command passed to the script (CMD) as the specified user/group
echo "Entrypoint: Switching to user ${APP_USER} (${PUID}:${PGID}) and executing command: $@"
exec su-exec "$PUID":"$PGID" "$@"
