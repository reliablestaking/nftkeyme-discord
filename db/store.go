package db

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type (
	// Store struct to store Db
	Store struct {
		Db *sqlx.DB
	}

	// DiscordUser struct to store
	DiscordUser struct {
		ID                   int            `db:"id"`
		DiscordUserID        string         `db:"discord_user_id"`
		NftkeymeAccessToken  sql.NullString `db:"nftkeyme_access_token"`
		NftkeymeRefreshToken sql.NullString `db:"nftkeyme_refresh_token"`
	}
)

// GetUserByDiscordID Gets a user using their discord id
func (s Store) GetUserByDiscordID(discordUserID string) (*DiscordUser, error) {
	discordUser := DiscordUser{}
	err := s.Db.Get(&discordUser, "SELECT * FROM discord_user where discord_user_id = $1", discordUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &discordUser, nil
}

// GetAllDiscordUsers Gets all users
func (s Store) GetAllDiscordUsers() ([]DiscordUser, error) {
	discordUsers := []DiscordUser{}
	err := s.Db.Select(&discordUsers, "SELECT * FROM discord_user")
	if err != nil {
		return nil, err
	}

	return discordUsers, nil
}

// InsertDiscordUser inserts a new user into the db
func (s Store) InsertDiscordUser(discordUserID, nftkeymeAccessToken, nftkeymeRefreshToken string) error {
	insertUserQuery := `INSERT INTO discord_user (discord_user_id,nftkeyme_access_token,nftkeyme_refresh_token) VALUES($1, $2, $3)`

	rows, err := s.Db.Query(insertUserQuery, discordUserID, nftkeymeAccessToken, nftkeymeRefreshToken)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}

// UpdateDiscordUser updates a new user in the db
func (s Store) UpdateDiscordUser(discordUserID, nftkeymeAccessToken, nftkeymeRefreshToken string) error {
	insertUserQuery := `UPDATE discord_user SET nftkeyme_access_token = $1, nftkeyme_refresh_token = $2 WHERE discord_user_id = $3`

	rows, err := s.Db.Query(insertUserQuery, nftkeymeAccessToken, nftkeymeRefreshToken, discordUserID)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}
