package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type doFunc func(req *http.Request) (*http.Response, error)

type Client struct {
	ServerURL  string
	HttpClient *http.Client
	Debug      bool
	Logger     *log.Logger
	do         doFunc
}

func NewClient(url, caCertPath string, isDebug bool) (*Client, error) {
	if url == "" {
		return nil, errors.New("url is empty error")
	}

	var httpClient *http.Client

	if caCertPath == "" {

		httpClient = &http.Client{}

	} else {
		pool := x509.NewCertPool()

		caCrt, err := ioutil.ReadFile(caCertPath + "/ca.crt")
		if err != nil {
			return nil, errors.New("ReadFile (ca.crt) error")
		}
		pool.AppendCertsFromPEM(caCrt)

		//添加ca证书， 证书文件名必须一致
		cliCrt, err := tls.LoadX509KeyPair(caCertPath+"/client.crt", caCertPath+"/client.key")
		if err != nil {
			log.Println("Loadx509keypair err:", err)
			return nil, err
		} else {
			log.Println("Loadx509keypair Suceess.")
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      pool,
				Certificates: []tls.Certificate{cliCrt},
			},
		}

		httpClient = &http.Client{Transport: tr}
	}

	return &Client{
		ServerURL:  url,
		HttpClient: httpClient,
		Debug:      isDebug,
		Logger:     log.New(os.Stderr, "lmSdkClient", log.LstdFlags),
	}, nil
}

func (c *Client) debug(format string, v ...interface{}) {
	if c.Debug {
		c.Logger.Printf(format, v...)
	}
}

func (c *Client) parseRequest(r *request, bodyString string) (err error) {

	err = r.validate()
	if err != nil {
		return err
	}

	fullURL := fmt.Sprintf("%s%s", c.ServerURL, r.endpoint)
	header := http.Header{}
	body := &bytes.Buffer{}

	if bodyString != "" {
		body = bytes.NewBufferString(bodyString)
	}

	header.Set("Accept", "application/json; charset=utf-8")
	header.Set("Content-Type", "application/json; charset=utf-8")

	queryString := r.query.Encode()
	if queryString != "" {
		fullURL = fmt.Sprintf("%s?%s", fullURL, queryString)
	}

	c.debug("full url: %s, body: %s", fullURL, bodyString)

	r.fullURL = fullURL
	r.header = header
	r.body = body
	return nil
}

func (c *Client) callApi(ctx context.Context, r *request, bodyString string) (data []byte, err error) {
	err = c.parseRequest(r, bodyString)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(r.method, r.fullURL, r.body)
	if err != nil {
		return []byte{}, err
	}

	req = req.WithContext(ctx)
	req.Header = r.header
	c.debug("request: %#v", req)
	f := c.do
	if f == nil {
		f = c.HttpClient.Do
	}

	res, err := f(req)
	if err != nil {
		log.Println("f(req) error")
		return []byte{}, err
	}
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		cerr := res.Body.Close()

		if err == nil && cerr != nil {
			err = cerr
		}
	}()

	c.debug("response: %#v", res)
	c.debug("response body: %s", string(data))

	return data, nil
}


func (c *Client) NewAuthService() *AuthService {
	return &AuthService{c: c}
}
