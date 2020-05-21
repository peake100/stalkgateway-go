# syntax = docker/dockerfile:1.0-experimental

# Use the alpine golang image as our builder
FROM golang:alpine AS builder

LABEL stage=builder
# Install any compiler-only dependencies
RUN apk add --no-cache gcc libc-dev
WORKDIR /workspace
# Copy all the source files
COPY . .
# Build the GO program
RUN CGO_ENABLED=0 GOOS=linux go build -a -o stalkforecaster

FROM alpine AS final

WORKDIR /
# Copy the compiled binary from builder
COPY --from=builder /workspace/stalkforecaster .

# Execute the program upon start
CMD [ "./stalkforecaster" ]

# Expose http & gRPC ports
EXPOSE 80
EXPOSE 50051
