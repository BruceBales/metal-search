FROM golang:1.22

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main ./random-band/cmd

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["./main"] 