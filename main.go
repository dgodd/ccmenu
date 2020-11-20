package main

import (
	"encoding/xml"
	"fmt"
	"github.com/getlantern/systray"
	// "github.com/getlantern/systray/example/icon"
	"github.com/skratchdot/open-golang/open"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Project struct {
	Name            string `xml:"name,attr"`
	Activity        string `xml:"activity,attr"`
	LastBuildStatus string `xml:"lastBuildStatus,attr"`
	LastBuildLabel  string `xml:"lastBuildLabel,attr"`
	LastBuildTime   string `xml:"lastBuildTime,attr"`
	WebURL          string `xml:"webUrl,attr"`
}

func main() {
	onExit := func() {
		now := time.Now()
		ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	greyICO, err := ioutil.ReadFile("/home/dgodd/workspace/ccmenu/grey.png")
	if err != nil {
		panic(err)
	}
	greenICO, err := ioutil.ReadFile("/home/dgodd/workspace/ccmenu/green.png")
	if err != nil {
		panic(err)
	}
	redICO, err := ioutil.ReadFile("/home/dgodd/workspace/ccmenu/red.png")
	if err != nil {
		panic(err)
	}

	systray.SetTemplateIcon(greyICO, greyICO)
	systray.SetTitle("CC")
	// systray.SetTooltip("CC Menu COPY")
	mQuitOrig := systray.AddMenuItem("Quit", "Quit the whole app")
	ticker := time.NewTicker(10 * time.Second)

	items := make(map[string]*systray.MenuItem)
	getData := func() {
		configs := []struct {
			URL     string   `yaml:"url"`
			Exclude []string `yaml:"exclude"`
		}{}
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		b, err := ioutil.ReadFile(filepath.Join(home, ".ccmenu.yaml"))
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(b, &configs)
		if err != nil {
			panic(err)
		}

		allStatus := 0
		for _, config := range configs {
			fmt.Println(config)
			resp, err := http.Get(config.URL)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			projects := struct {
				Projects []Project `xml:"Project"`
			}{}
			err = xml.NewDecoder(resp.Body).Decode(&projects)
			if err != nil {
				panic(err)
			}

			for _, project := range projects.Projects {
				fmt.Printf("%s - %s - %s\n", project.Name, project.LastBuildStatus, project.Activity)
				if _, exclude := Find(config.Exclude, project.Name); exclude {
					continue
				}
				item := items[project.Name]
				if item == nil {
					item = systray.AddMenuItem(project.Name, project.Name)
					item.SetTemplateIcon(greyICO, greyICO)
					url := project.WebURL
					go func() {
						for {
							<-item.ClickedCh
							open.Run(url)
						}
					}()
					items[project.Name] = item
				}

				// Activity:"Sleeping", LastBuildStatus:"Success"
				if project.LastBuildStatus == "Success" {
					item.SetTemplateIcon(greenICO, greenICO)
					if allStatus < 1 {
						allStatus = 1
					}
				} else if project.LastBuildStatus == "Failure" {
					item.SetTemplateIcon(redICO, redICO)
					if allStatus < 2 {
						allStatus = 2
					}
				}
			}
		}

		switch allStatus {
		case 0:
			systray.SetTemplateIcon(greenICO, greenICO)
		case 1:
			systray.SetTemplateIcon(redICO, redICO)
		}
	}
	getData()

	go func() {
		for {
			select {
			case <-mQuitOrig.ClickedCh:
				fmt.Println("Requesting quit")
				systray.Quit()
				fmt.Println("Finished quitting")
			case <-ticker.C:
				getData()
			}
		}
	}()
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
