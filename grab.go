package main

import (
	"compress/flate"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/andybalholm/brotli"
)

func grabWord(word string) {
	url := fmt.Sprintf("https://app.vocabgo.com/student/api/Student/Course/StudyWordInfo?course_id=%s&list_id=%s&word=%s&timestamp=%d&versions=%s&app_type=1",
		dataset.CurrentTask.TaskID,
		dataset.CurrentTask.TaskSet,
		word,
		time.Now().UnixMilli(),
		dataset.RequestInfo.Versions)
	req, _ := http.NewRequest("GET", url, nil)
	//Adds headers
	req.Header = dataset.RequestInfo.Header
	//Adds cookies
	for i := 0; i < len(dataset.RequestInfo.Cookies); i++ {
		req.AddCookie(dataset.RequestInfo.Cookies[i])
	}
	//Do request
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("[E]" + err.Error())
	}
	defer response.Body.Close()
	read, _ := switchContentEncoding(response)
	raw, _ := io.ReadAll(read)
	var rawData VocabRawJSONStruct
	json.Unmarshal(raw, &rawData)
	rawJSON, _ := base64.StdEncoding.DecodeString(rawData.Data[32:])
	var wordInfoRaw WordInfoJSON
	json.Unmarshal(rawJSON, &wordInfoRaw)
	//Add to word storage
	var wordInfo WordInfo
	wordInfo.Word = word
	for i := 0; i < len(wordInfoRaw.Options); i++ {
		var content WordInfoContent
		content.Meaning = wordInfoRaw.Options[i].Content.Mean
		//append usage
		for j := 0; j < len(wordInfoRaw.Options[i].Content.Usage); j++ {
			content.Usage = append(content.Usage, wordInfoRaw.Options[i].Content.Usage[j])
		}
		//append example's english expression
		for j := 0; j < len(wordInfoRaw.Options[i].Content.Example); j++ {
			content.ExampleEnglish = append(content.ExampleEnglish, wordInfoRaw.Options[i].Content.Example[j].SenContent)
		}
		wordInfo.Content = append(wordInfo.Content, content)
	}
	words = append(words, wordInfo)
}

func switchContentEncoding(res *http.Response) (bodyReader io.Reader, err error) {
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		bodyReader, err = gzip.NewReader(res.Body)
	case "deflate":
		bodyReader = flate.NewReader(res.Body)
	case "br":
		bodyReader = brotli.NewReader(res.Body)
	default:
		bodyReader = res.Body
	}
	return
}
