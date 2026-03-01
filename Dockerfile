# Stage 1: Build the frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend ./
RUN npm run build

# Stage 2: Build the Go backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server main.go

# Stage 3: Runtime image
FROM alpine:latest
WORKDIR /app

# Install standard CA certificates
RUN apk --no-cache add ca-certificates tzdata

# Copy built backend binary
COPY --from=backend-builder /app/server .

# Copy built frontend dist to the location expected by the Go server
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Expose the application port
EXPOSE 3000

# Run the server
CMD ["./server"]
