package def

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Entity struct {
	EName string `json:"ename"`
	EID   int    `json:"eid"`
	EType string `json:"etype"`
}

// Decode decode entityPtr
func (s *Entity) Decode(req *http.Request) (err error) {
	str := req.URL.Query().Get("entityID")
	if str != "" {
		s.EID, err = strconv.Atoi(str)
		if err != nil {
			return
		}
	} else {
		err = fmt.Errorf("illegal entity info")
	}
	s.EName = req.URL.Query().Get("entityName")
	s.EType = req.URL.Query().Get("entityType")
	return
}

// Encode encode entityPtr
func (s *Entity) Encode(vals url.Values) url.Values {
	vals.Set("entityID", fmt.Sprintf("%d", s.EID))
	vals.Set("entityName", s.EName)
	vals.Set("entityType", s.EType)
	return vals
}
