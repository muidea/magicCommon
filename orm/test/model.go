package test

// Group Group
type Group struct {
	ID       int      `orm:"id key auto"`
	Name     string   `orm:"name"`
	Users    []*User  `orm:"users"`
	Children []*Group `orm:"children"`
	Parent   *Group   `orm:"parent"`
}

// User User
type User struct {
	ID    int      `orm:"id key auto"`
	Name  string   `orm:"name"`
	EMail string   `orm:"email"`
	Group []*Group `orm:"group"`
}
