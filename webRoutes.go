package main

import (
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Param struct {
	Name string
	Ex   string
	Type string
}
type RoutesStore struct {
	Path   string
	Method string
	Params []Param
}

var routes []RoutesStore

func webStartServer() {
	//GIN
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.LoadHTMLGlob("resources/templates/*")
	router.Static("/static", "./static")
	router.GET("", webIndex)
	v1 := router.Group("/api/v1")
	{
		v1.GET("/reload", webReload)
		routes = append(routes, RoutesStore{Path: v1.BasePath() + "/reload", Method: "GET", Params: []Param{}})
		v1.GET("/start", webStart)
		routes = append(routes, RoutesStore{Path: v1.BasePath() + "/start", Method: "GET", Params: []Param{{Name: "code", Ex: "[abc123def]", Type: "string"}, {Name: "fadetime", Ex: "2000", Type: "int in ms"}}})
		v1.GET("/play", webPlay)
		routes = append(routes, RoutesStore{Path: v1.BasePath() + "/play", Method: "GET", Params: []Param{{Name: "code", Ex: "[abc123def]", Type: "string"}}})
		v1.GET("/pause", webPause)
		routes = append(routes, RoutesStore{Path: v1.BasePath() + "/pause", Method: "GET", Params: []Param{{Name: "code", Ex: "[abc123def]", Type: "string"}}})
		v1.GET("/stop", webStop)
		routes = append(routes, RoutesStore{Path: v1.BasePath() + "/stop", Method: "GET", Params: []Param{{Name: "code", Ex: "[abc123def]", Type: "string"}}})
		v1.GET("/fadeoutstop", webFadeOutStop)
		routes = append(routes, RoutesStore{Path: v1.BasePath() + "/fadeoutstop", Method: "GET", Params: []Param{{Name: "code", Ex: "[abc123def]", Type: "string"}, {Name: "fadetime", Ex: "2000", Type: "int in ms"}}})
		v1.GET("/fadeoutstopall", webFadeOutStopAll)
		routes = append(routes, RoutesStore{Path: v1.BasePath() + "/fadeoutstopall", Method: "GET", Params: []Param{{Name: "fadetime", Ex: "2000", Type: "int in ms"}}})
		v1.GET("/stopall", webStopAll)
		routes = append(routes, RoutesStore{Path: v1.BasePath() + "/stopall", Method: "GET", Params: []Param{}})
		v1.GET("/pauseall", webPauseAll)
		routes = append(routes, RoutesStore{Path: v1.BasePath() + "/pauseall", Method: "GET", Params: []Param{}})
		v1.GET("/playall", webPlayAll)
		routes = append(routes, RoutesStore{Path: v1.BasePath() + "/playall", Method: "GET", Params: []Param{}})
		v1.POST("/upload", webUpload)
		v1.GET("/delete", webDelete)
		v1.POST("/saveconfig", webSaveConfig)
	}
	CreateWebsocketServer(router)
	go router.Run("127.0.0.1:" + strconv.Itoa(conf.HttpPort))

}
func webSaveConfig(c *gin.Context) {
	newConfig := Configuration{}

	if err := c.ShouldBindJSON(&newConfig); err != nil {
		fmt.Println("bindjson failed")
		fmt.Println(newConfig)
		return
	}
	fmt.Println(newConfig)
	saveConfig(newConfig)
}
func webIndex(c *gin.Context) {
	playeronly, err := strconv.Atoi(c.DefaultQuery("playeronly", "0"))
	if err != nil {
		playeronly = 0
	}
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title":      "Main website",
		"AudioFiles": AudioFiles,
		"Playing":    Playing,
		"Conf":       conf,
		"Routes":     routes,
		"Playeronly": playeronly == 1,
	})
}
func webReload(c *gin.Context) {
	loadAudioFiles()
}

func webStopAll(c *gin.Context) {
	FadeoutStopAll(0)
}
func webUpload(c *gin.Context) {
	file, _ := c.FormFile("file")
	fmt.Println(file.Filename)

	// Upload the file to specific dst.
	c.SaveUploadedFile(file, path.Join(conf.SoundFolder, file.Filename))

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
}
func webFadeOutStopAll(c *gin.Context) {
	fadetime, err := strconv.Atoi(c.DefaultQuery("fadetime", "0"))
	if err != nil {
		fadetime = 0
	}
	FadeoutStopAll(fadetime)
}
func webPlay(c *gin.Context) {
	audiofile, err := GetAudioFileByCode(c.Query("code"))
	if err != nil {
		return
	}
	go PlaySound(audiofile)
}
func webStart(c *gin.Context) {
	audiofile, err := GetAudioFileByCode(c.Query("code"))
	if err != nil {
		return
	}
	fadetime, err := strconv.Atoi(c.DefaultQuery("fadetime", "0"))
	if err != nil {
		fadetime = 0
	}
	code := uuid.New().String()
	go StartSound(audiofile, false, fadetime, code)
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
func webPause(c *gin.Context) {
	audiofile, err := GetAudioFileByCode(c.Query("code"))
	if err != nil {
		return
	}
	go PauseSound(audiofile)
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
func webPauseAll(c *gin.Context) {
	for _, audiofile := range Playing {
		go PauseSound(audiofile)
	}
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func webPlayAll(c *gin.Context) {
	for _, audiofile := range Playing {
		go PlaySound(audiofile)
	}
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
func webStop(c *gin.Context) {
	audiofile, err := GetAudioFileByCode(c.Query("code"))
	if err != nil {
		return
	}
	go StopSound(audiofile)
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func webFadeOutStop(c *gin.Context) {
	defaultfadetime := strconv.Itoa(conf.DefaultFadeTime)
	fadetime, err := strconv.Atoi(c.DefaultQuery("fadetime", defaultfadetime))
	if err != nil {
		return
	}
	audiofile, err := GetAudioFileByCode(c.Query("code"))
	if err != nil {
		return
	}
	go FadeoutStop(audiofile, fadetime)
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func webDelete(c *gin.Context) {
	audiofile, err := GetAudioFileByCode(c.Query("code"))
	if err != nil {
		return
	}
	DeleteAudioFile(audiofile)
}

// OLD
/*
func handleCurrent(w http.ResponseWriter, r *http.Request) {
	// Rejct everything else then GET requests
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json") // Set the content type to json

	for i := 0; i < len(playbacks); i++ {
		plyB := playbacks[i]
		seeker := plyB.Streamer
		format := plyB.Format

		if seeker != nil {
			fmt.Println(format.SampleRate)
			// fmt.Println(plyB.Seeker.)
			position := plyB.Format.SampleRate.D(seeker.Position())
			length := plyB.Format.SampleRate.D(seeker.Len())
			remaining := length - position
			if remaining == 0 {
				plyB.Done <- true
			}
		}
	}

	var tempResultSet map[int]playbackWebReturn = make(map[int]playbackWebReturn) // Create a new map to store the results
	// Iterate through the playbacks map and add important information to the tempResultSet map
	for index, element := range playbacks {
		tempResultSet[index] = playbackWebReturn{File: element.File, IsLoaded: element.IsLoaded, Id: index}
	}
	// Convert the map to a JSON object and return it to the user
	j, err := json.Marshal(tempResultSet)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		fmt.Println(string(j))
		fmt.Fprintf(w, string(j))
	}
}

func handleListing(w http.ResponseWriter, r *http.Request) {
	// Rejct everything else then GET requests
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json") // Set the content type to json

	// Convert the array to a JSON object and return it to the user
	j, err := json.Marshal(AudioFiles)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		fmt.Println(string(j))
		fmt.Fprintf(w, string(j))
	}
}

func handleRemaining(w http.ResponseWriter, r *http.Request) {
	// Rejct everything else then GET requests
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")   // Set the content type to json
	var cnt, err = strconv.Atoi(r.URL.Query().Get("id")) // Retrieve the id, first convert it to an int
	if err != nil {
		fmt.Fprintf(w, "{\"status\":\"fail\", \"reason\":\"invalid id\"}")
		return
	}
	fmt.Println(cnt)
	plyB := playbacks[cnt]
	// fmt.Println(beep.SampleRate.D(plyB.Streamer.Stream().Len()))
	seeker := plyB.Streamer
	format := plyB.Format
	n := plyB.Format.SampleRate // Streamer.Stream() // .At(beep.SampleRate.D(plyB.Streamer.Stream().Len()))

	if seeker != nil {
		fmt.Println(format.SampleRate)
		// fmt.Println(plyB.Seeker.)
		position := plyB.Format.SampleRate.D(seeker.Position())
		length := plyB.Format.SampleRate.D(seeker.Len())
		remaining := length - position
		volume := plyB.Volume.Volume
		if remaining == 0 {
			plyB.Done <- true
		}
		fmt.Println(position)
		fmt.Fprintf(w, "{\"status\":\"ok\", \"id\":%d, \"SampleRate\":%d, \"Volume\":%f, \"Length\":%d, \"Position\":%d, \"Remaining\": %d, \"LengthSec\":\"%v\", \"PosSec\":\"%v\", \"RemaningSec\":\"%v\"}", cnt, n, volume, length, position, remaining, length, position, remaining)
	} else {
		fmt.Println("Seeker is nil")
		fmt.Fprintf(w, "{\"status\":\"ok\", \"SampleRate\":%d}", n)
	}

}
*/
