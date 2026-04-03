package seed

import roommodel "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/model"

// standardModels returns the predefined room model templates matching Habbo defaults.
func standardModels() []roommodel.RoomModel {
	return []roommodel.RoomModel{
		modelA(), modelB(), modelC(), modelD(), modelE(),
		modelF(), modelG(), modelH(), modelI(),
	}
}

// modelA returns the small square room template.
func modelA() roommodel.RoomModel {
	return roommodel.RoomModel{
		Slug: "model_a", DoorX: 3, DoorY: 5, DoorZ: 0, DoorDir: 2, WallHeight: -1,
		Heightmap: "xxxxxxxxxxxx\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxxxxxxxxxx\rxxxxxxxxxxxx",
	}
}

// modelB returns the medium rectangular room template.
func modelB() roommodel.RoomModel {
	return roommodel.RoomModel{
		Slug: "model_b", DoorX: 0, DoorY: 5, DoorZ: 0, DoorDir: 2, WallHeight: -1,
		Heightmap: "xxxxxxxxxxxx\rxxxxx0000000\rxxxxx0000000\rxxxxx0000000\rxxxxx0000000\rx00000000000\rx00000000000\rx00000000000\rx00000000000\rx00000000000\rx00000000000\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx",
	}
}

// modelC returns the compact open room template.
func modelC() roommodel.RoomModel {
	return roommodel.RoomModel{
		Slug: "model_c", DoorX: 4, DoorY: 7, DoorZ: 0, DoorDir: 2, WallHeight: -1,
		Heightmap: "xxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx",
	}
}

// modelD returns the tall vertical room template.
func modelD() roommodel.RoomModel {
	return roommodel.RoomModel{
		Slug: "model_d", DoorX: 4, DoorY: 7, DoorZ: 0, DoorDir: 2, WallHeight: -1,
		Heightmap: "xxxxxxxxxxxx\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxxxxxxxxx",
	}
}

// modelE returns the mid-size open room template.
func modelE() roommodel.RoomModel {
	return roommodel.RoomModel{
		Slug: "model_e", DoorX: 1, DoorY: 5, DoorZ: 0, DoorDir: 2, WallHeight: -1,
		Heightmap: "xxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxx0000000000\rxx0000000000\rxx0000000000\rxx0000000000\rxx0000000000\rxx0000000000\rxx0000000000\rxx0000000000\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx",
	}
}

// modelF returns the L-shaped room template.
func modelF() roommodel.RoomModel {
	return roommodel.RoomModel{
		Slug: "model_f", DoorX: 2, DoorY: 5, DoorZ: 0, DoorDir: 2, WallHeight: -1,
		Heightmap: "xxxxxxxxxxxx\rxxxxxxx0000x\rxxxxxxx0000x\rxxx00000000x\rxxx00000000x\rxxx00000000x\rxxx00000000x\rx0000000000x\rx0000000000x\rx0000000000x\rx0000000000x\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx",
	}
}

// modelG returns the elevated platform room template.
func modelG() roommodel.RoomModel {
	return roommodel.RoomModel{
		Slug: "model_g", DoorX: 1, DoorY: 7, DoorZ: 1, DoorDir: 2, WallHeight: -1,
		Heightmap: "xxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxx00000\rxxxxxxx00000\rxxxxxxx00000\rxx1111000000\rxx1111000000\rxx1111000000\rxx1111000000\rxx1111000000\rxxxxxxx00000\rxxxxxxx00000\rxxxxxxx00000\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx",
	}
}

// modelH returns the split-level room template.
func modelH() roommodel.RoomModel {
	return roommodel.RoomModel{
		Slug: "model_h", DoorX: 4, DoorY: 4, DoorZ: 1, DoorDir: 2, WallHeight: -1,
		Heightmap: "xxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxx111111x\rxxxxx111111x\rxxxxx111111x\rxxxxx111111x\rxxxxx111111x\rxxxxx000000x\rxxxxx000000x\rxxx00000000x\rxxx00000000x\rxxx00000000x\rxxx00000000x\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx",
	}
}

// modelI returns the extra-large room template.
func modelI() roommodel.RoomModel {
	return roommodel.RoomModel{
		Slug: "model_i", DoorX: 0, DoorY: 10, DoorZ: 0, DoorDir: 2, WallHeight: -1,
		Heightmap: "xxxxxxxxxxxxxxxxx\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rx0000000000000000\rxxxxxxxxxxxxxxxxx",
	}
}
