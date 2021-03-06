package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

var configFile = "~/.jira_config.toml"

//Page representation of a Page
type Page struct {
	Title  string
	Header string
}

//Issue is a representation of a Jira Issue
type Issue struct {
	Fields struct {
		Project struct {
			Key string `json:"key"`
		} `json:"project"`
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Issuetype   struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Priority struct {
			ID string `json:"id"`
		} `json:"priority"`
	} `json:"fields"`
}

//Transition defines a transition json object. Used for starting, stoppinp
//generally for state stranfer
type Transition struct {
	Fields struct {
		Resolution struct {
			Name string `json:"name"`
		} `json:"resolution"`
	} `json:"fields"`
	Transition struct {
		ID string `json:"id"`
	} `json:"transition"`
}

//Credentials a representation of a JIRA config which helds API permissions
type Credentials struct {
	Username string
	Password string
	URL      string
}

func (cred *Credentials) initConfig() {
	if _, err := os.Stat(configFile); err != nil {
		panic(err)
	}

	if _, err := toml.DecodeFile(configFile, cred); err != nil {
		panic(err)
	}
}

func main() {
	handlerChain := alice.New(Logging, PanicHandler)
	router := mux.NewRouter().StrictSlash(true)
	router.Handle("/create", handlerChain.ThenFunc(createIssue)).Methods("POST")
	router.Handle("/", handlerChain.ThenFunc(renderMainPage)).Methods("GET")
	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
	log.Printf("Starting server to listen on port: 8989...")
	http.ListenAndServe(":8989", router)
}

func renderMainPage(w http.ResponseWriter, r *http.Request) {
	page := Page{"JIRA Web", "Header"}
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		panic(err)
	}
	tmpl.Execute(w, page)
}

func closeIssue(issueKey string) {
	if issueKey == "" {
		log.Panic("Please provide an issueID with -k")
	}
	fmt.Println("Closing issue number: ", issueKey)

	var trans Transition

	//TODO: Add the ability to define a comment for the close reason
	trans.Fields.Resolution.Name = "Dummy Flag"
	trans.Transition.ID = "2"
	marhsalledTrans, err := json.Marshal(trans)
	if err != nil {
		log.Panic("Error occured when marshaling transition: ", err)
	}
	fmt.Println("Marshalled:", trans)
	sendRequest(marhsalledTrans, "POST", issueKey+"/transitions?expand=transitions.fields")
}

func startIssue(issueID string) {
	if issueID == "" {
		log.Panic("Please provide an issueID with -i")
	}

	fmt.Println("Starting issue number:", issueID)
}

func createIssue(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating new issue.")
	var issue Issue
	issue.Fields.Description = r.FormValue("description")
	issue.Fields.Priority.ID = r.FormValue("priority")
	issue.Fields.Summary = r.FormValue("summary")
	issue.Fields.Project.Key = r.FormValue("key")
	issue.Fields.Issuetype.Name = "Task"
	log.Println("Creating issue with values:", issue)
	marshalledIssue, err := json.Marshal(issue)
	if err != nil {
		log.Panic("Error occured when Marshaling Issue:", err)
	}
	sendRequest(marshalledIssue, "POST", "")
}

func sendRequest(jsonStr []byte, method string, url string) {
	cred := &Credentials{}
	cred.initConfig()
	fmt.Println("Json:", string(jsonStr))
	req, err := http.NewRequest(method, cred.URL+url, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cred.Username, cred.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

}

func (i Issue) String() string {
	return fmt.Sprintf("Des: {%s}, IssueType:{%s}, Priority:{%s}, Project:{%s}, Summary:{%s}", i.Fields.Description, i.Fields.Issuetype, i.Fields.Priority, i.Fields.Project, i.Fields.Summary)
}
