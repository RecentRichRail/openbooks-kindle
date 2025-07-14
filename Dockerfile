# Build frontend
FROM node:16 as web
WORKDIR /web
COPY server/app/package*.json ./
RUN npm ci --only=production --silent
COPY server/app/ ./
RUN npm run build

# Build backend
FROM golang:1.21 as build
WORKDIR /go/src/
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy source code (excluding problematic files via .dockerignore)
COPY . .
# Copy built frontend
COPY --from=web /web/dist/ ./server/app/dist/

ENV CGO_ENABLED=0
WORKDIR /go/src/cmd/openbooks/
RUN go build -ldflags="-w -s" -o openbooks .

# Runtime stage
FROM gcr.io/distroless/static-debian11 as app
WORKDIR /app

# Copy the binary
COPY --from=build /go/src/cmd/openbooks/openbooks .

# Create directories for volumes
USER root
RUN mkdir -p /books /app/logs && chown -R 1000:1000 /app /books
USER 1000:1000

EXPOSE 80
VOLUME ["/books"]
ENV BASE_PATH=/

ENTRYPOINT ["./openbooks", "server", "--dir", "/books", "--port", "80"]
