package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/go-github/github"
)

var owner string = "binary-com"
var repo string = "deriv-app"

func main() {
	client := InitClient()

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
		prId := 3463

		pr, _, err := client.PullRequests.Get(ctx, owner, repo, prId)
		if err != nil {
			panic(err)
		}

		mx.Lock()
		prs = append(prs, pr)
		mx.Unlock()
	}

	for _, notf := range notfs {
		if notf.GetReason() == "review_requested" {
			go fetchPR(notf.GetSubject().GetURL())
		}
	}

	wg.Wait()

	RenderTable(prs)
}
