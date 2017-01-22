package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// https://developer.github.com/v3/markdown/
type RequestJSON struct {
	Text    string `json:"text"`
	Mode    string `json:"mode"`
	Context string `json:"context"`
}

func main() {
	var inputFile string
	var outputFile string
	var accessToken string
	var templateFile string
	flag.StringVar(&accessToken, "access_token", "", "github account access token")
	flag.StringVar(&inputFile, "input_file", "README.md", "the input markdown filename")
	flag.StringVar(&outputFile, "output_file", "", "the output html filename")
	flag.StringVar(&templateFile, "template", "", "the output html template")
	flag.Parse()

	var outStream io.Writer
	if len(outputFile) == 0 {
		outStream = os.Stdout
	} else {
		f, err := os.OpenFile(outputFile, os.O_WRONLY, 0640)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		outStream = f
	}

	fin, err := os.Open(inputFile)
	if err != nil {
		panic(err)
	}
	buffer, err := ioutil.ReadAll(fin)
	if err != nil {
		panic(err)
	}
	reqJSON := RequestJSON{}
	reqJSON.Text = string(buffer[:])
	reqJSON.Mode = "gfm"
	reqJSON.Context = ""

	jsonStr, err := json.Marshal(reqJSON)
	if err != nil {
		panic(err)
	}

	var url = "https://api.github.com/markdown"
	if len(accessToken) != 0 {
		url += "?access_token=" + accessToken
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		outStream.Write(result)
	} else {
		err = tmpl.Execute(outStream, struct {
			Markdown template.HTML
		}{
			template.HTML(result[:]),
		})
		if err != nil {
			panic(err)
		}
	}
}
