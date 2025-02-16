package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

func workerThread(offset *int, mu *sync.Mutex, lines *[]string, wg *sync.WaitGroup, outLines *[]string) {
	defer wg.Done()

	var localOffset int = 0
	for {
		mu.Lock()
		*offset += 1
		localOffset = *offset
		mu.Unlock()

		if localOffset >= len(*lines) {
			break
		}
		var tmpElement string = (*lines)[localOffset]

		res, err := fetchGoogleThroughSocks5(tmpElement)
		if strings.Contains(res, "OK") && (err == nil) {
			fmt.Println(localOffset+1, " Checking: ", tmpElement, " GOOD")
			*outLines = append(*outLines, tmpElement)
		} else {
			fmt.Println(localOffset+1, " Checking: ", tmpElement, " BAD")
		}

	}
}

func doSyncCheck(lines []string, tCount int) []string {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var checkedList []string
	var listOffset int = -1

	for i := 1; i <= tCount; i++ {
		wg.Add(1)
		go workerThread(&listOffset, &mu, &lines, &wg, &checkedList)
	}
	wg.Wait()
	return checkedList
}

func fetchGoogleThroughSocks5(proxyAddr string) (string, error) {
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return "", fmt.Errorf("error socks5 dialer: %w", err)
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.(proxy.ContextDialer).DialContext(ctx, network, addr)
	}

	transport := &http.Transport{
		DialContext: dialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   7 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://fedoraproject.org/static/hotspot.txt", nil)
	if err != nil {
		return "", fmt.Errorf("error NewRequest: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error read clientDo: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error read tcp: %w", err)
	}

	return string(body), nil
}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, strings.Trim(scanner.Text(), "\r\n\t "))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func TouchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}
