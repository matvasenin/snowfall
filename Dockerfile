# Build the dependency management image
FROM golang:alpine AS builder

# Enable the static linking
ENV CGO_ENABLED=0
# Configure the target platform
ENV GOOS=linux

# Install the tzdata
RUN apk update --no-cache && apk add --no-cache tzdata

# Setup the working directory
WORKDIR /app

# Add dependency management files
ADD go.mod .
ADD go.sum .

# Install the dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the app
RUN go build -o main main.go

# Build the production image
FROM alpine AS production

# Setup the working directory
WORKDIR /app

# Copy binary from the builder
COPY --from=builder /app/main .

# Run the server
CMD ["/app/main"]