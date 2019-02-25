package insteon

// Group represents a device grouop.
type Group struct {
	groupID byte
	hub     Hub
}

// NewGroup creates a new group.
func NewGroup(hub Hub, groupID byte) *Group {
	return &Group{
		groupID: groupID,
		hub:     hub,
	}
}

// TurnOn turns on the group.
func (g *Group) TurnOn() error {
	return g.hub.SendGroupCommand(cmdControlOn, g.groupID)
}

// TurnOff turns off the group.
func (g *Group) TurnOff() error {
	return g.hub.SendGroupCommand(cmdControlOff, g.groupID)
}
