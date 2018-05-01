package torrentapi

// API documentation
// http://torrentapi.org/apidocs_v2.txt

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/bsm/ratelimit.v1"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const DefaultEndpoint = "https://torrentapi.org/pubapi_v2.php"

var (
	ErrUnexpectedContent  = errors.New("torrentapi: unexpected content")
	ErrNetworkRequest     = errors.New("torrentapi: remote server error")
	ErrApiRate            = errors.New("torrentapi: too many requests per second")
	ErrApiToken           = errors.New("torrentapi: token expired")
	ErrApiImdb            = errors.New("torrentapi: cant find imdb in database")
	ErrApiNoResult        = errors.New("torrentapi: no results found")
	ErrApiInvalidCategory = errors.New("torrentapi: invalid category")
	ErrApiInvalidSort     = errors.New("torrentapi: invalid sort")
)

type TorrentResults struct {
	Torrents []Torrent `json:"torrent_results"`
	ApiError
}

type Torrent struct {
	Title       string      `json:"title"`
	Size        int64       `json:"size"`
	Seeders     int64       `json:"seeders"`
	Leechers    int64       `json:"leechers"`
	PubDate     string      `json:"pubdate"`
	Category    string      `json:"category"`
	Ranked      int         `json:"ranked"`
	MagnetURL   string      `json:"download"`
	EpisodeInfo EpisodeInfo `json:"episode_info"`
}

type EpisodeInfo struct {
	AirDate    string `json:"airdate"`
	ShowTvdbID string `json:"tvdb"`
	ShowImdbID string `json:"imdb"`
	ShowTmdbID string `json:"themoviedb"`
	Season     string `json:seasonnum`
	Episode    string `json:epnum`
}

type ApiError struct {
	Msg  string `json:"error"`
	Code int    `json:"error_code"`
}

func (e *ApiError) Convert() error {
	switch e.Code {
	case 10:
		return ErrApiImdb
	case 30:
		return ErrApiInvalidCategory
	case 18:
		return ErrApiInvalidSort
	case 20:
		return ErrApiNoResult
	case 5:
		return ErrApiRate
	case 4:
		return ErrApiToken
	default:
		return e
	}
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("torrentapi: error: %d msg: %q", e.Code, e.Msg)
}

// Client represents the kickass client
type Client struct {
	Endpoint   *url.URL
	HTTPClient *http.Client
	Token      string
	rl         *ratelimit.RateLimiter
	AppId      int
}

// New creates a new client
func New(AppId int) (*Client, error) {
	endPoint, err := url.Parse(DefaultEndpoint)
	if err != nil {
		return nil, err
	}
	rl := ratelimit.New(1, time.Duration(2)*time.Second)
	c := &Client{
		Endpoint:   endPoint,
		HTTPClient: http.DefaultClient,
		rl:         rl,
		AppId:      AppId,
	}
	return c, nil
}

func (c *Client) Init() error {
	return c.GetToken()
}

func (c *Client) List(query map[string]string) (r TorrentResults, err error) {
	query["mode"] = "list"

	return c.Search(query)
}

func (c *Client) Search(query map[string]string) (r TorrentResults, err error) {
	var baseUrl url.URL
	baseUrl = *c.Endpoint // Copy the URL struct into a local one

	params := url.Values{}
	params.Add("token", c.Token)
	params.Add("app_id", fmt.Sprintf("%d", c.AppId))

	if _, ok := query["mode"]; !ok {
		params.Add("mode", "search")
	}
	if _, ok := query["format"]; !ok {
		params.Add("format", "json_extended")
	}

	for k, v := range query {
		params.Add(k, v)
	}
	baseUrl.RawQuery = params.Encode()

	if c.rl.Limit() {
		err = ErrApiRate
		return
	}

	var resp *http.Response
	resp, err = c.HTTPClient.Get(baseUrl.String())
	if resp.StatusCode == 429 {
		err = ErrApiRate
		return
	}

	fmt.Println(resp.StatusCode)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)

	fmt.Println(string(b))
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &r)
	if err != nil {
		return
	}
	if r.Code != 0 {
		err = r.ApiError.Convert()
		return
	}

	return
}

func (c *Client) GetToken() (err error) {
	var baseUrl url.URL
	baseUrl = *c.Endpoint // Copy the URL struct into a local one

	params := url.Values{}
	params.Add("app_id", fmt.Sprintf("%d", c.AppId))
	params.Add("get_token", "get_token")
	baseUrl.RawQuery = params.Encode()

	var resp *http.Response
	resp, err = c.HTTPClient.Get(baseUrl.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	r := struct {
		Token string `json:"token"`
		ApiError
	}{}
	err = json.Unmarshal(b, &r)
	if err != nil {
		return err
	}
	if r.Code != 0 {
		return r.ApiError.Convert()
	}
	c.Token = r.Token
	return nil
}
