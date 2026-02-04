package main

import (
	"net/http"

	"charles/career-break-learn/user-service-golang/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

var jsonMarshaler = protojson.MarshalOptions{
	UseProtoNames:   false,
	EmitUnpopulated: true,
}

var users = []*proto.User{
	{Id: "1", Name: "Alice", LastSeen: "2024-06-01T10:00:00Z"},
	{Id: "2", Name: "Bob", LastSeen: "2024-06-01T11:00:00Z"},
	{Id: "3", Name: "Charlie", LastSeen: "2024-06-01T12:00:00Z"},
}

var userActivities = []*proto.UserActivity{
	{
		FeedId: "feed1",
		Participants: []*proto.UserActivityParticipant{
			{Type: proto.UserType_USER, Id: "1"},
			{Type: proto.UserType_USER, Id: "2"},
		},
		ActionText: "Alice commented on Bob's post.",
	},
}

func protoUserToJSON(user *proto.User) map[string]interface{} {
	return map[string]interface{}{
		"id":       user.Id,
		"name":     user.Name,
		"lastSeen": user.LastSeen,
	}
}

func protoActivityToJSON(activity *proto.UserActivity) map[string]interface{} {
	participants := make([]map[string]interface{}, len(activity.Participants))
	for i, p := range activity.Participants {
		participants[i] = map[string]interface{}{
			"type": p.Type.String(),
			"id":   p.Id,
		}
	}
	return map[string]interface{}{
		"feedId":       activity.FeedId,
		"participants": participants,
		"actionText":   activity.ActionText,
	}
}

func getAllUsers(c *gin.Context) {
	response := make([]map[string]interface{}, len(users))
	for i, user := range users {
		response[i] = protoUserToJSON(user)
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getUserByID(c *gin.Context) {
	id := c.Param("id")
	for _, u := range users {
		if u.Id == id {
			c.IndentedJSON(http.StatusOK, protoUserToJSON(u))
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
}

func getUserActivities(c *gin.Context) {
	response := make([]map[string]interface{}, len(userActivities))
	for i, activity := range userActivities {
		response[i] = protoActivityToJSON(activity)
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getUserActivitiesByUserID(c *gin.Context) {
	id := c.Param("id")
	var activities []map[string]interface{}
	for _, activity := range userActivities {
		for _, participant := range activity.Participants {
			if participant.Id == id {
				activities = append(activities, protoActivityToJSON(activity))
				break
			}
		}
	}
	c.IndentedJSON(http.StatusOK, activities)
}

func postUserActivity(c *gin.Context) {
	var jsonData map[string]interface{}
	if err := c.BindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newActivity := &proto.UserActivity{
		FeedId:     jsonData["feedId"].(string),
		ActionText: jsonData["actionText"].(string),
	}

	if participants, ok := jsonData["participants"].([]interface{}); ok {
		for _, p := range participants {
			participant := p.(map[string]interface{})
			userType := proto.UserType_USER
			if typeStr, ok := participant["type"].(string); ok && typeStr == "user" {
				userType = proto.UserType_USER
			}
			newActivity.Participants = append(newActivity.Participants, &proto.UserActivityParticipant{
				Type: userType,
				Id:   participant["id"].(string),
			})
		}
	}

	userActivities = append(userActivities, newActivity)
	c.IndentedJSON(http.StatusCreated, protoActivityToJSON(newActivity))
}

func main() {
	r := gin.Default()
	r.GET("/users", getAllUsers)
	r.GET("/users/:id", getUserByID)
	r.GET("/activities", getUserActivities)
	r.GET("/users/:id/activities", getUserActivitiesByUserID)
	r.POST("/users/-/activities", postUserActivity)
	r.Run()
}
