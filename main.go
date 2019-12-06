package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env/v6"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, _ = fmt.Fprint(w, "Welcome to the Proxy Transform UptimeRobot GET parameters -> Discord POST parameters!\n")
}

func ProxyParameters(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var postParams proxyPostParams
	var err error
	var decoder *json.Decoder
	var params url.Values
	var finalJsonParamsMap = map[string]string{}
	var jsonBytes []byte
	var req *http.Request
	var resp *http.Response
	var respBytes []byte

	decoder = json.NewDecoder(r.Body)

	// Populate POST parameter token into struct
	if err = decoder.Decode(&postParams); err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("{\"success\": false,\"error\": \"unknown error\"}"))
		log.Print(err.Error())
		return
	}

	// Check if token is good (to verify that it's uptime-robot calling us)
	if postParams.Token != config.SecurityToken {
		w.WriteHeader(401)
		_, _ = w.Write([]byte("{\"success\": false,\"error\": \"wrong token\"}"))
		return
	}

	params = r.URL.Query()

	for param, value := range params {
		finalJsonParamsMap[param] = value[0]
	}

	// Prepare the POST parameter for discord.
	jsonBytes = []byte(
		fmt.Sprintf(
			"{\"content\":\"@everyone !\\nID [%d]\\nUrl [%s]\\nAlert Name [%s] Status is %s!\\nDetails : %s\\nDuration in seconds : %d\\n\"}",
			GetAtoiValue(finalJsonParamsMap["monitorID"]),
			finalJsonParamsMap["monitorURL"],
			finalJsonParamsMap["monitorFriendlyName"],
			finalJsonParamsMap["alertTypeFriendlyName"],
			finalJsonParamsMap["alertDetails"],
			GetAtoiValue(finalJsonParamsMap["alertDuration"]),
		))

	log.Printf("Built content is %s", string(jsonBytes))

	if req, err = http.NewRequest("POST", config.DiscordWebhookUrl, bytes.NewBuffer(jsonBytes)); err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("{\"success\": false,\"error\": \"unknown error\"}"))
		log.Print(err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")

	if resp, err = client.Do(req); err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("{\"success\": false,\"error\": \"unknown error\"}"))
		log.Print(err.Error())
		return
	}

	if respBytes, err = ioutil.ReadAll(resp.Body); err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("{\"success\": false,\"error\": \"unknown error\"}"))
		log.Print(err.Error())
		return
	}

	log.Printf("Discord returned [%s] status code [%d]", string(respBytes), resp.StatusCode)
}

func GetAtoiValue(param string) int {
	var result int

	result, _ = strconv.Atoi(param)

	return result
}

func main() {
	var router *httprouter.Router
	var err error

	if err = godotenv.Load(); err != nil {
		fmt.Printf("%+v\n", err)
	}

	if err = env.Parse(&config); err != nil {
		fmt.Printf("%+v\n", err)
	}

	router = httprouter.New()
	router.GET("/", Index)
	router.POST("/proxy", ProxyParameters)

	fmt.Printf("Listening on port %d...\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router))
}
