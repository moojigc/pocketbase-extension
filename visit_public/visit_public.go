package visit_public

import (
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/types"
)

// ensures that the VisitPublic struct satisfy the models.Model interface
var _ models.Model = (*VisitPublic)(nil)

type VisitPublic struct {
	models.BaseModel

	Id                 string         `db:"id" json:"id"`
	IpAddress          string         `db:"ip_address" json:"ip_address"`
	Origin             string         `db:"origin" json:"origin"`
	UserAgent          string         `db:"user_agent" json:"user_agent"`
	UniqueVisitorToken string         `db:"unique_visitor_token" json:"unique_visitor_token"`
	Created            types.DateTime `db:"created" json:"created"`
	Updated            types.DateTime `db:"updated" json:"updated"`
}

func (m *VisitPublic) TableName() string {
	return "visit" // the name of your collection
}

func (m *VisitPublic) MatchIp(ip_address string) *VisitPublic {
	if m.IpAddress != ip_address {
		m.IpAddress = ""
	}
	return m
}
