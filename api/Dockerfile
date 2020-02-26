# Golang base image
FROM golang:1.7.4

# Set the current working directory
WORKDIR /go/src/app

# Copy code file
COPY . .

# Download all dependencies
RUN go get -d -v ./...
RUN go install -v ./...

# Build the Go app
RUN go build -o api .

# Expose port
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["./api"]
