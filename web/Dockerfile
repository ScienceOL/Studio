FROM node:24-alpine AS base

# Dev stage: minimal image for hot-reload with volume mounts
FROM base AS dev
WORKDIR /app
RUN corepack enable

FROM base AS deps
WORKDIR /app

# Copy package manager files
COPY package.json yarn.lock* .yarnrc.yml* ./
COPY vite.config.ts ./
COPY tsconfig.json tsconfig.app.json tsconfig.node.json ./

# Install dependencies
RUN corepack enable

# Skip native addon compilation (fails under QEMU cross-platform emulation)
RUN YARN_ENABLE_SCRIPTS=0 yarn install --immutable

# Stage 1: Build the application
FROM base AS builder
WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
# Enable corepack to use the correct yarn version
RUN corepack enable

COPY . .

# Set build-time environment variables
ARG VITE_BACKEND_URL
ARG VITE_API_BASE_URL
ARG VITE_FRONTEND_BASE_URL

ENV NODE_ENV=production
ENV VITE_BACKEND_URL=${VITE_BACKEND_URL}
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_FRONTEND_BASE_URL=${VITE_FRONTEND_BASE_URL}

# Build the application
# The build script in package.json is "rimraf site && tsc -b && vite build"
RUN \
  if [ -f yarn.lock ]; then yarn run build; \
  else echo "Lockfile not found." && exit 1; \
  fi

# Stage 2: Serve the application using a lightweight Node.js server
FROM base AS runner
WORKDIR /app

# Copy only the built application assets from the build stage
COPY --from=builder /app/dist ./

# Expose the port the app will run on
EXPOSE 32234

# Start a production-ready server for SPAs using 'serve'
# npx will download and run 'serve' without needing it in package.json
# The '-s' flag handles SPA routing by redirecting all requests to index.html
CMD ["npx", "-y", "serve", "-s", ".", "-l", "32234"]
