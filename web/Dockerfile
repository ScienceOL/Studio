FROM node:23-alpine AS base

FROM base AS deps
WORKDIR /app

# Enable corepack first (as root)
RUN corepack enable

# Copy package manager files
COPY package.json yarn.lock* .yarnrc.yml* ./
COPY tsconfig.json ./

# Install dependencies
RUN yarn install --frozen-lockfile

# Stage 1: Build the application
FROM base AS builder
WORKDIR /app

# Enable corepack to use the correct yarn version
RUN corepack enable

COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Build the application
RUN \
  if [ -f yarn.lock ]; then yarn run build; \
  else echo "Lockfile not found." && exit 1; \
  fi

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV TZ=Asia/Shanghai
ENV NODE_ENV=production
# Uncomment the following line in case you want to disable telemetry during runtime.
ENV NEXT_TELEMETRY_DISABLED=1
RUN apk update && apk add --no-cache tzdata

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public

# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 32234

ENV PORT=32234

# server.js is created by next build from the standalone output
# https://nextjs.org/docs/pages/api-reference/next-config-js/output
ENV HOSTNAME="0.0.0.0"
CMD ["node", "server.js"]
