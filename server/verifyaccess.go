package server

import (
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// VerifyAccess rechecks that users are allowed access
func (s Server) VerifyAccess() {
	for true {
		logrus.Info("Verifying access...")
		discordUsers, err := s.Store.GetAllDiscordUsers()
		if err != nil {
			logrus.WithError(err).Error("Error getting all users")
		}

		for _, discordUser := range discordUsers {
			logrus.Infof("Verifying access for user %s", discordUser.DiscordUserID)
			t := oauth2.Token{
				RefreshToken: discordUser.NftkeymeRefreshToken.String,
			}

			tokenSource := s.NftkeymeOauthConfig.TokenSource(oauth2.NoContext, &t)
			newToken, err := tokenSource.Token()
			if err != nil {
				logrus.WithError(err).Error("Error getting token")
				continue
			}

			logrus.Info(newToken.AccessToken)
			logrus.Info(t.AccessToken)
			if newToken.AccessToken != discordUser.NftkeymeAccessToken.String {
				logrus.Infof("Updating discord user %s with new token", discordUser.DiscordUserID)
				err = s.Store.UpdateDiscordUser(discordUser.DiscordUserID, newToken.AccessToken, newToken.RefreshToken)
				if err != nil {
					logrus.WithError(err).Error("Error updating discord user")
					continue
				}
			}

			assets, err := s.NftkeymeClient.GetAssetsForUser(newToken.AccessToken)
			if err != nil {
				logrus.WithError(err).Error("Error getting assets")
				continue
			}

			//check for policy id
			hasPolicy := s.hasPolicyID(assets)
			if !hasPolicy {
				// removing access to channel
				logrus.Infof("Removing user %s from role", discordUser.DiscordUserID)
				err = s.DiscordSession.GuildMemberRoleRemove(s.DiscordServerID, discordUser.DiscordUserID, s.DiscordChannelID)
				if err != nil {
					logrus.WithError(err).Error("Error removing user from role")
					continue
				}
			}
		}

		time.Sleep(12 * time.Hour)
	}
}
