package data

import (
	"fmt"
	"github.com/sifterstudios/bitbucket-notifier/notification"
)

var (
	CurrentPrActivity []Activity
)

func HandlePrActivity(activePrs []PullRequest, allSlicesOfActivities [][]Activity) {
	if len(activePrs) != len(allSlicesOfActivities) {
		fmt.Println("Error: Mismatch of PRs and slices of activities returned")
		return
	}
	for i, sliceOfActivities := range allSlicesOfActivities {
		for _, a := range sliceOfActivities {
			handleDifference(activePrs[i].Title, a)
		}
	}
	if len(CurrentPrActivity) == 0 {
		CurrentPrActivity = flatten(allSlicesOfActivities)
	}
}

func flatten(activities [][]Activity) []Activity {
	var flattened []Activity
	for _, slice := range activities {
		for _, activity := range slice {
			flattened = append(flattened, activity)
		}
	}
	return flattened
}

func handleDifference(prTitle string, activity Activity) {
	if !containsActivity(CurrentPrActivity, activity) { // TODO: I think now every comment will be notified when there's an answer to that comment.
		handleNotifying(prTitle, activity, false)
		CurrentPrActivity = append(CurrentPrActivity, activity)
	} else if isUpdate(activity) {
		handleNotifying(prTitle, activity, true)
		CurrentPrActivity = update(CurrentPrActivity, activity)
	}
}

func handleNotifying(prTitle string, activity Activity, isUpdate bool) {
	if activity.Action == "COMMENTED" {
		configUsername := string(UserConfig.Credentials.Username)
		// NOTE: Different servers use email/username to authenticate
		if activity.User.Name != configUsername &&
			activity.User.EmailAddress != configUsername {
			notifyAboutNewComment(activity.User.DisplayName, activity.Comment.Text, activity.CommentAnchor.Path, prTitle)
		}
	}
}

func notifyAboutNewComment(authorName string, message string, filePath, prTitle string) {
	fmt.Printf("New comment by %s on PR %s: %s\n", authorName, filePath, message)
	err := notification.SendNotification(fmt.Sprintf("New comment by %s on PR %s", authorName, prTitle), fmt.Sprintf("%s/n %s", filePath, message))
	if err != nil {
		fmt.Println(err)
	}

}
func update(currentPrs []Activity, newActivity Activity) []Activity {
	for i, activity := range currentPrs {
		if activity.ID == newActivity.ID {
			currentPrs[i] = newActivity
		}
	}
	return currentPrs
}

func isUpdate(activity Activity) bool {
	return activity.Action == "UPDATED"
}

func containsActivity(currentPrActivity []Activity, newActivity Activity) bool {
	var foundComment bool
	var foundCommentThread bool
	for _, activity := range currentPrActivity {
		if activity.ID == newActivity.ID {
			foundComment = true
		}
		if len(newActivity.Comment.CommentThread) > len(activity.Comment.CommentThread) {
			foundCommentThread = true
		}
	}
	return foundComment && foundCommentThread
}

type PullRequestActivityResponse struct {
	Size       int        `json:"size"`
	Limit      int        `json:"limit"`
	IsLastPage bool       `json:"isLastPage"`
	Values     []Activity `json:"values"`
	Start      int        `json:"start"`
}
type Comment struct {
	Properties          CommentProperties `json:"properties"`
	ID                  int               `json:"id"`
	Version             int               `json:"version"`
	Text                string            `json:"text"`
	Author              User              `json:"author"`
	CreatedDate         int64             `json:"createdDate"`
	UpdatedDate         int64             `json:"updatedDate"`
	CommentThread       []Comment         `json:"comments"`
	Tasks               []Task            `json:"tasks"`
	Severity            string            `json:"severity"`
	State               string            `json:"state"`
	PermittedOperations struct {
		Editable       bool `json:"editable"`
		Transitionable bool `json:"transitionable"`
		Deletable      bool `json:"deletable"`
	} `json:"permittedOperations"`
	ResolvedDate int64 `json:"resolvedDate"`
	Resolver     User  `json:"resolver"`
}

type CommentProperties struct {
	RepositoryID int `json:"repositoryId"`
}

type Task struct {
	Properties          Properties `json:"properties"`
	ID                  int        `json:"id"`
	Version             int        `json:"version"`
	Text                string     `json:"text"`
	Author              User       `json:"author"`
	CreatedDate         int64      `json:"createdDate"`
	UpdatedDate         int64      `json:"updatedDate"`
	Comments            []Comment  `json:"comments"`
	Tasks               []Task     `json:"tasks"`
	Severity            string     `json:"severity"`
	State               string     `json:"state"`
	PermittedOperations struct {
		Editable       bool `json:"editable"`
		Transitionable bool `json:"transitionable"`
		Deletable      bool `json:"deletable"`
	} `json:"permittedOperations"`
}

type Diff struct {
	Source      interface{} `json:"source"`
	Destination struct {
		Components []string `json:"components"`
		Parent     string   `json:"parent"`
		Name       string   `json:"name"`
		Extension  string   `json:"extension"`
		ToString   string   `json:"toString"`
	} `json:"destination"`
	Hunks []struct {
		Context         string `json:"context"`
		SourceLine      int    `json:"sourceLine"`
		SourceSpan      int    `json:"sourceSpan"`
		DestinationLine int    `json:"destinationLine"`
		DestinationSpan int    `json:"destinationSpan"`
		Segments        []struct {
			Type  string `json:"type"`
			Lines []struct {
				Destination int    `json:"destination"`
				Source      int    `json:"source"`
				Line        string `json:"line"`
				Truncated   bool   `json:"truncated"`
			} `json:"lines"`
			Truncated bool `json:"truncated"`
		} `json:"segments"`
		Truncated bool `json:"truncated"`
	} `json:"hunks"`
	Truncated  bool `json:"truncated"`
	Properties struct {
		ToHash   string `json:"toHash"`
		Current  bool   `json:"current"`
		FromHash string `json:"fromHash"`
	} `json:"properties"`
}

type CommentAnchor struct {
	FromHash string `json:"fromHash"`
	ToHash   string `json:"toHash"`
	Line     int    `json:"line"`
	LineType string `json:"lineType"`
	FileType string `json:"fileType"`
	Path     string `json:"path"`
	DiffType string `json:"diffType"`
	Orphaned bool   `json:"orphaned"`
}

type Activity struct {
	ID            int           `json:"id"`
	CreatedDate   int64         `json:"createdDate"`
	User          User          `json:"user"`
	Action        string        `json:"action"`
	CommentAction string        `json:"commentAction"`
	Comment       Comment       `json:"comment"`
	CommentAnchor CommentAnchor `json:"commentAnchor"`
	Diff          Diff          `json:"diff"`
}
