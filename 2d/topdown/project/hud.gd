extends CanvasLayer

var player: KinematicBody2D

func _ready():
	player = get_node("../objects/player")
	var err = player.connect("moved", self, "_on_player_moved")
	if err != OK:
		print("failure to connect to moved player signal")


func _on_player_moved(velocity: Vector2) -> void:
	# $"top/container/speed".text = "SPD: %.2f" % velocity.length()
	$"top/container/speed".text = "SPD: (%.2f, %.2f)" % [velocity.x, velocity.y]
	var pos = player.position
	$"top/container/position".text = "POS: (%.2f, %.2f)" % [pos.x, pos.y]
