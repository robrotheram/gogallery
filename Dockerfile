FROM node:lts-alpine as UI_BUILDER
ARG VER
WORKDIR /frontend
ADD /frontend .
RUN npm i; npm run build; 

FROM golang:1.21.4 as GO_BUILDER
ARG VER
WORKDIR /server
ADD go.mod .
ADD go.sum .
ADD main.go main.go
ADD backend backend
ADD themes themes
COPY --from=UI_BUILDER /frontend/dist /server/frontend/dist
RUN CGO_ENABLED=1 GOOS=linux go build

FROM ubuntu
LABEL org.opencontainers.image.source="https://github.com/robrotheram/gogallery"
WORKDIR /app
COPY --from=GO_BUILDER /server/gogallery /app/gogallery
COPY config_sample.yml /app/config.yml
ENV GLLRY_SERVER_PORT ":80"
ENV GLLRY_GALLERY_BASEPATH "/app/pictures"
ENV GLLRY_GALLERY_THEME "default"
WORKDIR /app
ENTRYPOINT ["/app/gogallery", "--config", "/app/config.yml",  "serve"]