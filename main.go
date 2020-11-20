package main

import (
	"encoding/xml"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	// "github.com/skratchdot/open-golang/open"
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
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("CC")
	systray.SetTooltip("CC Menu COPY")
	mQuitOrig := systray.AddMenuItem("Quit", "Quit the whole app")
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for {
			select {
			case <-mQuitOrig.ClickedCh:
				fmt.Println("Requesting quit")
				systray.Quit()
				fmt.Println("Finished quitting")
			case <-ticker.C:
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
					//b, _ := ioutil.ReadAll(resp.Body)
					//fmt.Println(string(b))
					//err = xml.Unmarshal(b, &projects)
					err = xml.NewDecoder(resp.Body).Decode(&projects)
					if err != nil {
						panic(err)
					}
					fmt.Printf("%#+v\n", projects)

					for _, project := range projects.Projects {
						_ = systray.AddMenuItem(project.Name, project.Name)
					}
				}
			}
		}
	}()

	/*
		// We can manipulate the systray in other goroutines
		go func() {
			systray.SetTemplateIcon(icon.Data, icon.Data)
			systray.SetTitle("Awesome App")
			systray.SetTooltip("Pretty awesome棒棒嗒")
			mChange := systray.AddMenuItem("Change Me", "Change Me")
			mChecked := systray.AddMenuItemCheckbox("Unchecked", "Check Me", true)
			mEnabled := systray.AddMenuItem("Enabled", "Enabled")
			// Sets the icon of a menu item. Only available on Mac.
			mEnabled.SetTemplateIcon(icon.Data, icon.Data)

			systray.AddMenuItem("Ignored", "Ignored")

			subMenuTop := systray.AddMenuItem("SubMenuTop", "SubMenu Test (top)")
			subMenuMiddle := subMenuTop.AddSubMenuItem("SubMenuMiddle", "SubMenu Test (middle)")
			subMenuBottom := subMenuMiddle.AddSubMenuItemCheckbox("SubMenuBottom - Toggle Panic!", "SubMenu Test (bottom) - Hide/Show Panic!", false)
			subMenuBottom2 := subMenuMiddle.AddSubMenuItem("SubMenuBottom - Panic!", "SubMenu Test (bottom)")

			mUrl := systray.AddMenuItem("Open UI", "my home")
			mQuit := systray.AddMenuItem("退出", "Quit the whole app")

			// Sets the icon of a menu item. Only available on Mac.
			mQuit.SetIcon(icon.Data)

			systray.AddSeparator()
			mToggle := systray.AddMenuItem("Toggle", "Toggle the Quit button")
			shown := true
			toggle := func() {
				if shown {
					subMenuBottom.Check()
					subMenuBottom2.Hide()
					mQuitOrig.Hide()
					mEnabled.Hide()

					shown = false
				} else {
					subMenuBottom.Uncheck()
					subMenuBottom2.Show()
					mQuitOrig.Show()
					mEnabled.Show()
					shown = true
				}
			}

			for {
				select {
				case <-mChange.ClickedCh:
					mChange.SetTitle("I've Changed")
				case <-mChecked.ClickedCh:
					if mChecked.Checked() {
						mChecked.Uncheck()
						mChecked.SetTitle("Unchecked")
					} else {
						mChecked.Check()
						mChecked.SetTitle("Checked")
					}
				case <-mEnabled.ClickedCh:
					mEnabled.SetTitle("Disabled")
					mEnabled.Disable()
				case <-mUrl.ClickedCh:
					open.Run("https://www.getlantern.org")
				case <-subMenuBottom2.ClickedCh:
					panic("panic button pressed")
				case <-subMenuBottom.ClickedCh:
					toggle()
				case <-mToggle.ClickedCh:
					toggle()
				case <-mQuit.ClickedCh:
					systray.Quit()
					fmt.Println("Quit2 now...")
					return
				}
			}
		}()
	*/
}
