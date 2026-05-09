package game

import "github.com/Carlosnova1/strellas/game"

func main() {
game.Main()
}

// Main inicia el juego (para Android)
func Main() {
ebiten.SetWindowSize(360, 640)
ebiten.SetWindowTitle("✨ Strellas: Princesa del Saber ✨")
ebiten.SetWindowResizable(true)
ebiten.SetTPS(60)

game := NewGame()

if err := ebiten.RunGame(game); err != nil {
log.Fatal(err)
}
}
