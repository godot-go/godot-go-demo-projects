extends CanvasLayer

var player: KinematicBody2D
var speed: Label
var position: Label

func _ready():
	player = $"/root/World/Objects/Player"
	var err = player.connect("moved", self, "_on_player_moved")
	if err != OK:
		print("failure to connect to moved player signal")
	var stats = $"TopPanel/Margin/Rows/Stats"
	speed = stats.get_node("Speed")
	position = stats.get_node("Position")


func _on_player_moved(velocity: Vector2) -> void:
	speed.text = "(%.2f, %.2f)" % [velocity.x, velocity.y]
	var pos = player.position
	position.text = "(%.2f, %.2f)" % [pos.x, pos.y]
