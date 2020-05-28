package main

import (
	"fmt"
	"os"
	"peerdel/ripper"
	"sync"
)

func main()  {

	lines, err := ripper.ParseFileToLines(os.Args[1])
	if err != nil {
		panic("cant extract message line")
	}

	var wg sync.WaitGroup

	for _, line := range lines {
		wg.Add(1)
		cid, err := ripper.ParseCID(line)
		if err != err {
			panic(err)
		}

		ip, err := ripper.PeerIdToIP(ripper.ParsePeerId(line))
		if err != nil {
			panic(err)
		}

		fmt.Println(fmt.Sprintf("Deleting: %s from: %s", cid, ip.String()))
		go ripper.UnPinAndDelete(&wg, ip, cid)
	}

	wg.Wait()
	fmt.Println("Done deleting")
}
