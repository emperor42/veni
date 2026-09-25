FROM golang:1.21-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/veni-server .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/veni-server /app/veni-server
COPY templates ./templates
COPY static ./static

# The process binds all interfaces inside the container so a published port
# works. Startup refuses this non-loopback bind unless VENI_API_TOKEN is set;
# publish the port only to a protected network and provide that token.
ENV VENI_HOST=0.0.0.0 \
    VENI_PORT=8087
EXPOSE 8087

CMD ["/app/veni-server"]
