package custom

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/route"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func customRoutes(cfg *config.CustomBE, ch chan string) {

	var Routes *[]route.RouteDef
	var trans *http.Transport

	if cfg.CheckTLSSkipVerify {
		trans = &http.Transport{}

	} else {
		trans = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client := &http.Client{
		Transport: trans,
		Timeout:   cfg.Timeout * time.Second,
	}

	URL := fmt.Sprintf("%s://%s", cfg.Scheme, cfg.Host)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Printf("[ERROR] Can not generate new HTTP request")
	}
	req.Close = true

	fmt.Printf("custom config - %+v", cfg)

	for {

		resp, err := client.Do(req)
		if err != nil {
			ch <- fmt.Sprintf("Error Sending HTTPs Request To Custom BE - %s -%s", URL, err.Error())
			time.Sleep(cfg.PollingInterval * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			ch <- fmt.Sprintf("Error Non-200 return (%v) from  -%s", resp.StatusCode, URL)
			time.Sleep(cfg.PollingInterval * time.Second)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ch <- fmt.Sprintf("Error Can not read response from -%s,  %s", URL, err.Error())
			time.Sleep(cfg.PollingInterval * time.Second)
			continue
		}

		err = json.Unmarshal(body, Routes)
		if err != nil {
			ch <- fmt.Sprintf("Error Can not unmarshal response - %s,  %s", URL, err.Error())
			time.Sleep(cfg.PollingInterval * time.Second)
			continue
		}

		//TODO validate data

		t, err := route.NewTableCustomBE(Routes)
		if err != nil {
			ch <- err.Error()
		}
		route.SetTable(t)

		ch <- "OK"
		time.Sleep(cfg.PollingInterval * time.Second)

	}

}
