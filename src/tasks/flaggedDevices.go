package tasks

import (
	"bytes"
	"html/template"
	"time"

	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"

	"gopkg.in/mail.v2"
)

var (
	mailServer  *mail.Dialer
	mailMessage *mail.Message

	deviceCache = make(map[string]time.Time)

	emailTemplate = template.Must(template.New("").Parse(`
The following flagged devices were detected on the network:

{{range .}}
==================================
MAC:       {{.MAC.String}}
Username:  {{.Username}}
Last Seen: {{.LastSeen.Format "2006-01-02 15:04"}}
==================================

{{end}}
`))
)

func setupMailDialer(c *common.Config) {
	mailServer = mail.NewDialer(
		c.Email.Address,
		c.Email.Port,
		c.Email.Username,
		c.Email.Password,
	)
}

func setupMailMessage(c *common.Config) {
	mailMessage = mail.NewMessage()
	mailMessage.SetHeader("From", c.Email.FromAddress)
	mailMessage.SetHeader("To", c.Email.ToAddresses...)
	mailMessage.SetHeader("Subject", "A flagged devices has been detected")
}

func flaggedDevicesTask(e *common.Environment) {
	if e.Config.Email.Address == "" {
		e.Log.Info("SMTP server not set, won't alert about flagged devices")
		return
	}

	setupMailDialer(e.Config)
	setupMailMessage(e.Config)

	for {
		time.Sleep(60 * time.Second)
		now := time.Now()

		// Get flagged devices
		flaggedDevices, err := stores.GetDeviceStore(e).GetFlaggedDevices()
		if err != nil {
			e.Log.WithField("error", err).Error("Failed to get flagged devices")
			continue
		}

		if len(flaggedDevices) == 0 {
			e.Log.Debug("No flagged devices, sleeping")
			continue
		}

		// Filter devices
		filtered := make([]*models.Device, 0, len(flaggedDevices)/2)

		for _, d := range flaggedDevices {
			cl := d.GetCurrentLease()

			// No current lease available
			if cl == nil {
				continue
			}

			// Current lease is expired
			if now.After(cl.GetEndTime()) {
				continue
			}

			// Already sent an alert on the device
			if t, exists := deviceCache[d.MAC.String()]; exists {
				if d.LastSeen.Equal(t) {
					continue
				}
			}

			filtered = append(filtered, d)
		}

		if len(filtered) == 0 {
			e.Log.Debug("No flagged devices after filtering, sleeping")
			continue
		}

		// Compose email template
		var buf bytes.Buffer
		if err := emailTemplate.Execute(&buf, filtered); err != nil {
			e.Log.WithField("error", err).Error("Failed to create email template")
			continue
		}

		// Send email
		mailMessage.SetBody("text/plain", buf.String())
		e.Log.Debug("Sending flagged device email")
		if err := mailServer.DialAndSend(mailMessage); err != nil {
			e.Log.WithField("error", err).Error("Failed sending flagged device alert")
			continue // Don't update cache for next try
		}

		// Update memory cache of alerted devices
		for _, d := range filtered {
			deviceCache[d.MAC.String()] = d.LastSeen
		}

		trimmap := make(map[string]bool, len(flaggedDevices))
		for _, d := range flaggedDevices {
			trimmap[d.MAC.String()] = true
		}

		for m := range deviceCache {
			if !trimmap[m] {
				delete(deviceCache, m)
			}
		}
	}
}
