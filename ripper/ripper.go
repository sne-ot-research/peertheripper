package ripper

import (
	"bufio"
	"errors"
	"fmt"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	mh "github.com/multiformats/go-multihash"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	strings "strings"
	"sync"
)

func ParsePeerId(line string) string  {
	peerIdRegex := regexp.MustCompile(`peer\.ID=(\w)+`)
	peerId := strings.Split(peerIdRegex.FindString(line), "=")[1]
	return peerId
}

func ParseCID(line string) (string, error)  {

	cidRegex := regexp.MustCompile(`uint8=\[(.)+\]`)
	cidByteString := strings.Split(cidRegex.FindString(line), "=")[1]

	toByte, err := parseByteStringToByte(cidByteString)
	if err != nil {
		return "", err
	}

	cast, err := mh.Cast(toByte)

	if err != nil {
		return "", err
	}

	return cast.B58String(), nil
}


func ParseFileToLines(file string) ([]string, error) {

	jsonFile, err := os.Open(file)
	defer jsonFile.Close()


	if err != nil {
		return nil, err
	}

	var lines []string

	scanner := bufio.NewScanner(jsonFile)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "adding provider") {
			lines = append(lines, text)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines, nil
}

func parseByteStringToByte(byteString string) ([]byte, error) {

	var bb []byte
	for _, ps := range strings.Split(strings.Trim(byteString, "[]"), " ") {
		pi,_ := strconv.Atoi(ps)
		bb = append(bb,byte(pi))
	}

	return bb, nil
}

func PeerIdToIP(peerId string) (net.IP, error) {
	peers, err := exec.Command("ipfs", "swarm", "peers").Output()
	if err != nil {
		log.Fatal(err)
	}

	peersStrs := strings.Split(string(peers), "\n")

	for _, p := range peersStrs {
		if p == "" {
			continue
		}

		if strings.Contains(p, peerId) {
			multiaddr, err := ma.NewMultiaddr(p)
			if err != nil {
				log.Fatal(err)
				return nil , err
			}

			ip, err := manet.ToIP(multiaddr)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
			return ip, nil
		}

	}
	
	return nil, errors.New("no peer found")
}

func UnPinAndDelete(wg *sync.WaitGroup, ip net.IP, cid string)  {
	defer wg.Done()
	unpinurl := fmt.Sprintf("http://%s:%d/api/v0/pin/rm?arg=/ipfs/%s", ip.String(), 5001, cid)
	delurl := fmt.Sprintf("http://%s:%d/api/v0/block/rm?arg=%s&force=true", ip.String(), 5001, cid)

	http.Post(unpinurl, "", nil)
	http.Post(delurl, "", nil)
}