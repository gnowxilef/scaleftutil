package scaleftutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var key_id = os.Getenv("SCALEFT_KEY_ID")
var key_secret = os.Getenv("SCALEFT_KEY_SECRET")
var token = ""

var key_team = os.Getenv("SCALEFT_TEAM")
var project = os.Getenv("SCALEFT_PROJECT")

const api_url string = "https://app.scaleft.com/v1/"

func request(method string, url_part string, params *bytes.Buffer) (*http.Response, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%v/%v", api_url, url_part)

	req, err := http.NewRequest(method, url, params)
	req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", get_token))
	}
	resp, err := client.Do(req)

	return resp, err
}

func DeleteServersByHostname(hostname string) error {

	//log.Printf("[DEBUG] key_id:%s key_secret:%s key_team:%s project:%s hostname:%s", key_id, key_secret, key_team, project, hostname)

	list, err := get_servers

	if err != nil {
		return fmt.Errorf("Error getting server list. key_team:%s error:%v", key_team, err)
	}

	ids := get_ids_for_hostname(hostname, list)

	if len(ids) == 0 {
		//	return fmt.Errorf("Error, ScaleFT api returned no servers that matched hostname:%s", hostname)
		//      This should not happen, but if it does, it's ok?
		log.Printf("[WARN] No servers matched for Hostname:%s, Team:%s, Project:%s.  We'll keep going though.", hostname, key_team, project)
		return nil
	}

	for _, id := range ids {
		err := delete_server(id)
		if err != nil {
			log.Printf("[WARN] Failed to delete server with hostname: %s at ScaleFT ID:%s, error:%s", hostname, id, err)
			//              return fmt.Errorf("Error deleting server at id:%s and key_team:%s project: %s error:%v", id, key_team, project, err)
		}
	}

	return nil
}

func DeleteServersByPattern(pattern string) error {

	// log.Printf("[DEBUG] key_id:%s key_secret:%s key_team:%s project:%s hostname:%s", key_id, key_secret, key_team, project, hostname)

	list, err := get_servers

	if err != nil {
		return fmt.Errorf("Error getting server list. key_team:%s error:%v", key_team, err)
	}

	ids := get_ids_for_pattern(pattern, list)

	if len(ids) == 0 {
		//      return fmt.Errorf("Error, ScaleFT api returned no servers that matched hostname:%s", hostname)
		//      This should not happen, but if it does, it's ok?
		log.Printf("[WARN] No servers matched for pattern:%s, Team:%s, Project:%s.  We'll keep going though.", pattern, key_team, project)
		return nil
	}

	for _, id := range ids {
		err := delete_server(id)
		if err != nil {
			log.Printf("[WARN] Failed to delete server with ScaleFT ID:%s, error:%s", id, err)
			//              return fmt.Errorf("Error deleting server at id:%s and key_team:%s project: %s error:%v", id, key_team, project, err)
		}
	}

	return nil
}

type Body struct {
	Key_id     string `json:"key_id"`
	Key_secret string `json:"key_secret"`
}

type Bearer struct {
	Bearer_token string `json:"bearer_token"`
}

type Server struct {
	Id              string                 `json:"id"`
	ProjectName     string                 `json:"project_name"`
	Hostname        string                 `json:"hostname"`
	AltNames        []string               `json:"alt_names"`
	AccessAddress   string                 `json:"access_address"`
	OS              string                 `json:"os"`
	RegisteredAt    time.Time              `json:"registered_at"`
	LastSeen        time.Time              `json:"last_seen"`
	CloudProvider   string                 `json:"cloud_provider"`
	SSHHostKeys     []string               `json:"ssh_host_keys"`
	BrokerHostCerts []string               `json:"broker_host_certs"`
	InstanceDetails map[string]interface{} `json:"instance_details"`
	State           string                 `json:"state"`
}

type Servers struct {
	List []*Server `json:"list"`
}

func get_token() (string, error) {
	p := &Body{key_id, key_secret}
	jsonStr, err := json.Marshal(p)
	resp, err := request("POST", fmt.Sprintf("%v/service_token", key_team), bytes.NewBuffer(jsonStr))

	if err != nil {
		return "error", fmt.Errorf("Error getting token key_id:%s key_team:%s status:%s error:%v", key_id, key_team, string(resp.Status), err)
	}

	defer resp.Body.Close()
	b := Bearer{}
	json.NewDecoder(resp.Body).Decode(&b)

	return b.Bearer_token, err
}

func get_logs(bearer_token string, key_team string) string {
	resp, err := request("GET", fmt.Sprintf("%v/audits", key_team), nil)
	if err != nil {
		panic(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	s := string(bodyText)
	return s
}

func get_servers (Servers, error) {
	resp, err := request("GET", fmt.Sprintf("%v/projects/%v/servers", key_team, project), nil)
	if err != nil {
		fmt.Errorf("Error listing servers - key_team: %s, project: %s, status: %s, error: %v", key_team, project, string(resp.Status), err)
	}

	s := struct {
		List []*Server `json:"list"`
	}{nil}

	json.NewDecoder(resp.Body).Decode(&s)

	return s, err
}

func delete_server(server_id string) error {
	resp, err := request("DELETE", fmt.Sprintf("%v/projects/%v/servers/%v", key_team, project, server_id), nil)
	if err != nil {
		return fmt.Errorf("Error deleting server: %s - status: %s, error: %v", server_id, string(resp.Status), err)
	}
	return nil
}

func get_ids_for_hostname(hostname string, server_list Servers) []string {
	filtered := make([]string, len(server_list.List))
	for i, l := range server_list.List {
		if hostname == l.Hostname {
			filtered[i] = l.Id
		}
	}
	return filtered
}

func get_ids_for_pattern(pattern string, server_list Servers) []string {
	filtered := make([]string, len(server_list.List))
	for i, l := range server_list.List {
		if strings.Contains(l.Hostname, pattern) {
			filtered[i] = l.Id
		}
	}
	return filtered
}
