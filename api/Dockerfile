# Start from the latest golang base image
FROM golang:1.7.4

# Set the Current Working Directory inside the container
WORKDIR /go/src/app

# Copy go mod and sum files
COPY . .

# Download all dependencies
RUN go get -d -v ./...
RUN go install -v ./...

# Build the Go app
RUN go build -o api .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["./api"]