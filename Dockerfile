# ---- Base Node ----
# Use a specific Node.js version on Alpine for smaller images
FROM node:22-alpine AS base
WORKDIR /app
# Copy package.json and lock file separately to leverage Docker cache
COPY package.json package-lock.json* ./

# ---- Dependencies ----
# Install only production dependencies first
FROM base AS prod-deps
RUN npm install --omit=dev

# ---- Build ----
# Install all dependencies (including dev) and build the app
FROM base AS builder
# Copy production node_modules from the previous stage
COPY --from=prod-deps /app/node_modules ./node_modules
# Install dev dependencies needed for build
RUN npm install
# Copy the rest of the application source code
COPY . .
# Run the SvelteKit build command
RUN npm run build

# ---- Runner ----
# Use a clean Node.js Alpine image for the final stage
FROM node:22-alpine AS runner
WORKDIR /app

# Install su-exec (needed by the entrypoint script) and bash (often useful for scripts)
RUN apk add --no-cache su-exec bash

# Set environment variables
ENV NODE_ENV=production
# Set HOST for adapter-node to listen on all interfaces within the container
ENV HOST=0.0.0.0
# Set PORT (optional, adapter-node defaults to 3000 if not set)
# ENV PORT=3000
# Default PUID/PGID (can be overridden in docker-compose.yml)
ENV PUID=1000
ENV PGID=1000

# Copy necessary artifacts from previous stages
# Copy production dependencies
COPY --from=prod-deps /app/node_modules ./node_modules
# Copy the built application output
COPY --from=builder /app/build ./build
# Copy package.json (needed by adapter-node runner)
COPY --from=builder /app/package.json ./package.json
# Copy the settings file if it exists at build time (optional, usually mounted as a volume)
# COPY --from=builder /app/app-settings.json ./app-settings.json

# Copy the entrypoint script and make it executable
COPY scripts/docker/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# NOTE: User creation and ownership changes are now handled by the entrypoint script
# We run as root initially, and the entrypoint script switches user

# Expose the port the application runs on (default for adapter-node is 3000)
EXPOSE 3000

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Command to start the Node.js server (passed as arguments to the entrypoint)
CMD ["node", "build"]