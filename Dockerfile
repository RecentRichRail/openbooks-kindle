FROM node:16 as web
WORKDIR /web
COPY . .
WORKDIR /web/server/app/
RUN npm install
RUN npm run build

FROM golang:1.21 as build
WORKDIR /go/src/
COPY . .
COPY --from=web /web/ .

ENV CGO_ENABLED=0
RUN go mod download
RUN go mod verify
WORKDIR /go/src/cmd/openbooks/
RUN go build -o openbooks

FROM gcr.io/distroless/static-debian11 as app
WORKDIR /app
COPY --from=build /go/src/cmd/openbooks/openbooks .

# Create a non-root user for security
USER 1000:1000

EXPOSE 80
VOLUME [ "/books" ]
ENV BASE_PATH=/

ENTRYPOINT ["./openbooks", "server", "--dir", "/books", "--port", "80"]
