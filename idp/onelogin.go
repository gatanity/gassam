package idp

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"gatanity/assam/aws"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
)

const (
	oneloginSubdomainBaseUrlTemplate = "https://%s.onelogin.com"
	oneloginPortalUrlTemplate        = "https://%s.onelogin.com/portal"
	oneloginLoginUrlTemplate         = "https://%s.onelogin.com/trust/saml2/http-post/sso/%s"
)

type Onelogin struct {
	subdomain string
	appId     string
	msgChan   chan *network.EventRequestWillBeSent
}

func NewOnelogin(subdomain string, appId string) Onelogin {
	return Onelogin{
		subdomain: subdomain,
		appId:     appId,
		msgChan:   make(chan *network.EventRequestWillBeSent),
	}
}

func (o *Onelogin) Authenticate(ctx context.Context, userDataDir string) (string, error) {
	ctx, cancel := o.setupContext(ctx, userDataDir)
	defer cancel()

	// Need network.Enable() to handle network events.
	err := chromedp.Run(ctx, network.Enable())
	if err != nil {
		return "", err
	}
	o.listenNetworkRequest(ctx)

	err = o.establishOneloginSession(ctx)
	if err != nil {
		return "", err
	}

	err = o.navigateToLoginURL(ctx)
	if err != nil {
		return "", err
	}

	response, err := o.fetchSAMLResponse(ctx)
	if err != nil {
		return "", err
	}

	// Shut down gracefully to ensure that user data is stored.
	err = chromedp.Cancel(ctx)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (o *Onelogin) setupContext(ctx context.Context, userDataDir string) (context.Context, context.CancelFunc) {
	expandedDir := os.ExpandEnv(userDataDir)

	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserDataDir(expandedDir),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	}

	allocContext, _ := chromedp.NewExecAllocator(context.Background(), opts...)

	return chromedp.NewContext(allocContext)
}

func (o *Onelogin) listenNetworkRequest(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(v interface{}) {
		go func() {
			if req, ok := v.(*network.EventRequestWillBeSent); ok {
				o.msgChan <- req
			}
		}()
	})
}

func (o *Onelogin) establishOneloginSession(ctx context.Context) error {
	mainUrl := fmt.Sprintf(oneloginSubdomainBaseUrlTemplate, o.subdomain)
	if err := chromedp.Run(ctx, chromedp.Navigate(mainUrl)); err != nil {
		return err
	}

	for {
		var req *network.EventRequestWillBeSent
		select {
		case <-ctx.Done():
			return ctx.Err()
		case req = <-o.msgChan:
		}

		if req.Request.URL != fmt.Sprintf(oneloginPortalUrlTemplate, o.subdomain) {
			continue
		}

		return nil
	}
}

func (o *Onelogin) navigateToLoginURL(ctx context.Context) error {
	loginURL := fmt.Sprintf(oneloginLoginUrlTemplate, o.subdomain, o.appId)
	return chromedp.Run(ctx, chromedp.Navigate(loginURL))
}

func (o *Onelogin) fetchSAMLResponse(ctx context.Context) (string, error) {
	for {
		var req *network.EventRequestWillBeSent
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case req = <-o.msgChan:
		}

		if req.Request.URL != aws.EndpointURL {
			continue
		}

		form, err := url.ParseQuery(req.Request.PostData)
		if err != nil {
			return "", err
		}

		samlResponse, ok := form["SAMLResponse"]
		if !ok {
			return "", errors.New("no such key: SAMLResponse")
		}

		return samlResponse[0], nil
	}
}
