package main

import (
	"fmt"
	"net/http"
	"strings"

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
		SubjectReferring: []*proto.UserActivityReferring{
			{Type: proto.ReferringType_USER, Id: "1"},
		},
		ObjectReferring: []*proto.UserActivityReferring{
			{Type: proto.ReferringType_POST, Id: "1024", UserId: "2"},
		},
		ActionTextTemplate: "{subject} commented on {object} post.",
	},
}

func protoUserToJSON(user *proto.User) map[string]interface{} {
	return map[string]interface{}{
		"id":       user.Id,
		"name":     user.Name,
		"lastSeen": user.LastSeen,
	}
}

func protoActivityToJSON(activity *proto.UserActivity, you *proto.User) map[string]interface{} {
	subjectReferring := make([]map[string]interface{}, len(activity.SubjectReferring))
	for i, p := range activity.SubjectReferring {
		subjectReferring[i] = map[string]interface{}{
			"type": p.Type.String(),
			"id":   p.Id,
		}
	}
	var subjectText string
	if len(activity.SubjectReferring) > 1 {
		subjectText = "%s and %d others"
	} else if activity.SubjectReferring[0].Id == you.Id {
		subjectText = "%s"
	} else {
		subjectText = "you"
	}
	subjectText = fmt.Sprintf(subjectText, activity.SubjectReferring[len(activity.SubjectReferring)-1].Id)

	objectReferring := make([]map[string]interface{}, len(activity.ObjectReferring))
	for i, p := range activity.ObjectReferring {
		objectReferring[i] = map[string]interface{}{
			"type": p.Type.String(),
			"id":   p.Id,
		}
	}
	var objectText string
	if len(activity.ObjectReferring) > 1 {
		objectText = "%s and %d others"
	} else if activity.ObjectReferring[0].Id == you.Id {
		objectText = "%s"
	} else {
		subjectText = "you"
	}
	objectText = fmt.Sprintf(objectText, activity.ObjectReferring[len(activity.ObjectReferring)-1].Id)

	actionText := strings.ReplaceAll(activity.ActionTextTemplate, "{subject}", you.Name)
	actionText = strings.ReplaceAll(actionText, "{object}", you.Name)
	if len(actionText) > 0 {
		actionText = strings.ToUpper(actionText[:1]) + actionText[1:]
	}

	return map[string]interface{}{
		"feedId":             activity.FeedId,
		"subjectReferring":   subjectReferring,
		"objectReferring":    objectReferring,
		"actionTextTemplate": activity.ActionTextTemplate,
		"actionText":         actionText,
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
		response[i] = protoActivityToJSON(activity, nil)
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getUserActivitiesByUserID(c *gin.Context) {
	id := c.Param("id")
	var activities []map[string]interface{}
	for _, activity := range userActivities {
		for _, subjectReferring := range activity.SubjectReferring {
			if subjectReferring.Id == id {
				activities = append(activities, protoActivityToJSON(activity, nil))
				break
			}
		}
		for _, objectReferring := range activity.ObjectReferring {
			if objectReferring.Id == id {
				activities = append(activities, protoActivityToJSON(activity, nil))
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
		FeedId:             jsonData["feedId"].(string),
		ActionTextTemplate: jsonData["actionTextTemplate"].(string),
	}

	if subjectReferring, ok := jsonData["subjectReferring"].([]interface{}); ok {
		for _, p := range subjectReferring {
			subjectReferring := p.(map[string]interface{})
			referringType := proto.ReferringType_USER
			if typeStr, ok := subjectReferring["type"].(string); ok && typeStr == "user" {
				referringType = proto.ReferringType_USER
			}
			newActivity.SubjectReferring = append(newActivity.SubjectReferring, &proto.UserActivityReferring{
				Type: referringType,
				Id:   subjectReferring["id"].(string),
			})
		}
	}

	if objectReferring, ok := jsonData["objectReferring"].([]interface{}); ok {
		for _, p := range objectReferring {
			objectReferringItem := p.(map[string]interface{})
			referringType := proto.ReferringType_USER
			if typeStr, ok := objectReferringItem["type"].(string); ok && typeStr == "user" {
				referringType = proto.ReferringType_USER
			}
			newActivity.ObjectReferring = append(newActivity.ObjectReferring, &proto.UserActivityReferring{
				Type: referringType,
				Id:   objectReferringItem["id"].(string),
			})
		}
	}

	userActivities = append(userActivities, newActivity)
	c.IndentedJSON(http.StatusCreated, protoActivityToJSON(newActivity, nil))
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
