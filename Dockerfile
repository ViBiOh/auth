FROM scratch

HEALTHCHECK --retries=10 CMD http://localhost:1080/health

EXPOSE 1080
ENTRYPOINT [ "/bin/sh" ]

COPY script/ca-certificates.crt /etc/ssl/certs/
COPY bin/auth /bin/sh
