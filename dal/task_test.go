package dal

import (
	_ "github.com/lib/pq"
	"testing"
)

func newTaskForTest(t *testing.T) *Task {
	return NewTask(newDbForTest(t))
}

func TestTaskCRUD(t *testing.T) {
	u := newUserForTest(t)

	// Signup
	userRow, err := u.Signup(nil, newEmailForTest(), "abc123", "abc123")
	if err != nil {
		t.Errorf("Signing up user should work. Error: %v", err)
	}
	if userRow == nil {
		t.Fatal("Signing up user should work.")
	}
	if userRow.ID <= 0 {
		t.Fatal("Signing up user should work.")
	}

	cl := newClusterForTest(t)

	// Create cluster for user
	clusterRow, err := cl.Create(nil, userRow.ID, "cluster-name")
	if err != nil {
		t.Fatalf("Creating a cluster for user should work. Error: %v", err)
	}
	if clusterRow.ID <= 0 {
		t.Fatalf("Cluster ID should be assign properly. clusterRow.ID: %v", clusterRow.ID)
	}

	at := newAccessTokenForTest(t)

	// Create access token
	tokenRow, err := at.Create(nil, userRow.ID, clusterRow.ID, "execute")
	if err != nil {
		t.Fatalf("Creating a token should work. Error: %v", err)
	}
	if tokenRow.ID <= 0 {
		t.Fatalf("AccessToken ID should be assign properly. tokenRow.ID: %v", tokenRow.ID)
	}

	task := newTaskForTest(t)

	// Create a new task
	taskRow, err := task.Create(nil, tokenRow.ID, "true", "0 30 * * * *")
	if err != nil {
		t.Fatalf("Failed to create task. Error: %v", err)
	}

	// Update existing task
	data := make(map[string]interface{})
	data["user_id"] = tokenRow.UserID
	data["query"] = "false"
	data["cron"] = "1 30 * * * *"

	_, err = task.UpdateByID(nil, data, taskRow.ID)
	if err != nil {
		t.Fatalf("Failed to update task. Error: %v", err)
	}

	// SELECT FROM tasks
	taskRowFromDb, err := task.GetByAccessTokenQueryAndCron(nil, tokenRow, "false", "1 30 * * * *")
	if err != nil {
		t.Fatalf("Failed to get task. Error: %v", err)
	}
	if taskRow.ID != taskRowFromDb.ID {
		t.Fatalf("Failed to get the correct task.")
	}
	if taskRowFromDb.Query != "false" {
		t.Fatalf("Update did not work correctly. Query: %v", taskRowFromDb.Query)
	}

	// DELETE FROM tasks WHERE id=...
	_, err = task.DeleteById(nil, taskRowFromDb.ID)
	if err != nil {
		t.Fatalf("Deleting tasks by id should not fail. Error: %v", err)
	}

	// DELETE FROM access_tokens WHERE id=...
	_, err = at.DeleteById(nil, tokenRow.ID)
	if err != nil {
		t.Fatalf("Deleting access_tokens by id should not fail. Error: %v", err)
	}

	// DELETE FROM clusters WHERE id=...
	_, err = cl.DeleteById(nil, clusterRow.ID)
	if err != nil {
		t.Fatalf("Deleting clusters by id should not fail. Error: %v", err)
	}

	// DELETE FROM users WHERE id=...
	_, err = u.DeleteById(nil, userRow.ID)
	if err != nil {
		t.Fatalf("Deleting user by id should not fail. Error: %v", err)
	}
}
