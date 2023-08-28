package notification

import (
	"fmt"
	"os/exec"
)

func SendNotification(headline, message string) error {
	cmd := exec.Command("notify-send", headline, message, "-t", "0")

	err := cmd.Run()
	if err != nil {
		return err
	}
	fmt.Println("Notification sent: ", headline, message)

	return nil
}
