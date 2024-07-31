package functions

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dyatlov/go-opengraph/opengraph"
	"go.uber.org/zap"
	log "skhaz.dev/urlshortnen/logging"
)

func facebook(url string, wg *sync.WaitGroup) {
	defer wg.Done()

	accessToken := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	url = fmt.Sprintf("https://graph.facebook.com/v12.0/?id=%s&scrape=true&access_token=%s", url, accessToken)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Error("error creating facebook request", zap.Error(err))
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("error sending facebook request", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("non-200 response from facebook", zap.Int("status", resp.StatusCode))
		return
	}

	log.Info("facebook complete", zap.String("url", url), zap.Int("statusCode", resp.StatusCode))
}

func twitter(url string, wg *sync.WaitGroup) {
	defer wg.Done()

	url = fmt.Sprintf("https://cards-dev.twitter.com/validator?url=%s", url)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Error("error accessing twitter validator", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("non-200 response from twitter validator", zap.Int("status", resp.StatusCode))
		return
	}

	log.Info("twitter complete", zap.String("url", url), zap.Int("statusCode", resp.StatusCode))
}

func linkedin(url string, wg *sync.WaitGroup) {
	defer wg.Done()

	url = fmt.Sprintf("https://www.linkedin.com/sharing/share-offsite/?url=%s", url)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Error("error sending linkedin request", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("non-200 response from linkedin request", zap.Int("status", resp.StatusCode))
		return
	}

	log.Info("linkedin complete", zap.String("url", url), zap.Int("statusCode", resp.StatusCode))
}

func cdn(url string, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Error("error sending takealook request", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("non-200 response from takealook request", zap.Int("status", resp.StatusCode))
		return
	}

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("error reading takealook response body", zap.Error(err))
		return
	}

	og := opengraph.NewOpenGraph()
	if err = og.ProcessHTML(strings.NewReader(string(buffer))); err != nil {
		log.Error("error parsing open-graph", zap.Error(err))
		return
	}

	if len(og.Images) == 0 {
		log.Error("no open-graph images found")
		return
	}

	client.Timeout = 60 * time.Second
	resp, err = client.Get(og.Images[0].URL)
	if err != nil {
		log.Error("error downloading open-graph image", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("non-200 response when downloading open-graph image", zap.Int("status", resp.StatusCode))
		return
	}

	log.Info("cdn complete", zap.String("url", url), zap.Int("statusCode", resp.StatusCode))
}

func warmup(url string) {
	var wg sync.WaitGroup
	wg.Add(4)

	go facebook(url, &wg)
	go twitter(url, &wg)
	go linkedin(url, &wg)
	go cdn(url, &wg)

	wg.Wait()

	log.Info("warmup has been completed")
}
