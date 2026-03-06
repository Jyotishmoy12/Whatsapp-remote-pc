package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/joho/godotenv"
	"github.com/jyotishmoy12/whatsapp-remote-pc/commands"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

func eventHandler(client *whatsmeow.Client) whatsmeow.EventHandler {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if client.Store.ID == nil {
				return
			}
			_ = godotenv.Load()
			rawAllowed := os.Getenv("AUTHORIZED_IDS")
			allowedList := strings.Split(rawAllowed, ",")
			senderUser := v.Info.Sender.User
			isAuthorized := false
			for _, id := range allowedList {
				if senderUser == strings.TrimSpace(id) {
					isAuthorized = true
					break
				}
			}

			if !isAuthorized {
				fmt.Printf("Blocked unauthorized access from: %s\n", senderUser)
				return
			}
			msgText := v.Message.GetConversation()
			if msgText == "" {
				return
			}

			fmt.Printf("Authenticated Command: %s\n", msgText)
			reply := commands.HandleCommand(msgText)

			if reply == "ACTION_SCREENSHOT" {
				path, err := commands.CaptureScreen()
				if err != nil {
					client.SendMessage(context.Background(), v.Info.Sender, &waE2E.Message{
						Conversation: proto.String("Failed: " + err.Error()),
					})
					return
				}
				data, _ := os.ReadFile(path)
				resp, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
				if err != nil {
					fmt.Printf("Upload error: %v\n", err)
					return
				}
				client.SendMessage(context.Background(), v.Info.Sender, &waE2E.Message{
					ImageMessage: &waE2E.ImageMessage{
						URL:           &resp.URL,
						DirectPath:    &resp.DirectPath,
						MediaKey:      resp.MediaKey,
						Mimetype:      proto.String("image/png"),
						FileEncSHA256: resp.FileEncSHA256,
						FileSHA256:    resp.FileSHA256,
						FileLength:    &resp.FileLength,
					},
				})
				os.Remove(path)

			} else if reply != "" {
				client.SendMessage(context.Background(), v.Info.Sender, &waE2E.Message{
					Conversation: proto.String(reply),
				})
			}
			if strings.HasPrefix(reply, "ACTION_FETCH_FILE") {
				fileName := strings.TrimPrefix(msgText, "!get ")
				data, err := os.ReadFile(fileName)
				if err != nil {
					client.SendMessage(context.Background(), v.Info.Sender, &waE2E.Message{
						Conversation: proto.String("File not found: " + fileName),
					})
					return
				}
				resp, _ := client.Upload(context.Background(), data, whatsmeow.MediaDocument)

				client.SendMessage(context.Background(), v.Info.Sender, &waE2E.Message{
					DocumentMessage: &waE2E.DocumentMessage{
						URL:           &resp.URL,
						DirectPath:    &resp.DirectPath,
						MediaKey:      resp.MediaKey,
						Mimetype:      proto.String("application/octet-stream"),
						FileName:      &fileName,
						FileEncSHA256: resp.FileEncSHA256,
						FileSHA256:    resp.FileSHA256,
						FileLength:    &resp.FileLength,
					},
				})
			}
			if reply == "ACTION_LIST_FILES" {
				path := strings.TrimSpace(strings.TrimPrefix(msgText, "!ls"))
				if path == "" {
					path = commands.CurrentWorkDir
				}

				files, err := os.ReadDir(path)
				if err != nil {
					client.SendMessage(context.Background(), v.Info.Sender, &waE2E.Message{
						Conversation: proto.String("Error: " + err.Error()),
					})
					return
				}

				output := fmt.Sprintf("Contents of %s:\n\n", path)
				for _, file := range files {
					icon := "📄"
					if file.IsDir() {
						icon = "📁"
					}
					output += fmt.Sprintf("%s %s\n", icon, file.Name())
				}

				client.SendMessage(context.Background(), v.Info.Sender, &waE2E.Message{
					Conversation: proto.String(output),
				})
			}
			if reply == "ACTION_CHANGE_DIR" {
				newDir := strings.TrimPrefix(msgText, "!cd ")
				err := os.Chdir(newDir)
				if err == nil {
					commands.CurrentWorkDir, _ = os.Getwd()
					reply = "Now in: " + commands.CurrentWorkDir
				} else {
					reply = "Could not move: " + err.Error()
				}
			}
			if reply == "ACTION_FIND_FILE" {
				fileName := strings.TrimPrefix(msgText, "!find ")
				fileName = strings.TrimSpace(fileName)
				results, err := commands.FindFile(commands.CurrentWorkDir, fileName)

				var finalReply string
				if err != nil {
					finalReply = "Search error: " + err.Error()
				} else if len(results) == 0 {
					finalReply = "No matches found for: " + fileName
				} else {
					finalReply = fmt.Sprintf("🔍 Found %d match(es):\n", len(results))
					for _, path := range results {
						finalReply += "\n📍 " + path
					}
					finalReply += "\n\nTip: Use !get [path] to download."
				}
				client.SendMessage(context.Background(), v.Info.Sender, &waE2E.Message{
					Conversation: proto.String(finalReply),
				})
				return
			}
			if reply == "ACTION_HARD_RESET" {
				fmt.Println("!!! NUCLEAR RESET TRIGGERED !!!")

				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				if client.Store != nil && client.Store.ID != nil {
					err := client.Store.Delete(ctx)
					if err != nil {
						fmt.Printf("Store delete error: %v\n", err)
					}
				}
				_ = client.Logout(ctx)
				client.Disconnect()
				time.Sleep(1 * time.Second)
				filesToDelete := []string{"session.db", "scan_me.png", "current_screen.png"}
				for _, file := range filesToDelete {
					_ = os.Remove(file)
				}

				fmt.Println("Clean wipe completed. Exiting.")
				os.Exit(0)
			}
			if reply == "ACTION_EXECUTE_CMD" {
				commandStr := strings.TrimPrefix(msgText, "!cmd ")
				cmd := exec.Command("cmd", "/C", commandStr)
				cmd.Dir = commands.CurrentWorkDir
				output, err := cmd.CombinedOutput()

				finalReply := string(output)
				if err != nil && finalReply == "" {
					finalReply = "Execution Error: " + err.Error()
				} else if finalReply == "" {
					finalReply = "Command finished (no output)."
				}
				client.SendMessage(context.Background(), v.Info.Sender, &waE2E.Message{
					Conversation: proto.String(finalReply),
				})
				return
			}
			if reply == "ACTION_SHUTDOWN" {
				client.SendMessage(context.Background(), v.Info.Sender, &waE2E.Message{
					Conversation: proto.String("🔌 Shutting down PC in 5 seconds... Goodbye!"),
				})
				time.Sleep(5 * time.Second)
				exec.Command("shutdown", "/s", "/t", "0").Run()
			}

			if reply == "ACTION_RESTART" {
				client.SendMessage(context.Background(), v.Info.Sender, &waE2E.Message{
					Conversation: proto.String("🔄 Restarting PC... I'll be back online in a minute."),
				})
				time.Sleep(5 * time.Second)
				exec.Command("shutdown", "/r", "/t", "0").Run()
			}
		}
	}
}

func main() {
	ctx := context.Background()
	dbLog := waLog.Stdout("Database", "INFO", true)
	clientLog := waLog.Stdout("Client", "INFO", true)
	dbPath := "file:session.db?_pragma=foreign_keys(1)"
	container, err := sqlstore.New(ctx, "sqlite", dbPath, dbLog)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler(client))
	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(ctx)
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrFile := "scan_me.png"
				qrcode.WriteFile(evt.Code, qrcode.Medium, 256, qrFile)
				fmt.Printf("\n[ACTION] New QR Code generated: Open '%s' and scan it.\n", qrFile)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		fmt.Println(">> Successfully Auto-Connected!")
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Shutting down...")
	client.Disconnect()
}
