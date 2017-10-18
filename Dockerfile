FROM scratch

HEALTHCHECK --retries=10 CMD https://localhost:1080/health

EXPOSE 1080
ENTRYPOINT [ "/bin/sh" ]

COPY cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY bin/auth /bin/sh
