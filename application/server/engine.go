package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/chitoku-k/slack-to-ssh/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type engine struct {
	Port               string
	CertFile           string
	KeyFile            string
	SlackAppSecret     string
	ActionService      service.ActionService
	InteractionService service.InteractionService
}

type Engine interface {
	Start(ctx context.Context) error
}

func NewEngine(
	port string,
	certFile string,
	keyFile string,
	slackAppSecret string,
	action service.ActionService,
	interaction service.InteractionService,
) Engine {
	return &engine{
		Port:               port,
		CertFile:           certFile,
		KeyFile:            keyFile,
		SlackAppSecret:     slackAppSecret,
		ActionService:      action,
		InteractionService: interaction,
	}
}

func (e *engine) Start(ctx context.Context) error {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz"},
	}))

	router.Any("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.POST("/", func(c *gin.Context) {
		e.HandleSlackEvent(ctx, c)
	})

	server := http.Server{
		Addr:    net.JoinHostPort("", e.Port),
		Handler: router,
	}

	var eg errgroup.Group
	eg.Go(func() error {
		<-ctx.Done()
		return server.Shutdown(context.Background())
	})

	var err error
	if e.CertFile != "" && e.KeyFile != "" {
		server.TLSConfig = &tls.Config{
			GetCertificate: e.getCertificate,
		}
		err = server.ListenAndServeTLS("", "")
	} else {
		err = server.ListenAndServe()
	}

	if err == http.ErrServerClosed {
		return eg.Wait()
	}

	return err
}

func (e *engine) getCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(e.CertFile, e.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return &cert, nil
}
