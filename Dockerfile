# ==========================================
# PHASE 1: Compile the Go Project Modules
# ==========================================
FROM golang:1.26-alpine AS builder

# Install build dependencies for C-Go libraries (Required if using raw sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Cache module layers first to speed up future container builds
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire workspace code base tree into the sandbox
COPY . .

# Compile the specific client runner into a single standalone binary
# Change 'cmd/client_a/main.go' if the path matches a different directory setup
RUN CGO_ENABLED=1 GO111MODULE=on GOOS=linux go build -o /automation-engine cmd/client_a/main.go

# ==========================================
# PHASE 2: Secure Runtime Sandbox Environment
# ==========================================
FROM alpine:latest  

# Add core security certificates to allow scripts to access external HTTPS APIs safely
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy only the compiled production-ready binary over from Phase 1
COPY --from=builder /automation-engine .

# Initialize clean storage directories inside the container for local logs/DB caches
RUN mkdir -p logs downloads localDatabase

# Instruct the container to execute your compiled automation engine automatically on boot
CMD ["./automation-engine"]
