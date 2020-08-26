package main

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type CorpusItem struct {
	Reply           string
	InversePriority uint
}

type CorpusNode struct {
	Endpoint []CorpusItem
	Next     map[rune]*CorpusNode
}

func (c *CorpusNode) Add(keyword string, item CorpusItem) {
	if len(keyword) == 0 {
		c.Endpoint = append(c.Endpoint, item)
		return
	}
	r, width := utf8.DecodeRuneInString(keyword)
	if c.Next == nil {
		c.Next = make(map[rune]*CorpusNode)
	}
	v, ok := c.Next[r]
	if !ok {
		v = new(CorpusNode)
		c.Next[r] = v
	}
	v.Add(keyword[width:], item)
}

func (c *CorpusNode) QueryPrefix(prefix string) *CorpusItem {
	var width int
	var single rune
	curNode := c
	var result *CorpusItem = nil
	for str := prefix; len(str) > 0; str = str[width:] {
		single, width = utf8.DecodeRuneInString(str)
		var ok bool
		curNode, ok = curNode.Next[single]
		if !ok {
			break
		}
		if len(curNode.Endpoint) != 0 {
			curEndpoint := curNode.Endpoint[time.Now().UnixNano()%int64(len(curNode.Endpoint))]
			if result == nil || result.InversePriority > curEndpoint.InversePriority {
				result = &curEndpoint
				if result.InversePriority == 0 {
					break
				}
			}
		}
	}
	return result
}

func (c *CorpusNode) Query(context string) *CorpusItem {
	var width int
	var result *CorpusItem = nil
	for str := context; len(str) > 0; str = str[width:] {
		_, width = utf8.DecodeRuneInString(str)
		endpoint := c.QueryPrefix(str)
		if endpoint != nil {
			if result == nil || result.InversePriority > endpoint.InversePriority {
				result = endpoint
				if result.InversePriority == 0 {
					break
				}
			}
		}
	}
	return result
}

func (c *CorpusNode) Reset() {
	c.Endpoint = nil
	c.Next = nil
}

func (c *CorpusNode) Load(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	keyword := ""
	inversePriority := uint(0)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if keyword != "" {
			c.Add(keyword, CorpusItem{line, inversePriority})
			keyword = ""
			inversePriority = 0
			continue
		}
		if strings.HasPrefix(line, "*inverse_priority=") {
			inversePriority64, _ := strconv.ParseUint(strings.TrimPrefix(line, "*inverse_priority="), 10, 64)
			inversePriority = uint(inversePriority64)
			continue
		}
		keyword = line
	}
	return scanner.Err()
}

func (c *CorpusNode) LoadFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	err = c.Load(file)
	return err
}

func (c *CorpusNode) Count() int {
	r := 0
	if c.Endpoint != nil {
		r++
	}
	for _, next := range c.Next {
		r += next.Count()
	}
	return r
}
