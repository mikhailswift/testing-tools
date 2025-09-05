FROM golang:1.25.1-alpine AS build

WORKDIR /usr/src/app
RUN mkdir bin/
COPY go.mod main.go ./
RUN go build -v -o bin/app main.go

FROM alpine
COPY --from=build /usr/src/app/bin/app /usr/local/bin/app
ENTRYPOINT [ "app" ]
CMD [ "listen", "0.0.0.0:8080" ]
