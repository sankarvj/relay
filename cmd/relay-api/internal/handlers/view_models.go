package handlers

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/chart"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"gitlab.com/vjsideprojects/relay/internal/rule/flow"
	"gitlab.com/vjsideprojects/relay/internal/rule/node"
	"gitlab.com/vjsideprojects/relay/internal/team"
	"gitlab.com/vjsideprojects/relay/internal/timeseries"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"gitlab.com/vjsideprojects/relay/internal/visitor"
)

func createViewModelVisitor(v visitor.Visitor, entityName, itemName string) ViewModelVisitor {
	return ViewModelVisitor{
		ID:         v.VistitorID,
		TeamID:     v.TeamID,
		Name:       v.Name,
		Email:      v.Email,
		EntityID:   v.EntityID,
		ItemID:     v.ItemID,
		EntityName: entityName,
		ItemName:   itemName,
		Active:     v.Active,
		SignedIn:   v.SignedIn,
		ExpireAt:   v.ExpireAt,
	}
}

type ViewModelVisitor struct {
	ID         string    `json:"id"`
	TeamID     string    `json:"team_id"`
	EntityID   string    `json:"entity_id"`
	ItemID     string    `json:"item_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	EntityName string    `json:"entity_name"`
	ItemName   string    `json:"item_name"`
	Active     bool      `json:"active"`
	SignedIn   bool      `json:"signed_in"`
	ExpireAt   time.Time `json:"expire_at"`
	Body       string    `json:"body"`
}

func createViewModelUser(u user.User) user.ViewModelUser {
	return user.ViewModelUser{
		Name:   *u.Name,
		Avatar: *u.Avatar,
		Email:  u.Email,
		Phone:  *u.Phone,
		Roles:  u.Roles,
	}
}

func createViewModelUS(us user.UserSetting) user.ViewModelUserSetting {
	return user.ViewModelUserSetting{
		AccountID:           us.AccountID,
		UserID:              us.UserID,
		LayoutStyle:         us.LayoutStyle,
		SelectedTeam:        us.SelectedTeam,
		NotificationSetting: user.UnmarshalNotificationSettings(us.NotificationSetting),
	}
}

type UserToken struct {
	Token    string   `json:"token"`
	Accounts []string `json:"accounts"`
	Team     string   `json:"team"`
	Entity   string   `json:"entity"`
	Item     string   `json:"item"`
}

func createViewModelCharts(charts []chart.Chart, gridResMap map[string]Grid) []VMChart {
	vmCharts := make([]VMChart, 0)
	for _, ch := range charts {
		count := 0
		change := 0
		if grid, ok := gridResMap[ch.ID]; ok {
			count = grid.Count
			change = grid.Change
		}
		vmCharts = append(vmCharts, createViewModelChart(ch, []Series{}, count, change))
	}
	return vmCharts
}

func createViewModelChartNoChange(c chart.Chart, series []Series, count int) VMChart {
	return createViewModelChart(c, series, count, -1000000)
}

func createViewModelChart(c chart.Chart, series []Series, count, change int) VMChart {
	return VMChart{
		ID:       c.ID,
		EntityID: c.EntityID,
		Title:    c.Name,
		Type:     c.Type,
		Field:    c.GetField(),
		DataType: c.GetDType(),
		Duration: c.Duration,
		Series:   series,
		Count:    count,
		Change:   change,
	}
}

func createVMSeries(ts timeseries.Timeseries) Series {
	return Series{
		ID:          ts.ID,
		Event:       ts.Event,
		Description: ts.Description,
		Count:       ts.Count,
		StartTime:   ts.StartTime,
		EndTime:     ts.EndTime,
	}
}

func createVMSeriesFromMap(id, label string, value int) Series {
	return Series{
		ID:    id,
		Label: label,
		Count: value,
	}
}

type VMChart struct {
	ID       string   `json:"id"`
	EntityID string   `json:"entity_id"`
	Title    string   `json:"title"`
	Field    string   `json:"field"`
	Type     string   `json:"type"`
	DataType string   `json:"data_type"`
	Duration string   `json:"duration"`
	Series   []Series `json:"series"`
	Count    int      `json:"count"`
	Change   int      `json:"change"`
}

type Series struct {
	ID          string    `json:"timeseries_id"`
	Event       string    `json:"event"`
	Description string    `json:"description"`
	Count       int       `json:"count"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Label       string    `json:"label"`
}

type Grid struct {
	Count   int `json:"count"`
	Change  int `json:"change"`
	Insight int `json:"insight"`
}

type GridsResponse struct {
	Grids []VMChart `json:"grids"`
}

type ItemResultBody struct {
	Items      []item.Item    `json:"items"`
	TotalCount map[string]int `json:"total_count"`
}

type FilterBody struct {
	Name    string       `json:"name"`
	Queries []node.Query `json:"queries"`
}

type Segmenter struct {
	exp       string
	sortby    string
	direction string
	page      int
	doCount   bool
	source    *graphdb.Field
	useReturn bool //makes the get result in graphdb to use dst.alias instead of source.alias
}

func createViewModelNode(n node.Node) node.ViewModelNode {
	return node.ViewModelNode{
		ID:             n.ID,
		FlowID:         n.FlowID,
		StageID:        n.StageID,
		Name:           nameOfType(n.Type),
		Description:    n.Description,
		Expression:     n.Expression,
		ParentNodeID:   n.ParentNodeID,
		ActorID:        n.ActorID,
		EntityName:     "", //should I populate this?
		EntityCategory: -1, //should I populate this?
		Type:           n.Type,
		Actuals:        n.ActualsMap(),
	}
}

func createViewModelMember(id string, fields map[string]entity.Field, teamMap map[string]team.Team) ViewModelMember {
	return ViewModelMember{
		ID:     id,
		Name:   fields["name"].Value.(string),
		Email:  fields["email"].Value.(string),
		Avatar: fields["avatar"].Value.(string),
		Teams:  populateTeams(fields["team_ids"].Value, teamMap),
		Role:   fields["role"].Value.([]interface{}),
	}
}

func recreateFields(vm ViewModelMember, namedKeys map[string]string) map[string]interface{} {
	itemFields := make(map[string]interface{}, 0)
	itemFields[namedKeys["name"]] = vm.Name
	itemFields[namedKeys["user_id"]] = vm.UserID
	itemFields[namedKeys["email"]] = vm.Email
	itemFields[namedKeys["avatar"]] = vm.Avatar
	itemFields[namedKeys["team_ids"]] = stripeTeamIds(vm.Teams)
	itemFields[namedKeys["role"]] = vm.Role
	return itemFields
}

type ViewModelMember struct {
	ID     string        `json:"id"`
	UserID string        `json:"user_id"`
	Name   string        `json:"name"`
	Email  string        `json:"email"`
	Avatar string        `json:"avatar"`
	Teams  []ViewTeam    `json:"teams"`
	Role   []interface{} `json:"role"`
}

type ViewTeam struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Cypher struct {
	View   bool `json:"view"`
	Edit   bool `json:"edit"`
	Create bool `json:"create"`
}

func createViewModelItem(i item.Item) ViewModelItem {
	return ViewModelItem{
		ID:        i.ID,
		EntityID:  i.EntityID,
		StageID:   i.StageID,
		Name:      i.Name,
		Type:      i.Type,
		State:     i.State,
		Fields:    i.Fields(),
		Meta:      i.Meta(),
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}

func itemResponse(items []item.Item, uMap map[string]*user.User) []ViewModelItem {
	viewModelItems := make([]ViewModelItem, len(items))
	for i, item := range items {
		viewModelItems[i] = createViewModelItem(item)
		if item.UserID != nil {
			avatar, name, _ := userAvatarNameEmail(uMap[*item.UserID])
			viewModelItems[i].UserAvatar = avatar
			viewModelItems[i].UserName = name
		}
	}
	return viewModelItems
}

// ViewModelItem represents the view model of item
// (i.e) it has fields instead of attributes
type ViewModelItem struct {
	ID         string                 `json:"id"`
	EntityID   string                 `json:"entity_id"`
	StageID    *string                `json:"stage_id"`
	Name       *string                `json:"name"`
	UserName   string                 `json:"user_name"`
	UserAvatar string                 `json:"user_avatar"`
	Type       int                    `json:"type"`
	State      int                    `json:"state"`
	Fields     map[string]interface{} `json:"fields"`
	Meta       map[string]interface{} `json:"meta"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  int64                  `json:"updated_at"`
}

func createViewModelNodeActor(n node.NodeActor) node.ViewModelNode {
	return node.ViewModelNode{
		ID:             n.ID,
		FlowID:         n.FlowID,
		StageID:        n.StageID,
		Name:           n.Name,
		Description:    n.Description,
		Expression:     n.Expression,
		ParentNodeID:   n.ParentNodeID,
		ActorID:        n.ActorID,
		Weight:         n.Weight,
		EntityName:     n.EntityName.String,
		EntityCategory: int(n.EntityCategory.Int32),
		Type:           n.Type,
		Tokens:         n.Tokens(),
		Actuals:        n.ActualsMap(),
	}
}

func createViewModelActiveNode(n flow.ActiveNode) node.ViewModelActiveNode {
	return node.ViewModelActiveNode{
		ID:        n.NodeID,
		IsActive:  n.IsActive,
		Life:      n.Life,
		CreatedAt: n.CreatedAt,
	}
}

func createViewModelFlow(f flow.Flow, nodes []node.ViewModelNode) flow.ViewModelFlow {
	return flow.ViewModelFlow{
		ID:          f.ID,
		EntityID:    f.EntityID,
		Name:        f.Name,
		Description: f.Description,
		Expression:  f.Expression,
		Mode:        f.Mode,
		State:       f.State,
		Type:        f.Type,
		Status:      f.Status,
		Nodes:       nodes,
		Tokens:      f.Tokens(),
	}
}

type ViewModelEvent struct {
	EventID         string      `json:"event_id"`
	EventEntity     string      `json:"event_entity"`
	EventEntityName string      `json:"event_entity_name"`
	UserAvatar      string      `json:"user_avatar"`
	UserName        string      `json:"user_name"`
	UserEmail       string      `json:"user_email"`
	Action          interface{} `json:"action"` //lable:action - created, clicked, viewed, updated, etc
	Title           interface{} `json:"title"`  //lable:title  - task, deal, amazon.com
	Footer          interface{} `json:"footer"` //lable:footer - 8 times
	Time            time.Time   `json:"time"`
}

func createViewModelChildren(e entity.Entity, relationshipID string) ViewModelChildren {
	return ViewModelChildren{
		ID:             e.ID,
		TeamID:         e.TeamID,
		Name:           e.Name,
		DisplayName:    e.DisplayName,
		Category:       e.Category,
		State:          e.State,
		RelationshipID: relationshipID,
	}
}

type ViewModelChildren struct {
	ID             string `json:"id"`
	TeamID         string `json:"team_id"`
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	Category       int    `json:"category"`
	State          int    `json:"state"`
	RelationshipID string `json:"relationship_id"`
}

type Association struct {
	DstEntityID    string `json:"dst_entity_id"`
	RelationshipID string `json:"relationship_id"`
	Remove         bool   `json:"remove"`
}

type AssociationReqBody struct {
	AssociationReqs []Association `json:"association_reqs"`
}

func createNewVerifiedUser(ctx context.Context, name, email string, roles []string, db *sqlx.DB) (user.User, error) {
	nu := user.NewUser{
		Name:            name,
		Avatar:          util.String(""),
		Email:           email,
		Phone:           util.String(""),
		Provider:        util.String("default"),
		Password:        "",
		PasswordConfirm: "",
		Verified:        true,
		Roles:           roles,
	}
	u, err := user.Create(ctx, db, nu, time.Now())
	return u, err
}
