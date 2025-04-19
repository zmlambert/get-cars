package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Car struct {
	Number int
	Name   string
}

type SemaphoredWaitGroup struct {
	sem chan bool
	wg  sync.WaitGroup
}

func (s *SemaphoredWaitGroup) Add(delta int) {
	s.wg.Add(delta)
	s.sem <- true
}
func (s *SemaphoredWaitGroup) Done() {
	<-s.sem
	s.wg.Done()
}
func (s *SemaphoredWaitGroup) Wait() {
	s.wg.Wait()
}

func downloadCar(car Car, carFolder string) {
	resp, err := http.Get(fmt.Sprintf("https://awesomecars.neocities.org/ver2/%d.mp4", car.Number))
	if err != nil {
		fmt.Printf("couldn't get vid for car %d\n", car.Number)
		return
	}

	videoBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading video for %d\n", car.Number)
	}
	filename := fmt.Sprintf("%s/%d_%s.mp4", carFolder, car.Number, car.Name)
	_, err = os.Stat(filename)
	if !os.IsNotExist(err) {
		fmt.Printf("\tfile already exists - %d\r", car.Number)
		return
	}
	os.WriteFile(filename, videoBytes, os.FileMode(int(0644)))
	fmt.Printf("downloaded %d\n", car.Number)
}

func main() {
	carFolder := filepath.Join(".", "cars_v4")
	err := os.MkdirAll(carFolder, os.ModePerm)
	if err != nil {
		fmt.Println("error creating folder for cars")
		os.Exit(1)
	}
	// get the index.html
	resp, err := http.Get("https://awesomecars.neocities.org/")
	if err != nil {
		fmt.Printf("error making request: %s\n", err)
		os.Exit(1)
	}

	// alright so this is scuffed but the site's
	// code is even more scuffed lmao.

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading response body")
		os.Exit(2)
	}

	startOfList := "titles = ["
	endOfList := "// spaghetti code never changes"

	bodyString := string(bodyBytes)

	startIndex := strings.Index(bodyString, startOfList)
	endIndex := strings.Index(bodyString, endOfList)

	listOfCars := bodyString[startIndex+len(startOfList) : endIndex]
	splitListOfCars := strings.Split(listOfCars, "\n")
	fmt.Printf("Initiating download of %d cars...\n", len(splitListOfCars))

	// create list of cars to download
	var cars []Car

	for _, line := range splitListOfCars {
		line = strings.ReplaceAll(line, "\"", "")
		line = strings.ReplaceAll(line, ",", "")
		line = strings.ReplaceAll(line, "#", "")
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, " - ", 2)

		if len(parts) != 2 {
			fmt.Printf("error on car \"%s\"\n", line)
			continue
		}

		numberString := parts[0]
		name := strings.ReplaceAll(parts[1], " ", "_")

		number, err := strconv.Atoi(numberString)
		if err != nil {
			fmt.Printf("error converting from string to int: \"%s\"\n", numberString)
			continue
		}
		cars = append(cars, Car{Name: name, Number: number})
	}

	// goroutine(s)
	wg := SemaphoredWaitGroup{sem: make(chan bool, 12)}

	for _, car := range cars {
		wg.Add(1)
		go func(car Car) {
			defer wg.Done()
			downloadCar(car, carFolder)
		}(car)
	}

	wg.Wait()
}
