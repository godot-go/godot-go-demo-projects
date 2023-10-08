extends CanvasLayer

var player: CharacterBody2D

func _ready():
	player = get_node("../objects/player")
	var err = player.connect("moved",Callable(self,"_on_player_moved"))
	if err != OK:
		print("failure to connect to moved player signal")


func _on_player_moved(velocity: Vector2) -> void:
	var dir = player.direction
	var pos = player.position
	$"top/container/direction".text = "DIR: (%.2f, %.2f)" % [dir.x, dir.y]
	$"top/container/position".text = "POS: (%.2f, %.2f)" % [pos.x, pos.y]
