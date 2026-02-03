package main

import (
	"github.com/gin-gonic/gin"
)

// user repesents data about a user
type user struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LastSeen string `json:"lastSeen"`
}

type userActivity struct {
	ID           string                    `json:"feedId"`
	Participants []userActivityParticipant `json:"participants"`
	ActionText   string                    `json:"actionText"`
}

type userActivityParticipant struct {
	Type UserType `json:"type"`
	ID   string   `json:"id"`
}
type UserType string

const (
	User UserType = "user"
)

var users = []user{
	{ID: "1", Name: "Alice", LastSeen: "2024-06-01T10:00:00Z"},
	{ID: "2", Name: "Bob", LastSeen: "2024-06-01T11:00:00Z"},
	{ID: "3", Name: "Charlie", LastSeen: "2024-06-01T12:00:00Z"},
}

var userActivities = []userActivity{
	{
		ID: "feed1",
		Participants: []userActivityParticipant{
			{Type: "user", ID: "1"},
			{Type: "user", ID: "2"},
		},
		ActionText: "Alice commented on Bob's post.",
	},
}

func getAllUsers(c *gin.Context) {
	c.IndentedJSON(200, users)
}

func getUserByID(c *gin.Context) {
	id := c.Param("id")
	for _, u := range users {
		if u.ID == id {
			c.IndentedJSON(200, u)
			return
		}
	}
	c.IndentedJSON(404, gin.H{"message": "user not found"})
}

func getUserActivities(c *gin.Context) {
	c.IndentedJSON(200, userActivities)
}

func getUserActivitiesByUserID(c *gin.Context) {
	id := c.Param("id")
	var activities []userActivity
	for _, activity := range userActivities {
		for _, participant := range activity.Participants {
			if participant.ID == id {
				activities = append(activities, activity)
				break
			}
		}
	}
	c.IndentedJSON(200, activities)
}

func main() {
	r := gin.Default()
	r.GET("/users", getAllUsers)
	r.GET("/users/:id", getUserByID)
	r.GET("/activities", getUserActivities)
	r.GET("/users/:id/activities", getUserActivitiesByUserID)
	r.Run()
}
