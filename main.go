package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	fbGraph "github.com/huandu/facebook"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

const (
	Fields = "name,location,description,phone,emails,about,general_info"
	Type   = "page"
)

type category struct {
	name string
	fbGraph.Params
}

var chDone chan string
var chCategoryResults chan map[string]string

var accessToken *oauth2.Token
var session *fbGraph.Session
var (
	oauthConf = &oauth2.Config{
		ClientID:     "1426521580999261",
		ClientSecret: "5836e4eda9ecf1305af72f802a14252f",
		Scopes:       []string{"email"},
		RedirectURL:  "http://localhost:3000/facebook_cb",
		Endpoint:     facebook.Endpoint,
	}
	oauthStateString = "thisshouldberandom"
)

func init() {
	chDone = make(chan string, 10)
	chCategoryResults = make(chan map[string]string, 100)
}

func main() {
	if token, err := GetToken(); err == nil {
		session = &fbGraph.Session{
			Version:    "v2.4",
			HttpClient: oauthConf.Client(oauth2.NoContext, token),
		}

		if err := session.Validate(); err != nil {
			log.Println("Your access token is invalid! \n Please revoke token http://localhost:3000/login")
			RunHTTPServer()
		} else {
			log.Println("Start facebook graph calling")
			if categoryList, paramErr := buildCategoryList(); paramErr == nil {
				log.Printf("Handy %d categories", len(categoryList))

				for _, c := range categoryList {
					go invokeFBGraph(c)
				}
				var accumuratedResult = make([]interface{}, 0, 1000)
				for c := 0; c < len(categoryList); {
					select {
					case categoryResult := <-chCategoryResults:
						accumuratedResult = append(accumuratedResult, categoryResult)
					case name := <-chDone:
						c++
						log.Printf("done '%s'", name)
					}
				}
				log.Printf("All categories done %d", len(categoryList))
				log.Printf("Writing accumurated resuls of %d records", len(accumuratedResult))
				if jsonStr, marshalErr := json.MarshalIndent(accumuratedResult, "", "    "); marshalErr == nil {
					if writeErr := writeResult("", "all-lat-long.json", string(jsonStr)); writeErr != nil {
						log.Println(writeErr)
					}
				} else {
					log.Println(marshalErr)
				}
			} else {
				log.Panic(paramErr)
			}
		}
	} else {
		log.Panic(err)
	}
}

func buildCategoryList() ([]*category, error) {
	bytes, err := ioutil.ReadFile("type.json")
	if err != nil {
		return nil, err
	}
	var categoriesData struct {
		Categories []string `json:"categories"`
	}

	unMarshalErr := json.Unmarshal(bytes, &categoriesData)
	if unMarshalErr != nil {
		return nil, unMarshalErr
	}
	var categoryList = make([]*category, 0, 10)

	for _, c := range categoriesData.Categories {
		categoryList = append(categoryList, &category{
			c,
			fbGraph.Params{
				"q":      c,
				"type":   Type,
				"fields": Fields,
			},
		})
	}
	return categoryList, nil
}

func invokeFBGraph(cat *category) {
	defer func() {
		chDone <- cat.name
	}()
	res, err := session.Get("/search", cat.Params)
	if err != nil {
		log.Panic(err)
	}
	paging, _ := res.Paging(session)
	var pageNo = 1
	for {
		results := paging.Data()
		for _, r := range results {
			var data = make(map[string]string)
			data["category"] = cat.name
			location := r.Get("location")
			if location != nil {
				m := location.(map[string]interface{})
				for k, v := range m {
					switch k {
					case "latitude":
						data["lat"] = string(v.(json.Number))
					case "longitude":
						data["long"] = string(v.(json.Number))
					}
				}
				if data["lat"] != "" && data["long"] != "" {
					chCategoryResults <- data
				}
			}
		}
		if jsonStr, marshalErr := json.MarshalIndent(results, "", "    "); marshalErr == nil {
			filename := generateFilename(cat.name, pageNo)
			if writeErr := writeResult(cat.name, filename, string(jsonStr)); writeErr != nil {
				log.Println(writeErr)
			} else {
				log.Printf("wrote result : %s", filename)
				pageNo++
			}
		} else {
			log.Println(marshalErr)
		}

		if noMore, err := paging.Next(); noMore || err != nil {
			break
		}
	}

}

func writeResult(dir, filename, jsonString string) error {
	dirErr := os.Mkdir("./dataset/"+dir, 0777)
	if !os.IsExist(dirErr) {
		return dirErr
	}

	f, createErr := os.Create(filename)
	if createErr != nil {
		return createErr
	}
	_, writeErr := f.WriteString(jsonString)
	return writeErr
}

func generateFilename(dirname string, pageNo int) string {
	return "./dataset/" + dirname + "/page-" + strconv.Itoa(pageNo) + ".json"
}
