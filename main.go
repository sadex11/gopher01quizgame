package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

type AnswerCounter struct {
	answerCount int
	lock        sync.Mutex
}

type TestRecord struct {
	question string
	answer   string
}

func (c *AnswerCounter) AddAnswer() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.answerCount++
}

func (c *AnswerCounter) GetCount() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.answerCount
}

func loadQuestions(path string) []TestRecord {
	f, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	readed, err := csvReader.ReadAll()

	if err != nil {
		log.Fatal(err)
	}

	testRecords := []TestRecord{}

	for _, val := range readed {
		if len(val) != 2 {
			log.Println("Error line", val)
		} else {
			testRecords = append(testRecords, TestRecord{val[0], val[1]})
		}
	}

	return testRecords
}

func runTest(records []TestRecord, counter *AnswerCounter, responseChan chan<- bool) {
	for i, record := range records {
		fmt.Printf("Question %d: what is %v?\n", i+1, record.question)
		var userAnswer string

		fmt.Scanln(&userAnswer)
		userAnswer = strings.Trim(userAnswer, " ")
		userAnswer = strings.ToLower(userAnswer)

		if userAnswer == record.answer {
			counter.AddAnswer()
		}
	}

	responseChan <- true
}

func printResult(recordCount int, succes bool, counter *AnswerCounter) {
	if succes {
		fmt.Println("Finished within timeout!")
	} else {
		fmt.Println("Timeout expired!")
	}

	fmt.Println("Get", counter.GetCount(), "correct answers from", recordCount, "questions!")
}

func main() {
	questionFilePath := flag.String("questions", "problems.csv", "A path to a file with questions")
	timeout := flag.Int("timeout", 30, "Test timeout")
	shuffle := flag.Bool("shuffle", true, "Shuffle the questions")
	flag.Parse()

	records := loadQuestions(*questionFilePath)

	if len(records) == 0 {
		log.Fatal("Empty test records")
	}

	if *shuffle {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(records), func(i, j int) { records[i], records[j] = records[j], records[i] })
	}

	resultChan := make(chan bool)

	fmt.Println("Please press Enter to start the test")
	fmt.Scanln()

	timer := time.NewTimer(time.Duration(*timeout) * time.Second)
	counter := AnswerCounter{}

	go runTest(records, &counter, resultChan)

	var succes bool

	select {
	case <-resultChan:
		succes = true
		break
	case <-timer.C:
		succes = false
		break
	}

	printResult(len(records), succes, &counter)
}
