package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
	"os/exec"
)

func primary() {
	err := exec.Command("cmd", "/C", "start", "powershell", "go", "run", "C:/Users/phili/Documents/NTNU/Sanntids/exercise-7-philisg").Run()
	port := ":56456"
	pc, err := net.Dial("udp", port)
	if err != nil {
		fmt.Printf("An error occured while trying to send on UDP")
		return
	}
	
	
	count := 0
	for i := 0; i < 7; i++ {
		count += 1
		fmt.Fprint(pc, count)
		fmt.Println("counter: ",count)
		time.Sleep(time.Second * 1)
	}
	pc.Close()

}

func backup(wg *sync.WaitGroup) {
	err := exec.Command("cmd", "/C", "start", "powershell", "go", "run", "C:/Users/phili/Documents/NTNU/Sanntids/exercise-7-philisg").Run()


	UDPcounter := make(chan int)
	counter := 0
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
			fmt.Println("Received over UDP, Counter: ", counter)

		default:
			if time.Since(lastMsg) > 3*time.Second {
				isAlive = false
				break loop
			}
		}
	}

	fmt.Println("----------Backup-----------")
	go func() {
		wg.Add(1)
		backup(wg)
	}()

	pc, err := net.Dial("udp", port)
	if err != nil {
		fmt.Printf("An error occured while trying to send on UDP")
		return
	}
	count := counter
	for i := 0; i < 7; i++ {
		count += 1
		fmt.Fprint(pc, count)
		fmt.Println("backup counter: ", count)
		time.Sleep(time.Second * 1)
	}
	pc.Close()
	wg.Done()
} 

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		primary()
		wg.Done()
	}()

	go func() {
		backup(&wg)
	}()

	wg.Wait()
}
