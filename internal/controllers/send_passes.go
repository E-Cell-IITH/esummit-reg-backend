package controllers

import (
	"fmt"
	"reg/internal/database"
	email "reg/internal/emails"

	"github.com/gin-gonic/gin"
)

func SendPassesHandler(c *gin.Context) {
	userTickets := database.GetIDsForPasses()

	for _, userTicket := range userTickets {
		// send email with pass
		fmt.Println(userTicket)
		data, err := email.LoadPassEmailTemplate(userTicket.Name, userTicket.TicketTitle, userTicket.UID)
		if err != nil {
			// handle the error
			fmt.Println(err)
			continue
		}

		ok, err := email.SendPASSEmail(userTicket.Email, nil, "Your E-Summit 2025 Pass & Event Schedule Are Here!", data, "", userTicket.UID)

		if !ok {
			// handle the error
			fmt.Printf("Failed to send email to %s, ERR: %s\n", userTicket.Email, err)
			continue
		}

	}

	c.JSON(200, gin.H{"message": "Passes sent successfully"})

}
