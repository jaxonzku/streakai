# Use the official Golang image as a base image
FROM golang:1.21 AS build

# Set the Current Working Directory inside the container
WORKDIR /streakai-app

# Copy go mod and sum files
COPY streakai-app/go.mod streakai-app/go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY streakai-app/ .

# Build the Go app
RUN go build -o streakai-app .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./streakai-app"]