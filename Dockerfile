FROM golang:1.11 as builder

ENV APP_NAME auth
ENV WORKDIR ${GOPATH}/src/github.com/ViBiOh/auth

WORKDIR ${WORKDIR}
COPY ./ ${WORKDIR}/

RUN make ${APP_NAME} \
 && mkdir -p /app \
 && curl -s -o /app/cacert.pem https://curl.haxx.se/ca/cacert.pem \
 && cp bin/${APP_NAME} /app/

FROM scratch

ENV APP_NAME auth
HEALTHCHECK --retries=10 CMD [ "/auth", "-url", "https://localhost:1080/health" ]

EXPOSE 1080
ENTRYPOINT [ "/auth" ]

COPY --from=builder /app/cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/${APP_NAME} /auth
