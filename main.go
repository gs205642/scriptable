package main

import (
	"net/url"
	"os"
	"os/exec"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func main() {
	app := fiber.New(fiber.Config{})

	app.Post("/upload", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(Response{false, err.Error()})
		}

		filePath := path.Join("scripts", file.Filename)
		c.SaveFile(file, filePath)

		return c.JSON(Response{true, filePath})
	})

	app.Get("/sync/*", func(c *fiber.Ctx) error {
		filePath, err := url.QueryUnescape(c.Params("*"))
		if err != nil {
			return c.JSON(Response{false, err.Error()})
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return c.JSON(Response{false, "文件不存在：" + filePath})
		}

		if err := exec.Command("code", filePath).Run(); err != nil {
			return c.JSON(Response{false, err.Error()})
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return c.JSON(Response{false, err.Error()})
		}
		defer watcher.Close()

		done := make(chan bool)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Has(fsnotify.Write) {
						done <- true
						return
					}

				case err, ok := <-watcher.Errors:
					if !ok {
						done <- true
						return
					}
					log.Fatal(err)
				}
			}
		}()

		err = watcher.Add(filePath)
		if err != nil {
			log.Fatal(err)
		}

		<-done

		message, err := os.ReadFile(filePath)
		if err != nil {
			return c.JSON(Response{false, err.Error()})
		}
		return c.JSON(Response{true, string(message)})
	})

	app.Post("/console", func(c *fiber.Ctx) error {
		type Log struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		}
		l := new(Log)
		if err := c.BodyParser(l); err != nil {
			log.Error(err)
		}
		switch l.Type {
		case "warn":
			log.Warn(l.Message)
		case "error":
			log.Error(l.Message)
		default:
			log.Info(l.Message)
		}
		return c.SendStatus(201)
	})

	app.Listen(":8080")
}
