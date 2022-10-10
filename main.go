package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	gosxnotifier "github.com/deckarep/gosx-notifier"
	"github.com/google/go-github/github"
)

var owner string = "binary-com"
var repo string = "deriv-app"

func Notify(msg string) {
	note := gosxnotifier.NewNotification(msg)
	note.Title = "yay"
	note.Sound = gosxnotifier.Basso
	note.AppIcon = "gopher.png"
	note.ContentImage = "gopher.png"

	err := note.Push()

	//If necessary, check error
	if err != nil {
		log.Println("Uh oh!")
	}
}

func main() {
	client := InitClient()

	Notify("lmaoo")

	ctx := context.Background()
	opt := &github.NotificationListOptions{
		All: true,
	}
	notfs, _, err := client.Activity.ListNotifications(ctx, opt)
	if err != nil {
		fmt.Printf("Could not list notifications: %v\n", err)
	}

	prs := make([]*github.PullRequest, 0)

	var wg sync.WaitGroup
	var mx sync.Mutex

	fetchPR := func(urlStr string) {
		defer wg.Done()
		wg.Add(1)
		prId := 6719

		pr, _, err := client.PullRequests.Get(ctx, owner, repo, prId)
		if err != nil {
			panic(err)
		}

		mx.Lock()
		prs = append(prs, pr)
		mx.Unlock()
	}

	for _, notf := range notfs {
		go fetchPR(notf.GetSubject().GetURL())
		// if notf.GetReason() == "review_requested" {
		// 	go fetchPR(notf.GetSubject().GetURL())
		// }
	}

	wg.Wait()

	RenderTable(prs)
}
