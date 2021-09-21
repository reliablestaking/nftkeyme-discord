create table discord_user (
    id                         serial PRIMARY KEY, 
    discord_user_id            varchar(64) not null,
    nftkeyme_access_token      varchar(128),
    nftkeyme_refresh_token     varchar(128),
    UNIQUE(discord_user_id)
);