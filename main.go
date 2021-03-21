package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
	"os/exec"
	"flag"
)

func primary(count int) {
	fmt.Println("Primary up and running!")
	port := ":56456"
	pc, err := net.Dial("udp", port)
	if err != nil {
		fmt.Printf("An error occured while trying to send on UDP")
		return
	}
	
	for i := 0; i < 5; i++ {
		count += 1
		fmt.Fprint(pc, count)
		fmt.Println("counter: ",count)
		time.Sleep(time.Second * 1)
	}
	pc.Close()
}

func backup(wg *sync.WaitGroup) {
	UDPcounter := make(chan int)
	counter := 3
	port := ":56456"
	var isAlive bool = true 

	go func() {
		pc, err := net.ListenPacket("udp", port)
		if err != nil {
			fmt.Println("something went wrong with listening for UDP packets!")
		}
		
		buf := make([]byte, 1024)
		go func() {
			for {
				n, _, err := pc.ReadFrom(buf)
				if err != nil {
					fmt.Println("error receiving UDP, Sender may be dead :(")
					return
				}
				byteToInt, _ := strconv.Atoi(string(buf[:n]))
				UDPcounter <- byteToInt
			}
		}()
		for(isAlive){}
		pc.Close()
	}()

 	lastMsg := time.Now()

loop:
	for {
		select {
		case c := <-UDPcounter:
			lastMsg = time.Now()
			counter = c
			//fmt.Println("Received over UDP, Counter: ", counter) //for debuging
		default:
			if time.Since(lastMsg) > 2*time.Second {
				isAlive = false
				break loop
			}
		}
	}

	fmt.Println("Timeout reached.")
	fmt.Println("Starting backup!")
	strcounter := strconv.Itoa(counter)
	exec.Command("cmd", "/C", "start", "powershell", "go", "run", "C:/Users/phili/Documents/NTNU/Sanntids/exercise-7-philisg/main.go", "-count="+strcounter).Run()
	wg.Done()
} 

func main() {
	CountFromArg := flag.String("count","0","The counter start from this number")
	flag.Parse()
	startCounter, _ := strconv.Atoi(*CountFromArg)
	//fmt.Println("From command line: ", startCounter) //for debuging
	var wg sync.WaitGroup
	wg.Add(2)

	go func(count int) {
		primary(count)
		wg.Done()
	}(startCounter)

	go func() {
		backup(&wg)
	}()

	wg.Wait()
}
