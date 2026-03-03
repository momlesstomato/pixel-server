package room

import "github.com/mlange-42/ark/ecs"

// RoomWorld wraps an Ark ECS World for a single room instance.
// Mappers and filters are created once and reused every tick.
type RoomWorld struct {
	World *ecs.World

	// Mappers — used to create/get/set component data on entities.
	AvatarMapper *ecs.Map4[Position, TileRef, WalkPath, AvatarID]
	ItemMapper   *ecs.Map3[Position, TileRef, ItemInteraction]
	PetMapper    *ecs.Map3[Position, TileRef, PetAI]
	BotMapper    *ecs.Map3[Position, TileRef, BotAI]

	// Single-component mappers for optional components.
	StatusMapper   *ecs.Map1[Status]
	CooldownMapper *ecs.Map1[ChatCooldown]
	DirtyMapper    *ecs.Map1[Dirty]
	KindMapper     *ecs.Map1[EntityKind]

	// Filters — used to iterate entities matching a component set.
	WalkFilter   *ecs.Filter3[Position, TileRef, WalkPath]
	ChatFilter   *ecs.Filter2[AvatarID, ChatCooldown]
	ItemFilter   *ecs.Filter2[TileRef, ItemInteraction]
	PetFilter    *ecs.Filter2[TileRef, PetAI]
	BotFilter    *ecs.Filter2[TileRef, BotAI]
	DirtyFilter  *ecs.Filter1[Dirty]
}

// NewRoomWorld creates a new ECS world with all mappers and filters initialised.
func NewRoomWorld() *RoomWorld {
	w := ecs.NewWorld()
	return &RoomWorld{
		World:          w,
		AvatarMapper:   ecs.NewMap4[Position, TileRef, WalkPath, AvatarID](w),
		ItemMapper:     ecs.NewMap3[Position, TileRef, ItemInteraction](w),
		PetMapper:      ecs.NewMap3[Position, TileRef, PetAI](w),
		BotMapper:      ecs.NewMap3[Position, TileRef, BotAI](w),
		StatusMapper:   ecs.NewMap1[Status](w),
		CooldownMapper: ecs.NewMap1[ChatCooldown](w),
		DirtyMapper:    ecs.NewMap1[Dirty](w),
		KindMapper:     ecs.NewMap1[EntityKind](w),
		WalkFilter:     ecs.NewFilter3[Position, TileRef, WalkPath](w),
		ChatFilter:     ecs.NewFilter2[AvatarID, ChatCooldown](w),
		ItemFilter:     ecs.NewFilter2[TileRef, ItemInteraction](w),
		PetFilter:      ecs.NewFilter2[TileRef, PetAI](w),
		BotFilter:      ecs.NewFilter2[TileRef, BotAI](w),
		DirtyFilter:    ecs.NewFilter1[Dirty](w),
	}
}

// SpawnAvatar creates an avatar entity at the given tile position.
func (rw *RoomWorld) SpawnAvatar(userID int64, roomUnit int32, x, y int16, z float32) ecs.Entity {
	e := rw.AvatarMapper.NewEntity(
		&Position{X: float32(x), Y: float32(y), Z: z},
		&TileRef{X: x, Y: y},
		&WalkPath{},
		&AvatarID{UserID: userID, RoomUnit: roomUnit},
	)
	rw.KindMapper.Add(e, &EntityKind{Kind: KindAvatar})
	rw.StatusMapper.Add(e, &Status{Posture: PostureStand})
	rw.CooldownMapper.Add(e, &ChatCooldown{})
	return e
}

// SpawnBot creates a bot entity at the given tile position.
func (rw *RoomWorld) SpawnBot(behaviour uint8, chatLines []string, x, y int16, z float32) ecs.Entity {
	e := rw.BotMapper.NewEntity(
		&Position{X: float32(x), Y: float32(y), Z: z},
		&TileRef{X: x, Y: y},
		&BotAI{Behaviour: behaviour, ChatLines: chatLines},
	)
	rw.KindMapper.Add(e, &EntityKind{Kind: KindBot})
	return e
}

// SpawnPet creates a pet entity at the given tile position.
func (rw *RoomWorld) SpawnPet(happy, energy int32, x, y int16, z float32) ecs.Entity {
	e := rw.PetMapper.NewEntity(
		&Position{X: float32(x), Y: float32(y), Z: z},
		&TileRef{X: x, Y: y},
		&PetAI{HappyLevel: happy, Energy: energy},
	)
	rw.KindMapper.Add(e, &EntityKind{Kind: KindPet})
	return e
}

// SpawnItem creates an item entity at the given tile position.
func (rw *RoomWorld) SpawnItem(furniID int64, extraData string, x, y int16, z float32) ecs.Entity {
	e := rw.ItemMapper.NewEntity(
		&Position{X: float32(x), Y: float32(y), Z: z},
		&TileRef{X: x, Y: y},
		&ItemInteraction{FurniID: furniID, ExtraData: extraData},
	)
	rw.KindMapper.Add(e, &EntityKind{Kind: KindItem})
	return e
}

// RemoveEntity removes an entity from the world.
func (rw *RoomWorld) RemoveEntity(entity ecs.Entity) {
	rw.World.RemoveEntity(entity)
}
