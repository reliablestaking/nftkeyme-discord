package server

import (
	"net/http"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/reliablestaking/nftkeyme-discord/db"
	"github.com/reliablestaking/nftkeyme-discord/discord"
	"github.com/reliablestaking/nftkeyme-discord/nftkeyme"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type (
	// Server struct
	Server struct {
		Store               db.Store
		BuildTime           string
		Sha1ver             string
		DiscordAuthCodeURL  string
		DiscordOauthConfig  *oauth2.Config
		NftkeymeOauthConfig *oauth2.Config
		DiscordClient       discord.Client
		NftkeymeClient      nftkeyme.NftkeymeClient
		DiscordSession      *discordgo.Session
	}

	// Version struct
	Version struct {
		Sha       string `json:"sha"`
		BuildTime string `json:"buildTime"`
	}
)

// Start the server
func (s Server) Start() {
	e := echo.New()

	allowedOriginsCsv := make([]string, 0)
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins != "" {
		allowedOriginsCsv = strings.Split(allowedOrigins, ",")
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     allowedOriginsCsv,
		AllowMethods:     []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowCredentials: true,
	}))

	// asset / stake key endpoint
	e.GET("/init", s.InitFlow)
	e.GET("/discord", s.HandleDiscordAuthCode)
	e.GET("/nftkeyme", s.HandleNftkeymeAuthCode)

	// version endpoint
	e.GET("/version", s.GetVersion)

	port := os.Getenv("NFTKEYME_SERVICE_PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}

// GetVersion return build version info
func (s Server) GetVersion(c echo.Context) (err error) {
	version := Version{
		Sha:       s.Sha1ver,
		BuildTime: s.BuildTime,
	}

	return c.JSON(http.StatusOK, version)
}

// InitFlow initialize the flow
func (s Server) InitFlow(c echo.Context) (err error) {
	// redirect to discord auth flow
	return c.Redirect(302, s.DiscordAuthCodeURL)
}

// HandleDiscordAuthCode handle redirect
func (s Server) HandleDiscordAuthCode(c echo.Context) (err error) {
	authCode := c.QueryParam("code")
	logrus.Infof("Got auth code from discord %s", authCode)

	//exchange code for token
	token, err := s.DiscordOauthConfig.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		logrus.WithError(err).Error("Error exchange code for token")
		return c.JSON(http.StatusInternalServerError, nil)
	}
	logrus.Infof("Got token %s", token.AccessToken)

	// lookup user info
	userInfo, err := s.DiscordClient.GetUserInfo(token.AccessToken)
	if err != nil {
		logrus.WithError(err).Error("Error getting user info")
		return c.JSON(http.StatusInternalServerError, nil)
	}

	logrus.Infof("Got user with id %s and email %s", userInfo.ID, userInfo.Email)

	//TODO: persist this

	//redirect to nftkey me for now... use state of id
	url := s.NftkeymeOauthConfig.AuthCodeURL(userInfo.ID)

	return c.Redirect(302, url)
}

// HandleNftkeymeAuthCode handle redirect
func (s Server) HandleNftkeymeAuthCode(c echo.Context) (err error) {
	authCode := c.QueryParam("code")
	state := c.QueryParam("state")
	logrus.Infof("Got auth code from nftkeyme %s and state/discordid %s", authCode, state)

	//exchange code for token
	token, err := s.NftkeymeOauthConfig.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		logrus.WithError(err).Error("Error exchange code for token")
		return c.JSON(http.StatusInternalServerError, nil)
	}
	logrus.Infof("Got token %s", token.AccessToken)

	//TODO: persist this?

	// get assets
	assets, err := s.NftkeymeClient.GetAssetsForUser(token.AccessToken)
	if err != nil {
		logrus.WithError(err).Error("Error getting assets")
		return c.JSON(http.StatusInternalServerError, nil)
	}

	logrus.Infof("Found %d assets", len(assets))
	//check for policy id
	hasPolicy := hasPolicyID(assets)
	if hasPolicy {
		// grant access to channel
		logrus.Infof("Adding user %s to role", state)
		err = s.DiscordSession.GuildMemberRoleAdd("882816414652710912", state, "882817068112707626")
		if err != nil {
			logrus.WithError(err).Error("Error adding user to role")
			return c.JSON(http.StatusInternalServerError, nil)
		}
	} else {
		// tell user they don't have access
	}

	return c.JSON(http.StatusOK, nil)
}

func hasPolicyID(assets []nftkeyme.Asset) bool {
	for _, asset := range assets {
		if asset.PolicyId == "c6443f0c069487e1a8afbc0c8a3ac00fd26aee56f2f2f5d2bee12be4" {
			return true
		}
	}

	return false
}
