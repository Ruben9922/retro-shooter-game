package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"math/rand/v2"
	"strings"
)

// todo: decide on value properly later
var gameViewSize = vector2d{x: 50, y: 15}

const enemySpacing = 1
const enemyColumnCount = 10
const playerMoveIncrement = 2
const scorePerEnemyHit = 100
const scorePerBulletHit = 50

type vector2dMap map[int]map[int]struct{}

type status int

const (
	playing status = iota
	paused
	gameLost
	gameWon
	lifeLost
)

type gameView struct {
	playerPosition    vector2d
	enemyPositions    vector2dMap
	enemyYOffset      int // Easier to just store this instead of traversing through the map to find the min or max y value
	tickCount         int
	playerBullets     []vector2d
	score             int
	status            status
	enemyBullets      []vector2d
	livesRemaining    int
	lifeLostTickCount int
}

func newGameView() *gameView {
	return &gameView{
		playerPosition:    vector2d{x: gameViewSize.x / 2, y: gameViewSize.y - 1},
		enemyPositions:    generateEnemyPositions(),
		tickCount:         0,
		playerBullets:     make([]vector2d, 0),
		score:             0,
		enemyBullets:      make([]vector2d, 0),
		livesRemaining:    3,
		status:            playing,
		lifeLostTickCount: 0,
	}
}

func generateEnemyPositions() (enemyPositions vector2dMap) {
	const rowCount = 5
	enemyPositions = make(vector2dMap, rowCount)
	for i := 0; i < rowCount; i++ {
		enemyPositions[i] = make(map[int]struct{}, enemyColumnCount)

		for columnIndex := 0; columnIndex < enemyColumnCount; columnIndex++ {
			var x int
			if i%2 == 0 {
				x = columnIndex * 2
			} else {
				x = gameViewSize.x - (columnIndex * 2) - 1
			}
			enemyPositions[i][x] = struct{}{}
		}
	}
	return
}

func (gv *gameView) update(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch gv.status {
		case gameLost, gameWon:
			if msg.String() == "enter" {
				m.view = newGameView()
			}
			return m, tea.Batch(bulletTickCmd(), enemyTickCmd())
		case paused:
			if msg.String() == "p" {
				gv.status = playing
			}

			return m, tea.Batch(bulletTickCmd(), enemyTickCmd())
		case playing:
			switch msg.String() {
			case "left", "a":
				if gv.playerPosition.x > 1 {
					gv.playerPosition.x -= playerMoveIncrement
				}
			case "right", "d":
				if gv.playerPosition.x < gameViewSize.x-2 {
					gv.playerPosition.x += playerMoveIncrement
				}
			case " ":
				gv.createPlayerBullet()

				return m, nil
			case "p":
				gv.status = paused
			}
		case lifeLost:
		}
	case bulletTickMsg:
		if gv.status == playing {
			gv.updatePlayerBullets()
			gv.updateEnemyBullets()
			gv.handlePlayerBulletCollisions()
			gv.handleEnemyBulletCollisions()
			gv.handleBulletCollisions()

			if gv.status == lifeLost && gv.lifeLostTickCount == 0 {
				return m, tea.Batch(bulletTickCmd(), lifeLostTickCmd())
			}
			return m, bulletTickCmd()
		}
	case enemyTickMsg:
		if gv.status == playing {
			gv.updateEnemies()
			gv.handlePlayerBulletCollisions()
			gv.createEnemyBullets()

			return m, enemyTickCmd()
		}
	case lifeLostTickMsg:
		gv.lifeLostTickCount++

		if gv.lifeLostTickCount < 6 {
			return m, lifeLostTickCmd()
		} else {
			if gv.livesRemaining <= 0 {
				gv.status = gameLost
			} else {
				gv.status = playing
			}
			gv.lifeLostTickCount = 0
			return m, tea.Batch(bulletTickCmd(), enemyTickCmd())
		}
	}
	return m, nil
}

func (gv *gameView) createPlayerBullet() {
	newBullet := vector2d{
		x: gv.playerPosition.x,
		y: gameViewSize.y - 2,
	}
	gv.playerBullets = append(gv.playerBullets, newBullet)
}

func (gv *gameView) updateEnemies() {
	// If either end of the row is reached...
	if gv.tickCount >= gameViewSize.x-enemyColumnCount-(enemySpacing*(enemyColumnCount-1)) {
		gv.tickCount = 0

		if gv.enemyYOffset+len(gv.enemyPositions)-1 >= gameViewSize.y-2 {
			// If enemies have reached the bottom of the screen then it's game over
			gv.status = gameLost
		} else {
			// Else move enemies down
			updatedEnemyPositions := make(vector2dMap, len(gv.enemyPositions))
			for y, xMap := range gv.enemyPositions {
				for x := range xMap {
					position := vector2d{x: x, y: y}

					updatedPosition := position
					updatedPosition.y++

					if _, isYPresent := updatedEnemyPositions[updatedPosition.y]; !isYPresent {
						updatedEnemyPositions[updatedPosition.y] = make(map[int]struct{}, len(gv.enemyPositions[position.y]))
					}

					updatedEnemyPositions[updatedPosition.y][updatedPosition.x] = struct{}{}
				}
			}
			gv.enemyYOffset++
			gv.enemyPositions = updatedEnemyPositions
		}
	} else {
		gv.tickCount++

		// Move enemies left/right (move alternate rows in opposite directions, so 1st row left/right, then 2nd row right/left, etc.)
		updatedEnemyPositions := make(vector2dMap, len(gv.enemyPositions))
		for y, xMap := range gv.enemyPositions {
			for x := range xMap {
				position := vector2d{x: x, y: y}

				updatedPosition := position
				if updatedPosition.y%2 == 0 {
					updatedPosition.x++
				} else {
					updatedPosition.x--
				}

				if _, isYPresent := updatedEnemyPositions[updatedPosition.y]; !isYPresent {
					updatedEnemyPositions[updatedPosition.y] = make(map[int]struct{}, len(gv.enemyPositions[position.y]))
				}

				updatedEnemyPositions[updatedPosition.y][updatedPosition.x] = struct{}{}
			}
		}
		gv.enemyPositions = updatedEnemyPositions
	}
}

func (gv *gameView) updatePlayerBullets() {
	updatedPositions := make([]vector2d, 0, len(gv.playerBullets))
	for _, position := range gv.playerBullets {
		position.y--

		if isPositionValid(position) {
			updatedPositions = append(updatedPositions, position)
		}
	}
	gv.playerBullets = updatedPositions
}

func (gv *gameView) updateEnemyBullets() {
	updatedPositions := make([]vector2d, 0, len(gv.enemyBullets))
	for _, position := range gv.enemyBullets {
		position.y++

		if isPositionValid(position) {
			updatedPositions = append(updatedPositions, position)
		}
	}
	gv.enemyBullets = updatedPositions
}

func (gv *gameView) createEnemyBullets() {
	// On every tick, randomly decide whether *any* enemy should shoot a bullet
	// If so, then randomly pick an enemy that will shoot
	// Probability of any enemy shooting a bullet is proportional to the number of enemies
	// Otherwise the enemies will appear more aggressive as more of them are killed
	if rand.IntN(3) == 0 {
		if bullet := gv.createEnemyBullet(); bullet != emptyVector2d {
			gv.enemyBullets = append(gv.enemyBullets, bullet)
		}
	}
}

func (gv *gameView) createEnemyBullet() vector2d {
	// Convert map to slice of keys (y values), then randomly pick one
	ys := make([]int, 0, len(gv.enemyPositions))
	for y, xs := range gv.enemyPositions {
		containsValidXValues := false
		for x := range xs {
			// Currently the player can only move in increments of 2, so only pick positions that would actually hit the player
			if gv.canCollideWithPlayer(x) {
				containsValidXValues = true
				break
			}
		}

		if containsValidXValues {
			ys = append(ys, y)
		}
	}

	if len(ys) == 0 {
		return emptyVector2d
	}

	yIndex := rand.IntN(len(ys))
	y := ys[yIndex]

	// Convert map to slice of keys (x values), then randomly pick one
	xs := make([]int, 0, len(gv.enemyPositions[y]))
	for x := range gv.enemyPositions[y] {
		// Currently the player can only move in increments of 2, so only pick positions that would actually hit the player
		if gv.canCollideWithPlayer(x) {
			xs = append(xs, x)
		}
	}

	xIndex := rand.IntN(len(xs))
	x := xs[xIndex]

	enemyBulletPosition := vector2d{x: x, y: y}

	return enemyBulletPosition
}

func (gv *gameView) canCollideWithPlayer(x int) bool {
	return (x % playerMoveIncrement) == (gv.playerPosition.x % playerMoveIncrement)
}

// Handle collisions between player bullets and enemies
// Remove player bullet and enemy
func (gv *gameView) handlePlayerBulletCollisions() {
	updatedBulletPositions := make([]vector2d, 0, len(gv.playerBullets))
	for _, position := range gv.playerBullets {
		collision := gv.enemyPositions.checkIfPresent(position)
		if collision {
			gv.enemyPositions.delete(position)

			gv.score += scorePerEnemyHit

			if gv.enemyPositions.count() == 0 {
				gv.status = gameWon
			}
		} else {
			updatedBulletPositions = append(updatedBulletPositions, position)
		}
	}

	gv.playerBullets = updatedBulletPositions
}

// Handle collisions between enemy bullets and player
// Remove the enemy bullet and decrement lives
func (gv *gameView) handleEnemyBulletCollisions() {
	updatedBulletPositions := make([]vector2d, 0, len(gv.enemyBullets))
	for _, bulletPosition := range gv.enemyBullets {
		if gv.playerPosition == bulletPosition {
			gv.livesRemaining--
			gv.status = lifeLost
		} else {
			updatedBulletPositions = append(updatedBulletPositions, bulletPosition)
		}
	}
	gv.enemyBullets = updatedBulletPositions
}

// Handle collisions between enemy bullets and player bullets
// Remove both bullets (so player can shoot the enemy bullets to destroy them)
func (gv *gameView) handleBulletCollisions() {
	playerBulletsMap := vectorSliceToMap(gv.playerBullets)
	enemyBulletsMap := vectorSliceToMap(gv.enemyBullets)

	// Iterate through one of the maps and remove any matching items from both maps
	for y, xMap := range playerBulletsMap {
		for x := range xMap {
			playerBullet := vector2d{x: x, y: y}
			if enemyBulletsMap.checkIfPresent(playerBullet) {
				playerBulletsMap.delete(playerBullet)
				enemyBulletsMap.delete(playerBullet)
				gv.score += scorePerBulletHit
			}

			// Note that bullets with an even vertical gap won't actually collide on the same point
			// Hence also checking the point above `playerBullet`
			pointAbovePlayerBullet := vector2d{x: playerBullet.x, y: playerBullet.y + 1}
			if enemyBulletsMap.checkIfPresent(pointAbovePlayerBullet) {
				playerBulletsMap.delete(playerBullet)
				enemyBulletsMap.delete(pointAbovePlayerBullet)
				gv.score += scorePerBulletHit
			}
		}
	}

	gv.playerBullets = playerBulletsMap.toSlice()
	gv.enemyBullets = enemyBulletsMap.toSlice()
}

func vectorSliceToMap(s []vector2d) (m vector2dMap) {
	m = make(vector2dMap)
	for _, bullet := range s {
		if _, present := m[bullet.y]; !present {
			m[bullet.y] = make(map[int]struct{})
		}
		m[bullet.y][bullet.x] = struct{}{}
	}
	return
}

func (m vector2dMap) toSlice() (s []vector2d) {
	s = make([]vector2d, 0)
	for y, xMap := range m {
		for x := range xMap {
			s = append(s, vector2d{x: x, y: y})
		}
	}
	return
}

func (m vector2dMap) delete(v vector2d) {
	delete(m[v.y], v.x)

	// If the inner map is empty then delete corresponding entry in outer map as no longer needed
	if len(m[v.y]) == 0 {
		delete(m, v.y)
	}
}

func (m vector2dMap) checkIfPresent(v vector2d) bool {
	xMap, yPresent := m[v.y]
	if !yPresent {
		return false
	}

	_, xPresent := xMap[v.x]
	return xPresent
}

func (m vector2dMap) count() (count int) {
	count = 0
	for _, xMap := range m {
		count += len(xMap)
	}
	return
}

func (gv *gameView) draw(model) string {
	border := lipgloss.RoundedBorder()
	style := lipgloss.NewStyle().
		BorderForeground(accentColor).
		BorderStyle(border).
		Padding(0, 1)

	outputMatrix := newOutputMatrix()
	gv.drawEnemies(&outputMatrix)
	gv.drawPlayerBullets(&outputMatrix)
	gv.drawEnemyBullets(&outputMatrix)
	gv.drawPlayer(&outputMatrix)

	mainString := outputMatrixToString(outputMatrix)
	scoreString := fmt.Sprintf("Score: %d", gv.score)
	livesString := fmt.Sprintf("Lives: %d", gv.livesRemaining)
	statusString := gv.getStatusString()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.Render(mainString),
		fmt.Sprintf("%s; %s", scoreString, livesString),
		lipgloss.NewStyle().PaddingTop(1).Render(statusString),
	)
}

func (gv *gameView) getStatusString() string {
	switch gv.status {
	case gameLost:
		return "Game over! Press Enter to restart..."
	case paused:
		return "Paused; press P to resume..."
	case gameWon:
		return "You win! All enemies destroyed!\nPress Enter to restart..."
	case lifeLost:
		return "Lost a life!"
	default:
		return ""
	}
}

func newOutputMatrix() (outputMatrix [][]rune) {
	outputMatrix = make([][]rune, gameViewSize.y)
	for i := range outputMatrix {
		outputMatrix[i] = make([]rune, gameViewSize.x)
	}
	return
}

func (gv *gameView) drawEnemies(outputMatrix *[][]rune) {
	for y, xMap := range gv.enemyPositions {
		for x := range xMap {
			(*outputMatrix)[y][x] = '$'
		}
	}
}

func (gv *gameView) drawPlayerBullets(outputMatrix *[][]rune) {
	for _, position := range gv.playerBullets {
		(*outputMatrix)[position.y][position.x] = '.'
	}
}

func (gv *gameView) drawEnemyBullets(outputMatrix *[][]rune) {
	for _, position := range gv.enemyBullets {
		(*outputMatrix)[position.y][position.x] = '.'
	}
}

func (gv *gameView) drawPlayer(outputMatrix *[][]rune) {
	var playerRune rune
	if gv.status != lifeLost || gv.lifeLostTickCount%2 == 0 {
		playerRune = '*'
	} else {
		playerRune = ' '
	}
	(*outputMatrix)[gv.playerPosition.y][gv.playerPosition.x] = playerRune
}

func outputMatrixToString(outputMatrix [][]rune) string {
	var sb strings.Builder
	for y, outputMatrixRow := range outputMatrix {
		for _, outputMatrixRune := range outputMatrixRow {
			if outputMatrixRune == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(outputMatrixRune)
			}
		}

		if y < len(outputMatrix)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func isPositionValid(position vector2d) bool {
	return position.x >= 0 && position.x < gameViewSize.x && position.y >= 0 && position.y < gameViewSize.y
}
