package commands

type Permission int

const (
	RegularUserPermission Permission = iota
	VoicerPermission
	DriverPermission
	ModeratorPermission
	OwnerPermission
	AdminPermission
)

func (p Permission) String() string {
	return [...]string{"none", "+", "^", "%", "@", "#", "admin"}[p]
}

func (p Permission) Rune() rune {
	return [...]rune{' ', '+', '^', '%', '@', '#', '~'}[p]
}
