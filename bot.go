package cornacchia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type (

	// Timestamp is a helper type to handle RFC3339 timestamps
	Timestamp time.Time

	// AlertNotification is a bunch of alerts
	AlertNotification struct {
		Version           string            `json:"version"`
		GroupKey          string            `json:"groupKey"`
		Status            string            `json:"status"` // resolved|firing
		Receiver          string            `json:"receiver"`
		GroupLabels       map[string]string `json:"groupLabels"`
		CommonLabels      map[string]string `json:"commonLabels"`
		CommonAnnotations map[string]string `json:"commonAnnotations"`
		ExternalURL       string            `json:"externalURL"`
		Alerts            []Alert           `json:"alerts"`
	}

	// Alert represent a single alert
	Alert struct {
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		StartsAt    Timestamp         `json:"startsAt,omitempty"`
		EndsAt      Timestamp         `json:"endsAt,omitempty"`
	}
)

// healthHandler is a simple health-check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok\n")
}

// alertHandler is the main handler for Alertmanager's alerts
func alertHandler(mattermostURL string) http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received alert\n")

		var alertData AlertNotification
		err := json.NewDecoder(r.Body).Decode(&alertData)
		if err != nil {
			log.Printf("Cannot decode the incoming payload: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Build a markdown table to display each alert details.
		tableBuf := &bytes.Buffer{}
		table := tablewriter.NewWriter(tableBuf)
		table.SetHeader([]string{"Instance", "Summary"})
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")

		var text bytes.Buffer
		if alertData.Status == "firing" {
			text.WriteString(":warning:\n")
		} else {
			text.WriteString(":white_check_mark:\n")
		}

		if summary, ok := alertData.CommonAnnotations["summary"]; ok {
			text.WriteString(fmt.Sprintf("# %s\n", summary))
		} else {
			text.WriteString("# Unknown summary\n")
		}

		for _, alert := range alertData.Alerts {
			var instance, summary string
			var ok bool
			if instance, ok = alert.Labels["instance"]; !ok {
				instance = "???"
			}

			if summary, ok = alert.Annotations["summary"]; !ok {
				summary = "???"
			}

			table.Append([]string{instance, summary})
		}

		table.Render()

		text.WriteString(tableBuf.String())

		// Prepare the payload for Mattermost
		payload := map[string]string{
			"text":     text.String(),
			"icon_url": "https://www.mattermost.org/wp-content/uploads/2016/04/icon.png",
			"channel":  "shitta",
			"username": "Prometeo",
		}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Cannot encode the destination payload: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest("POST", mattermostURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			log.Printf("Cannot create POST request: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		log.Printf("Sending notification to mattermost\n")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Cannot send POST request: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println("Response status:", resp.Status)
		fmt.Println("Response:", string(body))
	})
	return h
}

// StartServer starts the baracca
func StartServer(address string, webhookURL string) {
	http.HandleFunc("/health", healthHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/alert", alertHandler(webhookURL))

	log.Fatal(http.ListenAndServe(address, nil))
}
