package v4

import "time"

type Actor struct {
	Login string
}

// PageInfo represents https://developer.github.com/v4/object/pageinfo/
type PageInfo struct {
	HasNextPage bool
	EndCursor   string
}

// Repository represents https://developer.github.com/v4/object/repository/
type Repository struct {
	RepositoryFields
	Issues       IssueConnection       `graphql:"issues(first: $pageList, after: $issuesCursor)"`
	PullRequests PullRequestConnection `graphql:"pullRequests(first: $pageList, after: $pullRequestsCursor)"`
} // `graphql:"repository(owner: $owner, name: $name)"`

type Ref struct {
	Id     string
	Name   string
	Prefix string
}

// RepositoryFields defines the fields for Repository, containing only the
// fields from https://developer.github.com/v4/interface/repositoryinfo/
type RepositoryFields struct {
	CreatedAt   time.Time
	DatabaseId  int
	Description string
	//DescriptionHTML  string `graphql:"descriptionHTML"`
	ForkCount        int
	HasIssuesEnabled bool
	HasWikiEnabled   bool
	HomepageUrl      string
	IsArchived       bool
	IsFork           bool
	IsLocked         bool
	IsMirror         bool
	IsPrivate        bool
	IsTemplate       bool
	//LicenseInfo: License
	//LockReason: RepositoryLockReason
	MirrorUrl         string
	Name              string
	NameWithOwner     string
	OpenGraphImageUrl string
	Owner             Actor
	PushedAt          time.Time
	ResourcePath      string
	//ShortDescriptionHTML     string `graphql:"shortDescriptionHTML"`
	UpdatedAt                time.Time
	Url                      string
	UsesCustomOpenGraphImage bool
}

// IssueConnection represents https://developer.github.com/v4/object/issueconnection/
type IssueConnection struct {
	PageInfo PageInfo
	Nodes    []Issue
} //`graphql:"issues(first: $pageList, after: $issuesCursor)"`

type IssueCommentsConnection struct {
	//TotalCount int
	PageInfo PageInfo
	Nodes    []IssueComment
} // `graphql:"comments(first: $pageList, after: $issueCommentsCursor)"`

// Issue represents https://developer.github.com/v4/object/issue/
type Issue struct {
	IssueFields
	Comments IssueCommentsConnection `graphql:"comments(first: $pageList, after: $issueCommentsCursor)"`
} // `graphql:"issue(number: $issueNumber)"`

type IssueFields struct {
	//activeLockReason string
	//assignees

	Author Actor

	//authorAssociation: CommentAuthorAssociation!
	Body string
	//BodyHTML string `graphql:"bodyHTML"`
	//BodyText string
	Closed   bool
	ClosedAt time.Time
	//comments
	CreatedAt       time.Time
	CreatedViaEmail bool
	DatabaseId      int
	//Editor: Actor
	//Id                  string
	IncludesCreatedEdit bool
	//labels
	LastEditedAt time.Time
	Locked       bool
	//Milestone: Milestone
	Number int
	//participants
	//projectCards
	PublishedAt time.Time
	//ReactionGroups
	//reactions
	//Repository: Repository!
	ResourcePath string
	State        string
	// timelineItems
	Title     string
	UpdatedAt time.Time
	Url       string
	//userContentEdits
	//ViewerCanReact bool
	//ViewerCanSubscribe bool
	//ViewerCanUpdate bool
	//ViewerCannotUpdateReasons: [CommentCannotUpdateReason!]!
	//ViewerDidAuthor bool
	//ViewerSubscription: SubscriptionState
}

type IssueComment struct {
	Author Actor
	//authorAssociation: CommentAuthorAssociation!
	Body string
	//bodyHTML: HTML!
	//BodyText            string
	CreatedAt       time.Time
	CreatedViaEmail bool
	DatabaseId      int
	Editor          Actor
	//Id                  string
	IncludesCreatedEdit bool
	IsMinimized         bool
	LastEditedAt        time.Time
	MinimizedReason     string
	PublishedAt         time.Time
	//reactionGroups: [ReactionGroup!]
	//reactions
	ResourcePath string
	UpdatedAt    time.Time
	Url          string
	//userContentEdits
	// ViewerCanDelete bool
	// ViewerCanMinimize bool
	// ViewerCanReact bool
	// ViewerCanUpdate bool
	// viewerCannotUpdateReasons: [CommentCannotUpdateReason!]!
	// ViewerDidAuthor bool
}

type PullRequestConnection struct {
	PageInfo PageInfo
	Nodes    []PullRequest
} //`graphql:"pullRequests(first: $pageList, after: $pullRequestsCursor)"`

type PullRequest struct {
	PullRequestFields
	Comments IssueCommentsConnection     `graphql:"comments(first: $pageList, after: $issueCommentsCursor)"`
	Reviews  PullRequestReviewConnection `graphql:"reviews(first: $pageList, after: $pullRequestReviewsCursor)"`
} // `graphql:"pullRequest(number: $issueNumber)"`

type PullRequestFields struct {
	ActiveLockReason string
	Additions        int
	// assignees
	Author      Actor
	BaseRef     Ref
	BaseRefName string
	// BaseRefOid: GitObjectID!
	// BaseRepository: Repository
	Body string
	//BodyHTML: HTML!
	//BodyText string
	ChangedFiles int
	Closed       bool
	ClosedAt     time.Time
	//comments
	// commits
	CreatedAt       time.Time
	CreatedViaEmail bool
	DatabaseId      int
	Deletions       int
	Editor          Actor
	//files
	HeadRef     Ref
	HeadRefName string
	//HeadRefOid: GitObjectID!
	//HeadRepository: Repository
	//HeadRepositoryOwner: RepositoryOwner
	Id                  string
	IncludesCreatedEdit bool
	IsCrossRepository   bool
	// labels
	LastEditedAt        time.Time
	Locked              bool
	MaintainerCanModify bool
	//MergeCommit: Commit
	Mergeable string
	Merged    bool
	MergedAt  time.Time
	MergedBy  Actor
	//Milestone: Milestone
	Number int
	//participants
	Permalink string
	//PotentialMergeCommit: Commit
	//projectCards
	PublishedAt time.Time
	//ReactionGroups: [ReactionGroup!]
	//reactions
	ResourcePath       string
	RevertResourcePath string
	RevertUrl          string
	//reviewRequests
	//reviewThreads
	// reviews
	State string
	//SuggestedReviewers: [SuggestedReviewer]!
	//timelineItems
	Title     string
	UpdatedAt time.Time
	Url       string
	//userContentEdits
	//ViewerCanApplySuggestion bool
	//ViewerCanReact bool
	//ViewerCanSubscribe bool
	//ViewerCanUpdate bool
	//ViewerCannotUpdateReasons: [CommentCannotUpdateReason!]!
	//ViewerDidAuthor bool
	//ViewerSubscription: SubscriptionState
}

type PullRequestReviewConnection struct {
	//TotalCount int
	PageInfo PageInfo
	Nodes    []PullRequestReview
} // `graphql:"reviews(first: $pageList, after: $pullRequestReviewsCursor)"`

type PullRequestReview struct {
	Author Actor
	//AuthorAssociation: CommentAuthorAssociation!
	Body string
	//BodyHTML: HTML!
	//BodyText string
	Comments PullRequestReviewCommentConnection `graphql:"comments(first: $pageList, after: $pullRequestReviewCommentsCursor)"`
	//Commit: Commit
	CreatedAt           time.Time
	CreatedViaEmail     bool
	DatabaseId          int
	Editor              Actor
	Id                  string
	IncludesCreatedEdit bool
	LastEditedAt        time.Time
	//onBehalfOf(
	PublishedAt time.Time
	//PullRequest: PullRequest!
	//ReactionGroups: [ReactionGroup!]
	//reactions
	//Repository: Repository!
	ResourcePath string
	State        string
	SubmittedAt  time.Time
	UpdatedAt    time.Time
	Url          string
	//userContentEdits
	//ViewerCanDelete bool
	//ViewerCanReact bool
	//ViewerCanUpdate bool
	//ViewerCannotUpdateReasons: [CommentCannotUpdateReason!]!
	//ViewerDidAuthor bool
}

type PullRequestReviewCommentConnection struct {
	//TotalCount int
	PageInfo PageInfo
	Nodes    []PullRequestReviewComment
}

type PullRequestReviewComment struct {
	Author Actor
	//authorAssociation: CommentAuthorAssociation!
	Body string
	//bodyHTML: HTML!
	//BodyText            string
	CreatedAt           time.Time
	CreatedViaEmail     bool
	DatabaseId          int
	Editor              Actor
	Id                  string
	IncludesCreatedEdit bool
	IsMinimized         bool
	LastEditedAt        time.Time
	MinimizedReason     string
	PublishedAt         time.Time
	//reactionGroups: [ReactionGroup!]
	//reactions
	ResourcePath string
	UpdatedAt    time.Time
	Url          string
	//userContentEdits
	// ViewerCanDelete bool
	// ViewerCanMinimize bool
	// ViewerCanReact bool
	// ViewerCanUpdate bool
	// viewerCannotUpdateReasons: [CommentCannotUpdateReason!]!
	// ViewerDidAuthor bool
}
