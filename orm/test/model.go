package test

// Group Group
type Group struct {
	ID          int      `orm:"id key auto"`
	Name        string   `orm:"name"`
	User        []*User  `orm:"user"`
	SubGroup    []*Group `orm:"subGoup"`
	ParentGroup *Group   `orm:"parentGroup"`
}

// User User
type User struct {
	ID    int      `orm:"id key auto"`
	Name  string   `orm:"name"`
	EMail string   `orm:"email"`
	Group []*Group `orm:"group"`
}
