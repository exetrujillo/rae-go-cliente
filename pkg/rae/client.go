package rae

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const (
	BaseURL   = "https://dle.rae.es/data/"
	AuthToken = "Basic cDY4MkpnaFMzOmFHZlVkQ2lFNDM0"
	UserAgent = "Diccionario/2 CFNetwork/808.2.16 Darwin/16.3.0"
)

type Client struct {
	http tls_client.HttpClient
}

func NewClient() *Client {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Safari_IOS_16_0),
	}

	client, _ := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	return &Client{
		http: client,
	}
}

func (c *Client) SendRequest(endpoint string, withConjugations bool) ([]byte, error) {
	if strings.HasPrefix(endpoint, "/") {
		endpoint = endpoint[1:]
	}
	reqURL := BaseURL + endpoint
	
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Authorization", AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Procesar respuesta JSONP - eliminar envoltorio como json(...) o jsonp123(...)
	bodyStr := string(body)
	bodyStr = strings.TrimSpace(bodyStr)
	
	// Verificar si es un artículo HTML (necesita análisis)
	if strings.Contains(bodyStr, "<article id=") {
		// Analizar HTML a JSON
		parsed, err := ParseHTMLDefinitions(bodyStr, withConjugations)
		if err != nil {
			return nil, fmt.Errorf("error al analizar HTML: %v", err)
		}
		return parsed, nil
	}
	
	// Eliminar envoltorio json(...)
	if strings.HasPrefix(bodyStr, "json(") && strings.HasSuffix(bodyStr, ")") {
		bodyStr = bodyStr[5 : len(bodyStr)-1]
	}
	
	// Eliminar envoltorio jsonp123(...)
	if strings.HasPrefix(bodyStr, "jsonp123(") && strings.HasSuffix(bodyStr, ")") {
		bodyStr = bodyStr[9 : len(bodyStr)-1]
	}

	// Limpiar entidades HTML y superíndices
	bodyStr = strings.ReplaceAll(bodyStr, "&#xE1;", "á")
	bodyStr = strings.ReplaceAll(bodyStr, "&#xE9;", "é")
	bodyStr = strings.ReplaceAll(bodyStr, "&#xED;", "í")
	bodyStr = strings.ReplaceAll(bodyStr, "&#xF3;", "ó")
	bodyStr = strings.ReplaceAll(bodyStr, "&#xFA;", "ú")
	bodyStr = strings.ReplaceAll(bodyStr, "&#xF1;", "ñ")
	bodyStr = strings.ReplaceAll(bodyStr, "&#x2016;", "||")
	
	// Eliminar patrones <sup>N</sup>
	re := regexp.MustCompile(`<sup>\d+</sup>`)
	bodyStr = re.ReplaceAllString(bodyStr, "")
	re = regexp.MustCompile(`<sup>\d+\\/sup>`)
	bodyStr = re.ReplaceAllString(bodyStr, "")

	return []byte(bodyStr), nil
}

// FetchRaw returns the raw body without parsing (for debugging)
func (c *Client) FetchRaw(endpoint string) ([]byte, error) {
	reqURL := BaseURL + endpoint
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Authorization", AuthToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) GetWordOfTheDay() ([]byte, error) {
	return c.SendRequest("wotd?callback=json", false)
}

func (c *Client) GetRandomWord() ([]byte, error) {
	return c.SendRequest("random", false)
}

func (c *Client) SearchWord(query string) ([]byte, error) {
	return c.SendRequest("search?w="+url.QueryEscape(query), false)
}

func (c *Client) FetchWord(id string, withConjugations bool) ([]byte, error) {
	return c.SendRequest("fetch?id="+url.QueryEscape(id), withConjugations)
}

func (c *Client) KeyQuery(query string) ([]byte, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("callback", "jsonp123")
	return c.SendRequest("keys?"+params.Encode(), false)
}

func (c *Client) SearchAnagram(word string) ([]byte, error) {
	return c.SendRequest("anagram?w="+url.QueryEscape(word), false)
}
