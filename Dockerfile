FROM scratch

HEALTHCHECK --retries=10 CMD [ "/auth", "-url", "https://localhost:1080/health" ]

EXPOSE 1080
ENTRYPOINT [ "/auth" ]

COPY cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY bin/auth /auth
