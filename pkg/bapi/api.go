package bapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/opentracing/opentracing-go"
)

type API struct {
	URL    string
	Client *http.Client
}

type AccessToken struct {
	Token string `json:"token"`
}

const (
	APP_KEY    = "eddycjy"
	APP_SECRET = "go-programming-tour-book"
)

func NewAPI(url string) *API {
	return &API{URL: url, Client: http.DefaultClient}
}

func (a *API) GetTagList(ctx context.Context, name string) ([]byte, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	body, err := a.httpGet(ctx, fmt.Sprintf("%s?token=%s&name=%s", "api/v1/tags", token, name))
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (a *API) httpGet(ctx context.Context, path string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", a.URL, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	span, _ := opentracing.StartSpanFromContext(
		ctx, "HTTP GET: "+a.URL,
		opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
	)
	span.SetTag("url", url)
	_ = opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	req = req.WithContext(context.Background())
	client := http.Client{} //Timeout: time.Second * 60
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	defer span.Finish()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (a *API) getAccessToken(ctx context.Context) (string, error) {
	body, err := a.httpGet(ctx, fmt.Sprintf("%s?app_key=%s&app_secret=%s", "auth", APP_KEY, APP_SECRET))
	if err != nil {
		return "", err
	}

	var accessToken AccessToken
	err = json.Unmarshal(body, &accessToken)
	if err != nil {
		return "", err
	}

	return accessToken.Token, nil
}
