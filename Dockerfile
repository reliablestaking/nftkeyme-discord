FROM alpine

COPY nftkeyme-discord /usr/local/bin/nftkeyme-discord

ENTRYPOINT ["nftkeyme-discord"]
