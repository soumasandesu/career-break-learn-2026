package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"charles/career-break-learn/user-service-golang/proto"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/protobuf/encoding/protojson"
)

var jsonMarshaler = protojson.MarshalOptions{
	UseProtoNames:   false,
	EmitUnpopulated: true,
}

var db *sql.DB

func initDB() {
	var err error
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "postgres"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Successfully connected to database")
}

func loadUsers() ([]*proto.User, error) {
	rows, err := db.Query("SELECT id, name, last_seen FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*proto.User
	for rows.Next() {
		var user proto.User
		var lastSeen sql.NullTime
		if err := rows.Scan(&user.Id, &user.Name, &lastSeen); err != nil {
			return nil, err
		}
		if lastSeen.Valid {
			user.LastSeen = lastSeen.Time.Format("2006-01-02T15:04:05Z")
		}
		users = append(users, &user)
	}
	return users, rows.Err()
}

func loadUserActivities() ([]*proto.UserActivity, error) {
	// Load activities
	activityRows, err := db.Query("SELECT feed_id, action_text_template FROM user_activities ORDER BY feed_id")
	if err != nil {
		return nil, err
	}
	defer activityRows.Close()

	var activities []*proto.UserActivity
	activityMap := make(map[string]*proto.UserActivity)

	for activityRows.Next() {
		var activity proto.UserActivity
		if err := activityRows.Scan(&activity.FeedId, &activity.ActionTextTemplate); err != nil {
			return nil, err
		}
		activityMap[activity.FeedId] = &activity
		activities = append(activities, &activity)
	}

	// Load subject referring
	subjectRows, err := db.Query(`
		SELECT feed_id, referring_type, referring_id, user_id 
		FROM user_activity_subject_referring 
		ORDER BY feed_id, referring_id
	`)
	if err != nil {
		return nil, err
	}
	defer subjectRows.Close()

	for subjectRows.Next() {
		var feedId, referringType, referringId string
		var userId sql.NullString
		if err := subjectRows.Scan(&feedId, &referringType, &referringId, &userId); err != nil {
			return nil, err
		}

		activity := activityMap[feedId]
		if activity == nil {
			continue
		}

		referring := &proto.UserActivityReferring{
			Id: referringId,
		}

		switch referringType {
		case "USER":
			referring.Type = proto.ReferringType_USER
		case "POST":
			referring.Type = proto.ReferringType_POST
		default:
			referring.Type = proto.ReferringType_USER
		}

		if userId.Valid {
			referring.UserId = userId.String
		}

		activity.SubjectReferring = append(activity.SubjectReferring, referring)
	}

	// Load object referring
	objectRows, err := db.Query(`
		SELECT feed_id, referring_type, referring_id, user_id 
		FROM user_activity_object_referring 
		ORDER BY feed_id, referring_id
	`)
	if err != nil {
		return nil, err
	}
	defer objectRows.Close()

	for objectRows.Next() {
		var feedId, referringType, referringId string
		var userId sql.NullString
		if err := objectRows.Scan(&feedId, &referringType, &referringId, &userId); err != nil {
			return nil, err
		}

		activity := activityMap[feedId]
		if activity == nil {
			continue
		}

		referring := &proto.UserActivityReferring{
			Id: referringId,
		}

		switch referringType {
		case "USER":
			referring.Type = proto.ReferringType_USER
		case "POST":
			referring.Type = proto.ReferringType_POST
		default:
			referring.Type = proto.ReferringType_USER
		}

		if userId.Valid {
			referring.UserId = userId.String
		}

		activity.ObjectReferring = append(activity.ObjectReferring, referring)
	}

	return activities, nil
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
	users, err := loadUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]map[string]interface{}, len(users))
	for i, user := range users {
		response[i] = protoUserToJSON(user)
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getUserByID(c *gin.Context) {
	id := c.Param("id")
	var user proto.User
	var lastSeen sql.NullTime
	err := db.QueryRow("SELECT id, name, last_seen FROM users WHERE id = $1", id).
		Scan(&user.Id, &user.Name, &lastSeen)
	if err == sql.ErrNoRows {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if lastSeen.Valid {
		user.LastSeen = lastSeen.Time.Format("2006-01-02T15:04:05Z")
	}
	c.IndentedJSON(http.StatusOK, protoUserToJSON(&user))
}

func getUserActivities(c *gin.Context) {
	activities, err := loadUserActivities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]map[string]interface{}, len(activities))
	for i, activity := range activities {
		response[i] = protoActivityToJSON(activity, nil)
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getUserActivitiesByUserID(c *gin.Context) {
	id := c.Param("id")
	activities, err := loadUserActivities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var filteredActivities []map[string]interface{}
	for _, activity := range activities {
		for _, subjectReferring := range activity.SubjectReferring {
			if subjectReferring.Id == id {
				filteredActivities = append(filteredActivities, protoActivityToJSON(activity, nil))
				goto nextActivity
			}
		}
		for _, objectReferring := range activity.ObjectReferring {
			if objectReferring.Id == id {
				filteredActivities = append(filteredActivities, protoActivityToJSON(activity, nil))
				goto nextActivity
			}
		}
	nextActivity:
	}
	c.IndentedJSON(http.StatusOK, filteredActivities)
}

func postUserActivity(c *gin.Context) {
	var jsonData map[string]interface{}
	if err := c.BindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	feedId := jsonData["feedId"].(string)
	actionTextTemplate := jsonData["actionTextTemplate"].(string)

	_, err = tx.Exec("INSERT INTO user_activities (feed_id, action_text_template) VALUES ($1, $2) ON CONFLICT (feed_id) DO UPDATE SET action_text_template = $2",
		feedId, actionTextTemplate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Clear existing referring
	_, err = tx.Exec("DELETE FROM user_activity_subject_referring WHERE feed_id = $1", feedId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, err = tx.Exec("DELETE FROM user_activity_object_referring WHERE feed_id = $1", feedId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert subject referring
	if subjectReferring, ok := jsonData["subjectReferring"].([]interface{}); ok {
		for _, p := range subjectReferring {
			subjectReferringItem := p.(map[string]interface{})
			referringType := "USER"
			if typeStr, ok := subjectReferringItem["type"].(string); ok {
				referringType = strings.ToUpper(typeStr)
			}
			referringId := subjectReferringItem["id"].(string)
			var userId sql.NullString
			if uid, ok := subjectReferringItem["userId"].(string); ok {
				userId = sql.NullString{String: uid, Valid: true}
			}

			_, err = tx.Exec("INSERT INTO user_activity_subject_referring (feed_id, referring_type, referring_id, user_id) VALUES ($1, $2, $3, $4)",
				feedId, referringType, referringId, userId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	// Insert object referring
	if objectReferring, ok := jsonData["objectReferring"].([]interface{}); ok {
		for _, p := range objectReferring {
			objectReferringItem := p.(map[string]interface{})
			referringType := "USER"
			if typeStr, ok := objectReferringItem["type"].(string); ok {
				referringType = strings.ToUpper(typeStr)
			}
			referringId := objectReferringItem["id"].(string)
			var userId sql.NullString
			if uid, ok := objectReferringItem["userId"].(string); ok {
				userId = sql.NullString{String: uid, Valid: true}
			}

			_, err = tx.Exec("INSERT INTO user_activity_object_referring (feed_id, referring_type, referring_id, user_id) VALUES ($1, $2, $3, $4)",
				feedId, referringType, referringId, userId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload the activity
	activities, err := loadUserActivities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, activity := range activities {
		if activity.FeedId == feedId {
			c.IndentedJSON(http.StatusCreated, protoActivityToJSON(activity, nil))
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload created activity"})
}

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()
	r.GET("/users", getAllUsers)
	r.GET("/users/:id", getUserByID)
	r.GET("/activities", getUserActivities)
	r.GET("/users/:id/activities", getUserActivitiesByUserID)
	r.POST("/users/-/activities", postUserActivity)
	r.Run()
}
