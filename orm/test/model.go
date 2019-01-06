package test

// Group Group
type Group struct {
	ID     int      `orm:"id key auto"`
	Name   string   `orm:"name"`
	Users  *[]*User `orm:"users"`
	Parent *Group   `orm:"parent"`
}

// User User
type User struct {
	ID    int      `orm:"id key auto"`
	Name  string   `orm:"name"`
	EMail string   `orm:"email"`
	Group []*Group `orm:"group"`
}

// System System
type System struct {
	ID    int      `orm:"id key auto"`
	Name  string   `orm:"name"`
	Users *[]User  `orm:"users"`
	Tags  []string `orm:"tags"`
}

// Equle Equle
func (s *Group) Equle(r *Group) bool {
	if s.ID != r.ID {
		return false
	}
	if s.Name != r.Name {
		return false
	}

	if s.Users == nil && r.Users != nil {
		return false
	}

	if s.Users != nil && r.Users == nil {
		return false
	}

	if s.Users != nil && r.Users != nil {
		if len(*(s.Users)) != len(*(r.Users)) {
			return false
		}

		for idx := 0; idx < len(*(s.Users)); idx++ {
			l := (*(s.Users))[idx]
			r := (*(r.Users))[idx]
			if !l.Equle(r) {
				return false
			}
		}
	}
	if s.Parent == nil && r.Parent != nil {
		return false
	}

	if s.Parent != nil && r.Parent == nil {
		return false
	}
	if s.Parent != nil && r.Parent != nil {
		if !s.Parent.Equle(r.Parent) {
			return false
		}
	}

	return true
}

// Equle check user Equle
func (s *User) Equle(r *User) bool {
	if s.ID != r.ID {
		return false
	}
	if s.Name != r.Name {
		return false
	}
	if s.EMail != r.EMail {
		return false
	}
	if len(s.Group) != len(r.Group) {
		return false
	}

	for idx := 0; idx < len(s.Group); idx++ {
		l := s.Group[idx]
		r := r.Group[idx]
		if !l.Equle(r) {
			return false
		}
	}

	return true
}

// Equle Equle
func (s *System) Equle(r *System) bool {
	if s.ID != r.ID {
		return false
	}
	if s.Name != r.Name {
		return false
	}

	if s.Users == nil && r.Users != nil {
		return false
	}

	if s.Users != nil && r.Users == nil {
		return false
	}

	if s.Users != nil && r.Users != nil {
		if len(*(s.Users)) != len(*(r.Users)) {
			return false
		}

		for idx := 0; idx < len(*(s.Users)); idx++ {
			l := (*(s.Users))[idx]
			r := (*(r.Users))[idx]
			if !l.Equle(&r) {
				return false
			}
		}
	}
	if len(s.Tags) != len(r.Tags) {
		return false
	}

	for idx := 0; idx < len(s.Tags); idx++ {
		l := s.Tags[idx]
		r := r.Tags[idx]
		if l != r {
			return false
		}
	}

	return true
}
