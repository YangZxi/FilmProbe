FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

# build
# RUN go mod init FilmProbe
RUN go build -o film-probe


FROM alpine:latest

WORKDIR /app/

# copy all from builder
COPY --from=builder /app/film-probe .
COPY --from=builder /app/index.html .

EXPOSE 8080

CMD ["./film-probe"]
