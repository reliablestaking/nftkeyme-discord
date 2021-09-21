FROM alpine

COPY nftkeyme-discord /usr/local/bin/

ENTRYPOINT ["nftkeyme-discord"]