package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	ubot "github.com/UBotPlatform/UBot.Common.Go"
)

var api *ubot.AppApi
var corpus = new(CorpusNode)

func onReceiveChatMessage(bot string, msgType ubot.MsgType, source string, sender string, message string, info ubot.MsgInfo) (ubot.EventResultType, error) {
	result := corpus.Query(message)
	if result != nil {
		_ = api.SendChatMessage(bot, msgType, source, sender, result.Reply)
		return ubot.CompleteEvent, nil
	}
	return ubot.IgnoreEvent, nil
}

func main() {
	executableFile, err := os.Executable()
	ubot.AssertNoError(err)
	corpusFolder := filepath.Join(filepath.Dir(executableFile), "corpus")
	corpusFiles, err := ioutil.ReadDir(corpusFolder)
	if err != nil {
		fmt.Println("Failed to read corpus folder:", corpusFolder)
	}
	for _, corpusFile := range corpusFiles {
		if corpusFile.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(corpusFile.Name())) != ".txt" {
			continue
		}
		fullPath := filepath.Join(corpusFolder, corpusFile.Name())
		err = corpus.LoadFile(fullPath)
		if err != nil {
			fmt.Println("Failed to load:", corpusFile.Name())
		} else {
			fmt.Println("Loaded file:", corpusFile.Name())
		}
	}
	fmt.Println("Corpus item count:", corpus.Count())
	err = ubot.HostApp("Corpus", func(e *ubot.AppApi) *ubot.App {
		api = e
		return &ubot.App{
			OnReceiveChatMessage: onReceiveChatMessage,
		}
	})
	ubot.AssertNoError(err)
}
