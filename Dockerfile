FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /pc-waker .

FROM scratch
COPY --from=build /pc-waker /pc-waker
ENTRYPOINT ["/pc-waker"]
