package server

import (
	"html/template"
	"io"
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
		PolicyIDCheck       string
		DiscordServerID     string
		DiscordChannelID    string
	}

	// Version struct
	Version struct {
		Sha       string `json:"sha"`
		BuildTime string `json:"buildTime"`
	}
)

// Template struct to store templates
type Template struct {
	templates *template.Template
}

// Render render a template
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Start the server
func (s Server) Start() {
	logrus.Info("Starting server...")
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

	// static CSS/images
	e.Static("/static", "assets")

	t := &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
	e.Renderer = t

	// start / end urls
	e.GET("/", s.RenderStart)
	e.GET("/end", s.RenderEnd)

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
	logrus.Infof("Handling auth code from discord")
	authCode := c.QueryParam("code")

	//exchange code for token
	token, err := s.DiscordOauthConfig.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		logrus.WithError(err).Error("Error exchange code for token")
		return c.JSON(http.StatusInternalServerError, nil)
	}

	// lookup user info
	userInfo, err := s.DiscordClient.GetUserInfo(token.AccessToken)
	if err != nil {
		logrus.WithError(err).Error("Error getting user info")
		return c.JSON(http.StatusInternalServerError, nil)
	}

	logrus.Infof("Got user with id %s and email %s", userInfo.ID, userInfo.Email)

	//redirect to nftkey me use state of discord user id
	url := s.NftkeymeOauthConfig.AuthCodeURL(userInfo.ID)

	return c.Redirect(302, url)
}

// HandleNftkeymeAuthCode handle redirect
func (s Server) HandleNftkeymeAuthCode(c echo.Context) (err error) {
	authCode := c.QueryParam("code")
	state := c.QueryParam("state")
	logrus.Infof("Handling auth code from nftkeyme with state/discord id %s", state)

	//exchange code for token
	token, err := s.NftkeymeOauthConfig.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		logrus.WithError(err).Error("Error exchange code for token")
		return c.JSON(http.StatusInternalServerError, nil)
	}

	// persist tokens
	logrus.Infof("Checking if user already exsists in db %s", state)
	discordUser, err := s.Store.GetUserByDiscordID(state)
	if err != nil {
		logrus.WithError(err).Errorf("Error getting discord user %s", state)
		return c.JSON(http.StatusInternalServerError, nil)
	}
	if discordUser == nil {
		logrus.Infof("Inserting discord user record %s", state)
		err = s.Store.InsertDiscordUser(state, token.AccessToken, token.RefreshToken)
		if err != nil {
			logrus.WithError(err).Errorf("Error persisting discord user %s", state)
			return c.JSON(http.StatusInternalServerError, nil)
		}
	} else {
		logrus.Infof("Updating discord user record %s", state)
		err = s.Store.UpdateDiscordUser(state, token.AccessToken, token.RefreshToken)
		if err != nil {
			logrus.WithError(err).Errorf("Error persisting discord user %s", state)
			return c.JSON(http.StatusInternalServerError, nil)
		}
	}

	// get assets
	assets, err := s.NftkeymeClient.GetAssetsForUser(token.AccessToken)
	if err != nil {
		logrus.WithError(err).Error("Error getting assets")
		return c.JSON(http.StatusInternalServerError, nil)
	}

	logrus.Infof("Found %d assets", len(assets))

	//check for policy id
	hasPolicy := s.hasPolicyID(assets)
	if hasPolicy {
		// grant access to channel
		logrus.Infof("Adding user %s to role", state)
		err = s.DiscordSession.GuildMemberRoleAdd(s.DiscordServerID, state, s.DiscordChannelID)
		if err != nil {
			logrus.WithError(err).Error("Error adding user to role")
			return c.JSON(http.StatusInternalServerError, nil)
		}
	} else {
		logrus.Errorf("No trybbles detected in user linked wallet %s", state)
		// tell user they don't have access
		s.RenderError("No trybbles detected in your linked wallets", c)
		return nil
	}

	return c.Redirect(302, "/end")
}

func (s Server) hasPolicyID(assets []nftkeyme.Asset) bool {
	for _, asset := range assets {
		if asset.PolicyId == s.PolicyIDCheck {
			return true
		}
	}

	return false
}

// RenderStart renders start page
func (s Server) RenderStart(c echo.Context) error {
	start := struct {
		Title       string
		Description string
		Link        string
	}{
		Title:       "Trybbles Discord",
		Description: "Gain access to Trybbles Club discord channel using NFT Key Me!",
		Link:        "/init",
	}
	err := c.Render(http.StatusOK, "start.html", start)
	if err != nil {
		logrus.WithError(err).Error("Error rendering start template")
	}
	return err
}

// RenderEnd renders end page
func (s Server) RenderEnd(c echo.Context) error {
	start := struct {
		Description string
		Link        string
	}{
		Description: "You can now access the Trybbles Club discord channel!",
		Link:        "",
	}
	err := c.Render(http.StatusOK, "end.html", start)
	if err != nil {
		logrus.WithError(err).Error("Error rendering start template")
	}
	return err
}

// RenderError renders an error page
func (s Server) RenderError(errorMsg string, c echo.Context) error {
	errorEnd := struct {
		Error string
	}{
		Error: errorMsg,
	}
	err := c.Render(http.StatusOK, "error.html", errorEnd)
	if err != nil {
		logrus.WithError(err).Error("Error rendering start template")
	}
	return err
}
