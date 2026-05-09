package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/Carlosnova1/strellas/data"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 720
	screenHeight = 1280
)

type Scene int

const (
	SceneMenu Scene = iota
	SceneLanguages
	SceneCode
	SceneProfile
	SceneShop
	SceneMiniGame
	SceneAchievements
)

type Button struct {
	X      float64
	Y      float64
	W      float64
	H      float64
	Label  string
	Active bool
	Scale  float64
}

type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   int
	Color  color.Color
	Size   float64
}

type Sparkle struct {
	X, Y   float64
	Angle  float64
	Life   int
	Size   float64
}

type Achievement struct {
	Name        string
	Description string
	Unlocked    bool
	Progress    int
	Goal        int
}

type MiniGame struct {
	Active       bool
	Stars        []Star
	Collected    int
	TimeLeft     int
	StartTime    time.Time
	Score        int
}

type Star struct {
	X, Y   float64
	Size   float64
	Angle  float64
	Speed  float64
}

type Game struct {
	scene           Scene
	score           int
	level           int
	message         string
	messageTimer    int
	lessonIndex     int
	mouseWasDown    bool
	particles       []Particle
	sparkles        []Sparkle
	achievements    []Achievement
	miniGame        MiniGame
	animationFrame  int
	selectedLanguage string
	streak          int
	comboMultiplier int
	dailyBonus      int
	lastPlayDate    string
	menuButtons     []Button
	floatOffset     float64
	shakeAmount     float64
	shakeTimer      int
	transitionAlpha float64
	transitionScene *Scene
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	
	g := &Game{
		scene:        SceneMenu,
		score:        0,
		level:        1,
		message:      "✨ ¡Bienvenida Princesa! ✨",
		lessonIndex:  0,
		particles:    []Particle{},
		sparkles:     []Sparkle{},
		achievements: initAchievements(),
		streak:       0,
		comboMultiplier: 1,
		dailyBonus:   100,
		lastPlayDate: time.Now().Format("2006-01-02"),
		floatOffset:  0,
	}
	
	g.initMenuButtons()
	g.addWelcomeSparkles()
	return g
}

func initAchievements() []Achievement {
	return []Achievement{
		{Name: "🌟 Aprendiz Estelar", Description: "Completa 5 lecciones", Unlocked: false, Progress: 0, Goal: 5},
		{Name: "💻 Reina del Código", Description: "Completa el reino de programación", Unlocked: false, Progress: 0, Goal: 10},
		{Name: "🗣️ Políglota", Description: "Completa el reino de idiomas", Unlocked: false, Progress: 0, Goal: 10},
		{Name: "🔥 Racha Mágica", Description: "Responde 3 correctas seguidas", Unlocked: false, Progress: 0, Goal: 3},
		{Name: "💎 Maestra Estrella", Description: "Alcanza nivel 5", Unlocked: false, Progress: 0, Goal: 5},
	}
}

func (g *Game) initMenuButtons() {
	g.menuButtons = []Button{
		{230, 550, 260, 70, "📚 IDIOMAS", true, 1.0},
		{230, 640, 260, 70, "💻 CODIGO", true, 1.0},
		{230, 730, 260, 70, "👸 PERFIL", true, 1.0},
		{230, 820, 260, 70, "🛍️ TIENDA", true, 1.0},
		{230, 910, 260, 70, "🏆 LOGROS", true, 1.0},
		{230, 1000, 260, 70, "🎮 MINI-JUEGO", true, 1.0},
	}
}

func (g *Game) addWelcomeSparkles() {
	for i := 0; i < 100; i++ {
		g.addSparkle(float64(rand.Intn(screenWidth)), float64(rand.Intn(screenHeight)))
	}
}

func (g *Game) addParticle(x, y float64, col color.Color) {
	g.particles = append(g.particles, Particle{
		X:    x,
		Y:    y,
		VX:   (rand.Float64() - 0.5) * 6,
		VY:   (rand.Float64() - 0.5) * 6 - 3,
		Life: 40,
		Color: col,
		Size: float64(rand.Intn(8) + 4),
	})
}

func (g *Game) addSparkle(x, y float64) {
	g.sparkles = append(g.sparkles, Sparkle{
		X:     x,
		Y:     y,
		Angle: rand.Float64() * math.Pi * 2,
		Life:  60,
		Size:  float64(rand.Intn(10) + 5),
	})
}

func (g *Game) addExplosion(x, y float64) {
	for i := 0; i < 30; i++ {
		g.addParticle(x, y, color.RGBA{255, uint8(rand.Intn(200) + 55), uint8(rand.Intn(100)), 255})
	}
	for i := 0; i < 15; i++ {
		g.addSparkle(x, y)
	}
}

func (g *Game) Update() error {
	g.updateParticles()
	g.updateSparkles()
	g.updateMessageTimer()
	g.updateShake()
	
	g.floatOffset += 0.02
	
	if g.shakeTimer > 0 {
		g.shakeTimer--
	}
	
	for i := range g.menuButtons {
		if g.menuButtons[i].Scale < 1.0 {
			g.menuButtons[i].Scale += 0.1
		}
	}
	
	if g.transitionScene != nil {
		g.transitionAlpha += 0.05
		if g.transitionAlpha >= 1 {
			g.scene = *g.transitionScene
			g.transitionScene = nil
			g.transitionAlpha = 0
		}
		return nil
	}
	
	if g.miniGame.Active {
		g.updateMiniGame()
		return nil
	}
	
	clicked, mx, my := g.getClick()

	switch g.scene {
	case SceneMenu:
		g.updateMenu(clicked, mx, my)
	case SceneLanguages:
		g.updateLesson(clicked, mx, my, data.LanguageLessons)
	case SceneCode:
		g.updateLesson(clicked, mx, my, data.CodeLessons)
	case SceneProfile:
		g.updateProfile(clicked, mx, my)
	case SceneShop:
		g.updateShop(clicked, mx, my)
	case SceneAchievements:
		g.updateAchievements(clicked, mx, my)
	}

	return nil
}

func (g *Game) updateShake() {
	if g.shakeTimer > 0 {
		g.shakeAmount = float64(g.shakeTimer) / 10
	} else {
		g.shakeAmount = 0
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		g.particles[i].X += g.particles[i].VX
		g.particles[i].Y += g.particles[i].VY
		g.particles[i].VY += 0.3
		g.particles[i].Life--
		
		if g.particles[i].Life <= 0 || g.particles[i].Y > screenHeight {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) updateSparkles() {
	for i := len(g.sparkles) - 1; i >= 0; i-- {
		g.sparkles[i].Life--
		g.sparkles[i].Angle += 0.1
		
		if g.sparkles[i].Life <= 0 {
			g.sparkles = append(g.sparkles[:i], g.sparkles[i+1:]...)
		}
	}
}

func (g *Game) updateMessageTimer() {
	if g.messageTimer > 0 {
		g.messageTimer--
	}
}

func (g *Game) getClick() (bool, float64, float64) {
	down := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	x, y := ebiten.CursorPosition()

	clicked := down && !g.mouseWasDown
	g.mouseWasDown = down

	return clicked, float64(x), float64(y)
}

func insideButton(b Button, x, y float64) bool {
	return x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H
}

func (g *Game) showMessage(msg string) {
	g.message = msg
	g.messageTimer = 90
}

func (g *Game) updateMenu(clicked bool, mx, my float64) {
	for i, btn := range g.menuButtons {
		if clicked && insideButton(btn, mx, my) {
			g.menuButtons[i].Scale = 0.95
			g.addExplosion(btn.X+btn.W/2, btn.Y+btn.H/2)
			
			switch btn.Label {
			case "📚 IDIOMAS":
				g.transitionTo(SceneLanguages)
				g.showMessage("🎉 ¡Reino de los Idiomas!")
			case "💻 CODIGO":
				g.transitionTo(SceneCode)
				g.showMessage("💫 ¡Reino del Código!")
			case "👸 PERFIL":
				g.transitionTo(SceneProfile)
				g.showMessage("✨ Tu progreso ✨")
			case "🛍️ TIENDA":
				g.transitionTo(SceneShop)
				g.showMessage("🛒 Tienda Mágica")
			case "🏆 LOGROS":
				g.transitionTo(SceneAchievements)
				g.showMessage("🏅 Tus Logros")
			case "🎮 MINI-JUEGO":
				g.startMiniGame()
			}
			return
		}
	}
}

func (g *Game) transitionTo(scene Scene) {
	g.transitionScene = &scene
	g.transitionAlpha = 0
}

func (g *Game) startMiniGame() {
	g.miniGame = MiniGame{
		Active:    true,
		Stars:     []Star{},
		Collected: 0,
		TimeLeft:  30,
		StartTime: time.Now(),
	}
	
	for i := 0; i < 25; i++ {
		g.miniGame.Stars = append(g.miniGame.Stars, Star{
			X:     float64(rand.Intn(screenWidth-100)) + 50,
			Y:     float64(rand.Intn(screenHeight-300)) + 150,
			Size:  float64(rand.Intn(20) + 15),
			Angle: rand.Float64() * math.Pi * 2,
			Speed: rand.Float64()*2 + 1,
		})
	}
	g.showMessage("⭐ ¡Atrapa las estrellas! ⭐")
}

func (g *Game) updateMiniGame() {
	elapsed := time.Since(g.miniGame.StartTime).Seconds()
	g.miniGame.TimeLeft = 30 - int(elapsed)
	
	for i := range g.miniGame.Stars {
		g.miniGame.Stars[i].Angle += 0.05
	}
	
	if g.miniGame.TimeLeft <= 0 {
		g.endMiniGame()
		return
	}
	
	clicked, mx, my := g.getClick()
	if clicked {
		for i := len(g.miniGame.Stars) - 1; i >= 0; i-- {
			star := g.miniGame.Stars[i]
			if mx >= star.X-15 && mx <= star.X+star.Size && 
			   my >= star.Y-15 && my <= star.Y+star.Size {
				g.miniGame.Stars = append(g.miniGame.Stars[:i], g.miniGame.Stars[i+1:]...)
				g.miniGame.Collected++
				g.addExplosion(star.X, star.Y)
				g.score += 50
				g.showMessage(fmt.Sprintf("⭐ +50 estrellas! %d restantes", len(g.miniGame.Stars)))
			}
		}
	}
}

func (g *Game) endMiniGame() {
	bonus := g.miniGame.Collected * 20
	g.score += bonus
	g.showMessage(fmt.Sprintf("🎉 Completaste! %d estrellas +%d bonus 🎉", 
		g.miniGame.Collected, bonus))
	g.addExplosion(screenWidth/2, screenHeight/2)
	g.miniGame.Active = false
}

func (g *Game) updateLesson(clicked bool, mx, my float64, lessons []data.Lesson) {
	if len(lessons) == 0 {
		return
	}

	lesson := lessons[g.lessonIndex]

	buttons := []Button{
		{120, 750, 480, 70, lesson.Options[0], true, 1.0},
		{120, 840, 480, 70, lesson.Options[1], true, 1.0},
		{120, 930, 480, 70, lesson.Options[2], true, 1.0},
	}

	backButton := Button{120, 1080, 480, 70, "🏠 VOLVER", true, 1.0}

	if clicked && insideButton(backButton, mx, my) {
		g.transitionTo(SceneMenu)
		g.showMessage("✨ Volviste al menú ✨")
		return
	}

	if clicked {
		for i, btn := range buttons {
			if insideButton(btn, mx, my) {
				g.answerQuestion(i, lesson)
				return
			}
		}
	}
}

func (g *Game) answerQuestion(selected int, lesson data.Lesson) {
	reward := lesson.Reward * g.comboMultiplier
	
	if selected == lesson.Answer {
		g.score += reward
		g.streak++
		g.comboMultiplier = 1 + (g.streak / 3)
		
		messages := []string{
			"⭐ ¡CORRECTO! +" + fmt.Sprint(reward) + " ⭐",
			"🎉 ¡MAGNIFICO! 🎉",
			"💫 ¡ERES UNA ESTRELLA! 💫",
		}
		g.showMessage(messages[rand.Intn(len(messages))])
		g.addExplosion(screenWidth/2, 400)
		
		if g.streak >= 3 {
			g.unlockAchievement(3)
		}
	} else {
		g.streak = 0
		g.comboMultiplier = 1
		errorMessages := []string{
			"📚 ¡SIGUE INTENTANDO!",
			"💪 ¡CASI LO LOGRASTE!",
			"✨ ¡LA PRÁCTICA HACE LA MAGIA!",
		}
		g.showMessage(errorMessages[rand.Intn(len(errorMessages))])
		g.shakeTimer = 10
	}

	g.lessonIndex++

	var lessons []data.Lesson
	if g.scene == SceneLanguages {
		lessons = data.LanguageLessons
	} else {
		lessons = data.CodeLessons
	}

	if g.lessonIndex >= len(lessons) {
		g.completeRealm()
	}
	
	g.checkAchievementProgress(lessons)
}

func (g *Game) completeRealm() {
	g.lessonIndex = 0
	g.level++
	g.streak = 0
	g.comboMultiplier = 1
	
	realmBonus := 500
	g.score += realmBonus
	
	if g.scene == SceneLanguages {
		g.unlockAchievement(2)
	} else {
		g.unlockAchievement(1)
	}
	
	g.showMessage(fmt.Sprintf("🎊 ¡REINO COMPLETADO! Nivel %d +%d 🎊", g.level, realmBonus))
	g.addExplosion(screenWidth/2, screenHeight/2)
	g.transitionTo(SceneMenu)
}

func (g *Game) checkAchievementProgress(lessons []data.Lesson) {
	if g.scene == SceneLanguages {
		g.updateAchievementProgress(2, g.lessonIndex)
	} else {
		g.updateAchievementProgress(1, g.lessonIndex)
	}
	
	g.updateAchievementProgress(0, g.lessonIndex)
	
	if g.level >= 5 {
		g.unlockAchievement(4)
	}
}

func (g *Game) updateAchievementProgress(index int, progress int) {
	if index < len(g.achievements) && !g.achievements[index].Unlocked {
		g.achievements[index].Progress = progress
		if g.achievements[index].Progress >= g.achievements[index].Goal {
			g.unlockAchievement(index)
		}
	}
}

func (g *Game) unlockAchievement(index int) {
	if index < len(g.achievements) && !g.achievements[index].Unlocked {
		g.achievements[index].Unlocked = true
		bonus := 200
		g.score += bonus
		g.showMessage(fmt.Sprintf("🏆 ¡%s! +%d 🏆", g.achievements[index].Name, bonus))
		g.addExplosion(screenWidth/2, screenHeight/2)
	}
}

func (g *Game) updateProfile(clicked bool, mx, my float64) {
	buttons := []Button{
		{120, 1000, 480, 70, "🎁 BONUS DIARIO", g.canClaimDailyBonus(), 1.0},
		{120, 1090, 480, 70, "🏠 VOLVER", true, 1.0},
	}

	for _, btn := range buttons {
		if clicked && insideButton(btn, mx, my) && btn.Active {
			if btn.Label == "🎁 BONUS DIARIO" {
				g.claimDailyBonus()
			} else {
				g.transitionTo(SceneMenu)
				g.showMessage("✨ Volviste al menú ✨")
			}
			return
		}
	}
}

func (g *Game) canClaimDailyBonus() bool {
	today := time.Now().Format("2006-01-02")
	return g.lastPlayDate != today
}

func (g *Game) claimDailyBonus() {
	if g.canClaimDailyBonus() {
		g.score += g.dailyBonus
		g.lastPlayDate = time.Now().Format("2006-01-02")
		g.showMessage(fmt.Sprintf("🎁 +%d estrellas! 🎁", g.dailyBonus))
		g.dailyBonus += 25
		if g.dailyBonus > 500 {
			g.dailyBonus = 500
		}
		g.addExplosion(screenWidth/2, screenHeight/2)
	}
}

func (g *Game) updateShop(clicked bool, mx, my float64) {
	backBtn := Button{120, 1100, 480, 70, "🏠 VOLVER", true, 1.0}
	
	if clicked && insideButton(backBtn, mx, my) {
		g.transitionTo(SceneMenu)
		g.showMessage("✨ Volviste al menú ✨")
		return
	}
}

func (g *Game) updateAchievements(clicked bool, mx, my float64) {
	backBtn := Button{120, 1100, 480, 70, "🏠 VOLVER", true, 1.0}
	
	if clicked && insideButton(backBtn, mx, my) {
		g.transitionTo(SceneMenu)
		g.showMessage("✨ Volviste al menú ✨")
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawGradientBackground(screen)
	g.drawOrnaments(screen)
	
	if g.miniGame.Active {
		g.drawMiniGame(screen)
	} else {
		switch g.scene {
		case SceneMenu:
			g.drawTitle(screen)
			g.drawMenuButtons(screen)
		case SceneLanguages:
			g.drawLessonScreen(screen, "📚 REINO DE LOS IDIOMAS", data.LanguageLessons)
		case SceneCode:
			g.drawLessonScreen(screen, "💻 REINO DEL CÓDIGO", data.CodeLessons)
		case SceneProfile:
			g.drawProfileScreen(screen)
		case SceneShop:
			g.drawShopScreen(screen)
		case SceneAchievements:
			g.drawAchievementsScreen(screen)
		}
	}
	
	g.drawHeader(screen)
	g.drawParticles(screen)
	g.drawSparkles(screen)
	
	if g.messageTimer > 0 {
		g.drawMessage(screen)
	}
	
	if g.transitionScene != nil {
		g.drawTransition(screen)
	}
}

func (g *Game) drawGradientBackground(screen *ebiten.Image) {
	for i := 0; i < screenHeight; i++ {
		ratio := float64(i) / float64(screenHeight)
		r := uint8(120 + float64(80)*math.Sin(ratio*math.Pi))
		gVal := uint8(80 + float64(100)*math.Sin(ratio*math.Pi))
		b := uint8(180 + float64(75)*math.Cos(ratio*math.Pi))
		ebitenutil.DrawRect(screen, 0, float64(i), screenWidth, 1, color.RGBA{r, gVal, b, 255})
	}
	
	// Estrellas de fondo
	for i := 0; i < 50; i++ {
		x := float64((i * 131) % screenWidth)
		y := float64((i * 253) % screenHeight)
		alpha := uint8(100 + rand.Intn(155))
		ebitenutil.DrawRect(screen, x, y, 2, 2, color.RGBA{255, 255, 200, alpha})
	}
}

func (g *Game) drawOrnaments(screen *ebiten.Image) {
	// Flores decorativas en las esquinas
	flowerColor := color.RGBA{255, 180, 200, 100}
	for i := 0; i < 8; i++ {
		angle := float64(i) * math.Pi * 2 / 8
		x := 50 + math.Cos(angle+float64(g.animationFrame)/30)*20
		y := 50 + math.Sin(angle+float64(g.animationFrame)/30)*20
		ebitenutil.DrawRect(screen, x, y, 8, 8, flowerColor)
		
		x2 := float64(screenWidth-50) + math.Cos(angle+float64(g.animationFrame)/30)*20
		y2 := 50 + math.Sin(angle+float64(g.animationFrame)/30)*20
		ebitenutil.DrawRect(screen, x2, y2, 8, 8, flowerColor)
	}
	g.animationFrame++
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	titleY := 200 + math.Sin(g.floatOffset)*10
	
	// Sombra del título
	ebitenutil.DebugPrintAt(screen, "✨ STRELLAS ✨", 240, int(titleY+5))
	ebitenutil.DebugPrintAt(screen, "PRINCESA DEL SABER", 250, int(titleY+55))
	
	// Título principal con colores brillantes
	titleColor := color.RGBA{255, 215, 0, 255}
	ebitenutil.DrawRect(screen, 180, titleY-10, 360, 80, titleColor)
	ebitenutil.DebugPrintAt(screen, "✨ STRELLAS ✨", 240, int(titleY))
	ebitenutil.DebugPrintAt(screen, "PRINCESA DEL SABER", 250, int(titleY+50))
}

func (g *Game) drawMenuButtons(screen *ebiten.Image) {
	for i, btn := range g.menuButtons {
		// Efecto de brillo
		alpha := uint8(150 + int(math.Sin(float64(g.animationFrame)/10+float64(i))*100))
		ebitenutil.DrawRect(screen, btn.X-2, btn.Y-2, btn.W+4, btn.H+4, color.RGBA{255, 255, 200, alpha})
		
		// Botón principal
		btnColor := color.RGBA{255, 100 + uint8(i*20), 150 + uint8(i*10), 255}
		ebitenutil.DrawRect(screen, btn.X, btn.Y, btn.W, btn.H, btnColor)
		
		// Borde dorado
		ebitenutil.DrawRect(screen, btn.X, btn.Y, btn.W, 4, color.RGBA{255, 215, 0, 255})
		ebitenutil.DrawRect(screen, btn.X, btn.Y+btn.H-4, btn.W, 4, color.RGBA{255, 215, 0, 255})
		
		// Icono decorativo
		ebitenutil.DebugPrintAt(screen, btn.Label, int(btn.X+80), int(btn.Y+30))
	}
}

func (g *Game) drawLessonScreen(screen *ebiten.Image, title string, lessons []data.Lesson) {
	if len(lessons) == 0 {
		return
	}
	
	lesson := lessons[g.lessonIndex]
	
	// Marco decorativo
	ebitenutil.DrawRect(screen, 40, 300, 640, 920, color.RGBA{255, 245, 255, 230})
	ebitenutil.DrawRect(screen, 40, 300, 640, 10, color.RGBA{255, 170, 70, 255})
	ebitenutil.DrawRect(screen, 40, 1210, 640, 10, color.RGBA{255, 170, 70, 255})
	
	// Título con decoración
	ebitenutil.DebugPrintAt(screen, title, 220, 330)
	
	// Pregunta
	ebitenutil.DrawRect(screen, 80, 400, 560, 200, color.RGBA{255, 230, 200, 255})
	ebitenutil.DebugPrintAt(screen, "📝 "+lesson.Question, 100, 430)
	
	// Opciones
	buttons := []Button{
		{120, 660, 480, 70, lesson.Options[0], true, 1.0},
		{120, 750, 480, 70, lesson.Options[1], true, 1.0},
		{120, 840, 480, 70, lesson.Options[2], true, 1.0},
	}
	
	for i, btn := range buttons {
		btnColor := color.RGBA{200 + uint8(i*20), 180 + uint8(i*30), 255, 255}
		ebitenutil.DrawRect(screen, btn.X, btn.Y, btn.W, btn.H, btnColor)
		ebitenutil.DebugPrintAt(screen, btn.Label, int(btn.X+30), int(btn.Y+30))
	}
	
	// Botón volver
	ebitenutil.DrawRect(screen, 120, 1100, 480, 70, color.RGBA{200, 180, 255, 255})
	ebitenutil.DebugPrintAt(screen, "🏠 VOLVER", 280, 1130)
}

func (g *Game) drawProfileScreen(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 50, 280, 620, 950, color.RGBA{255, 245, 255, 230})
	ebitenutil.DrawRect(screen, 50, 280, 620, 10, color.RGBA{255, 170, 70, 255})
	
	ebitenutil.DebugPrintAt(screen, "👸 PERFIL DE LA PRINCESA", 220, 320)
	
	// Avatar decorativo
	ebitenutil.DrawRect(screen, 260, 380, 200, 200, color.RGBA{255, 210, 230, 255})
	ebitenutil.DebugPrintAt(screen, "✨", 345, 470)
	
	y := 630
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("🌟 Nombre: Princesa Strella"), 140, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("👑 Nivel: %d", g.level), 140, y+50)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("⭐ Estrellas: %d", g.score), 140, y+100)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("⚡ Racha: %d", g.streak), 140, y+150)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("🔥 Combo: x%d", g.comboMultiplier), 140, y+200)
	
	// Botones
	bonusText := "🎁 BONUS DIARIO"
	if !g.canClaimDailyBonus() {
		bonusText = "✅ BONUS RECLAMADO"
	}
	ebitenutil.DrawRect(screen, 120, 950, 480, 70, color.RGBA{255, 200, 100, 255})
	ebitenutil.DebugPrintAt(screen, bonusText, 250, 980)
	
	ebitenutil.DrawRect(screen, 120, 1050, 480, 70, color.RGBA{200, 180, 255, 255})
	ebitenutil.DebugPrintAt(screen, "🏠 VOLVER", 310, 1080)
}

func (g *Game) drawShopScreen(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 50, 280, 620, 950, color.RGBA{255, 245, 255, 230})
	ebitenutil.DrawRect(screen, 50, 280, 620, 10, color.RGBA{255, 170, 70, 255})
	
	ebitenutil.DebugPrintAt(screen, "🛒 TIENDA MÁGICA", 250, 320)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("⭐ Tus estrellas: %d ⭐", g.score), 260, 370)
	
	items := []struct {
		icon string
		name string
		cost int
		y    int
	}{
		{"🔮", "Varita Mágica", 1000, 450},
		{"👑", "Corona Real", 2000, 550},
		{"📚", "Libro de Hechizos", 1500, 650},
		{"💎", "Gema Estelar", 3000, 750},
	}
	
	for _, item := range items {
		ebitenutil.DrawRect(screen, 100, float64(item.y), 520, 80, color.RGBA{200, 180, 255, 255})
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s %s - %d⭐", item.icon, item.name, item.cost), 
			150, item.y+35)
	}
	
	ebitenutil.DrawRect(screen, 120, 1100, 480, 70, color.RGBA{200, 180, 255, 255})
	ebitenutil.DebugPrintAt(screen, "🏠 VOLVER", 310, 1130)
}

func (g *Game) drawAchievementsScreen(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 30, 280, 660, 950, color.RGBA{255, 245, 255, 230})
	ebitenutil.DrawRect(screen, 30, 280, 660, 10, color.RGBA{255, 170, 70, 255})
	
	ebitenutil.DebugPrintAt(screen, "🏆 LOGROS ESPECIALES", 240, 320)
	
	y := 380
	for _, ach := range g.achievements {
		status := "🔒"
		if ach.Unlocked {
			status = "✅"
		}
		
		text := fmt.Sprintf("%s %s", status, ach.Name)
		if !ach.Unlocked {
			text += fmt.Sprintf(" (%d/%d)", ach.Progress, ach.Goal)
		}
		
		ebitenutil.DrawRect(screen, 60, float64(y), 600, 60, color.RGBA{240, 230, 255, 255})
		ebitenutil.DebugPrintAt(screen, text, 80, y+25)
		ebitenutil.DebugPrintAt(screen, ach.Description, 80, y+45)
		y += 80
	}
	
	ebitenutil.DrawRect(screen, 120, 1100, 480, 70, color.RGBA{200, 180, 255, 255})
	ebitenutil.DebugPrintAt(screen, "🏠 VOLVER", 310, 1130)
}

func (g *Game) drawMiniGame(screen *ebiten.Image) {
	// Fondo oscuro con estrellas
	for i := 0; i < screenHeight; i++ {
		ratio := float64(i) / float64(screenHeight)
		r := uint8(50 + float64(30)*ratio)
		gVal := uint8(30 + float64(40)*ratio)
		b := uint8(80 + float64(50)*ratio)
		ebitenutil.DrawRect(screen, 0, float64(i), screenWidth, 1, color.RGBA{r, gVal, b, 255})
	}
	
	// Título
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("⭐ MINI-JUEGO ⭐ Tiempo: %d", g.miniGame.TimeLeft), 230, 100)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Atrapadas: %d", g.miniGame.Collected), 290, 150)
	
	// Estrellas
	for idx := range g.miniGame.Stars {
		star := &g.miniGame.Stars[idx]
		// Brillo pulsante
		alpha := uint8(150 + int(math.Sin(float64(g.animationFrame)/5)*100))
		
		// Estrella con rotación
		star.X += math.Cos(star.Angle) * star.Speed
		star.Y += math.Sin(star.Angle) * star.Speed
		star.Angle += 0.05
		
		ebitenutil.DrawRect(screen, star.X, star.Y, star.Size, star.Size, color.RGBA{255, 215, 0, alpha})
		ebitenutil.DrawRect(screen, star.X+star.Size/3, star.Y-star.Size/3, star.Size/3, star.Size*1.5, color.RGBA{255, 215, 0, alpha})
		ebitenutil.DrawRect(screen, star.X-star.Size/3, star.Y+star.Size/3, star.Size*1.5, star.Size/3, color.RGBA{255, 215, 0, alpha})
		
		// Reaparecer si salen de la pantalla
		if star.X < -50 || star.X > screenWidth+50 || star.Y < -50 || star.Y > screenHeight+50 {
			star.X = float64(rand.Intn(screenWidth-100)) + 50
			star.Y = float64(rand.Intn(screenHeight-300)) + 150
		}
	}
	
	if g.miniGame.TimeLeft <= 0 {
		ebitenutil.DrawRect(screen, 200, 500, 320, 100, color.RGBA{0, 0, 0, 200})
		ebitenutil.DebugPrintAt(screen, "¡TIEMPO!", 310, 540)
		ebitenutil.DebugPrintAt(screen, "Presiona para continuar", 260, 570)
	}
}

func (g *Game) drawHeader(screen *ebiten.Image) {
	// Header transparente
	ebitenutil.DrawRect(screen, 0, 0, screenWidth, 140, color.RGBA{0, 0, 0, 100})
	
	// Decoración header
	for i := 0; i < 5; i++ {
		x := float64(i*150 + g.animationFrame%150)
		ebitenutil.DrawRect(screen, x, 140, 50, 4, color.RGBA{255, 215, 0, 200})
	}
	
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("👑 Nivel %d", g.level), 30, 30)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("⭐ %d ⭐", g.score), 30, 70)
	
	if g.comboMultiplier > 1 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("🔥 x%d", g.comboMultiplier), 30, 110)
	}
}

func (g *Game) drawMessage(screen *ebiten.Image) {
	alpha := uint8(200)
	if g.messageTimer < 30 {
		alpha = uint8(float64(g.messageTimer) / 30 * 200)
	}
	
	// Marco del mensaje
	ebitenutil.DrawRect(screen, 60, 160, 600, 60, color.RGBA{0, 0, 0, alpha})
	ebitenutil.DrawRect(screen, 60, 160, 600, 4, color.RGBA{255, 215, 0, alpha})
	ebitenutil.DrawRect(screen, 60, 216, 600, 4, color.RGBA{255, 215, 0, alpha})
	
	ebitenutil.DebugPrintAt(screen, g.message, 100, 185)
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		alpha := uint8(float64(p.Life) / 40 * 255)
		r, gVal, b, _ := p.Color.RGBA()
		ebitenutil.DrawRect(screen, p.X, p.Y, p.Size, p.Size, color.RGBA{uint8(r >> 8), uint8(gVal >> 8), uint8(b >> 8), alpha})
	}
}

func (g *Game) drawSparkles(screen *ebiten.Image) {
	for _, s := range g.sparkles {
		alpha := uint8(float64(s.Life) / 60 * 255)
		radius := s.Size * (1 - float64(s.Life)/60)
		
		for i := 0; i < 4; i++ {
			angle := s.Angle + float64(i)*math.Pi/2
			x := s.X + math.Cos(angle)*radius
			y := s.Y + math.Sin(angle)*radius
			ebitenutil.DrawRect(screen, x-2, y-2, 4, 4, color.RGBA{255, 255, 200, alpha})
		}
	}
}

func (g *Game) drawTransition(screen *ebiten.Image) {
	alpha := uint8(g.transitionAlpha * 255)
	ebitenutil.DrawRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, alpha})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(360, 640)
	ebiten.SetWindowTitle("✨ Strellas: Princesa del Saber ✨")
	ebiten.SetWindowResizable(true)
	ebiten.SetTPS(60)
	
	game := NewGame()
	
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}