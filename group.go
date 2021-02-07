package insteon

import "context"

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
func (g *Group) TurnOn(ctx context.Context) error {
	return g.hub.SendGroupCommand(ctx, cmdControlOn, g.groupID)
}

// TurnOff turns off the group.
func (g *Group) TurnOff(ctx context.Context) error {
	return g.hub.SendGroupCommand(ctx, cmdControlOff, g.groupID)
}
