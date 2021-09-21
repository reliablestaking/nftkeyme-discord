### Overview 

This application provides an example integration between https://nftkey.me and Discord. It grants access to a role in discord based on a user having an NFT from a specified policy id. 

An overview of the flow in the app...

1. User clicks get started on main page
1. This directs users to /init which in turn directs user to discord oauth auth url (redirect url in discord should be /discord in this app)
1. User login / consents in discord
1. User is redirected back to /discord, 
   1. auth code is exchanged for access token
   1. discord user id is queried using access token
   1. user is redirected to NFT Key auth code url
1. User login / consents in NFT Key 
1. User is directed back to /nftkeyme
   1. auth code is exchanged for access token
   1. access and refresh tokens are persisted to db
   1. NFTs/Assets are queried from NFT Key using access token
   1. discord user is added to role if policy check is successfull
1. There is a periodic check to ensure user doesn't move NFT out of their wallet

### Env Vars To Run

```
export DISCORD_URL=https://discordapp.com/api
export DISCORD_AUTH_URL=
export DISCORD_BOT_TOKEN=
export DISCORD_CLIENT_ID=
export DISCORD_CLIENT_SECRET=
export DISCORD_TOKEN_URL="https://discord.com/api/oauth2/token"
export DISCORD_REDIRECT_URL=http://localhost:8080/discord

export NFTKEYME_URL=https://service.nftkey.me/service/api
export NFTKEYME_CLIENT_ID=
export NFTKEYME_CLIENT_SECRET=
export NFTKEYME_TOKEN_URL="https://service.nftkey.me/oauth/oauth2/token"
export NFTKEYME_AUTH_URL="https://service.nftkey.me/oauth/oauth2/auth"
export NFTKEYME_REDIRECT_URL=http://localhost:8080/nftkeyme

export DISCORD_SERVER_ID=
export DISCORD_CHANNEL_ID=

export POLICY_ID_CHECK=

export DB_ADDR=127.0.0.1
export DB_PORT=5432
export DB_USER=nftkeyme_discord_user
export DB_PASS=nftkeyme_discord_password
export DB_NAME=nftkeyme_discord
export DB_SSL=true
```

## Service TODO items

1. Database migration scripts

## UI TODO items

1. background image scaling (apply whatever fix we do in the other app)
1. favicon
