FROM asia.gcr.io/staging-host-12358/rapido-runtime-golang-alpine-3.11:1.14.2

COPY ./artifacts/silent-assassin /usr/app/
WORKDIR /usr/app/

CMD ["sh", "-c", " ./silent-assassin start"]