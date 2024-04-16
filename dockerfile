# Use the official golang image as a base
FROM golang:alpine

# Set the working directory inside the container
WORKDIR /go/src/app

# Copy the Go source code into the container
COPY . .

# Build the Go application
RUN go build -o app .

# Set the entry point of the container to the Go application
CMD ["./app"]

