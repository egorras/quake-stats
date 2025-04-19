FROM golang:1.24.2-bullseye

# Install required dependencies
RUN apt-get update && apt-get install -y \
    libzmq3-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy and build the application
COPY . .
RUN go mod tidy
RUN go build -o collector

# Run the application
CMD ["./collector"] 