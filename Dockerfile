# ==========================================
# Secure Runtime Sandbox Environment
# ==========================================
FROM alpine:latest

# Add core security certificates to allow your scripts to access external HTTPS APIs safely
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the local pre-compiled binary straight into the sandbox tracking space
COPY ./automation-engine .

# Ensure the operating system layer registers the binary as an executable file
RUN chmod +x ./automation-engine

# Initialize clean storage directories inside the container for local logs/DB caches
RUN mkdir -p logs downloads database

# Instruct the container to execute your compiled automation engine automatically on boot
CMD ["./automation-engine"]