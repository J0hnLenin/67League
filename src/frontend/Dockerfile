# Build stage
FROM node:24-alpine AS builder

ARG API_PROXY_TARGET
ARG ADMIN_PROXY_TARGET
ENV API_PROXY_TARGET=${API_PROXY_TARGET}
ENV ADMIN_PROXY_TARGET=${ADMIN_PROXY_TARGET}

WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Final stage
FROM node:24-alpine

WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json
# public директория опциональна — создаём пустую если нет
RUN mkdir -p /app/public

EXPOSE 3000

CMD ["npm", "start"]
