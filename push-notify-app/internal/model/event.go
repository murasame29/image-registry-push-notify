package model

type ECRPushEvent struct {
	Account    string   `json:"account"`
	Detail     Detail   `json:"detail"`
	DetailType string   `json:"detail-type"`
	ID         string   `json:"id"`
	Region     string   `json:"region"`
	Resource   []string `json:"resource"`
	Source     string   `json:"source"`
	Time       string   `json:"time"`
	Version    string   `json:"version"`
}

type ECRActionType string

const (
	ECRAcTionPush ECRActionType = "PUSH"
)

type Detail struct {
	ActionType        ECRActionType `json:"action-type"`
	ImageDigest       string        `json:"image-digest"`
	ImageTag          string        `json:"image-tag"`
	RepositoryName    string        `json:"repository-name"`
	Result            string        `json:"result"`
	ManifestMediaType string        `json:"manifest-media-type"`
	ArtifactMadiaType string        `json:"artifact-media-type"`
}
