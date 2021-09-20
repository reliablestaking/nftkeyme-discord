FROM alpine

COPY nftkeyme-discord /usr/local/bin/
EXPOSE 8080

ENTRYPOINT ["nftkeyme-discord"]
