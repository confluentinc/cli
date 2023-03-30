package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func okta(url, username, password string) string {
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithLogf(log.Printf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var token string

	actions := []chromedp.Action{
		chromedp.Navigate(url),
		chromedp.WaitVisible(`//input[@name="username"]`),
		chromedp.SendKeys(`//input[@id="okta-signin-username"]`, username),
		chromedp.SendKeys(`//input[@id="okta-signin-password"]`, password),
		chromedp.Click(`//input[@id="okta-signin-submit"]`),
		chromedp.WaitVisible(`//div[@id="cc-root"]`),
		chromedp.Text(`//div[@id="token"]`, &token),
	}

	_ = chromedp.Run(ctx, actions...)

	return token
}
