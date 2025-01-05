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

type gameView struct {
	playerPosition       vector2d
	enemyPositions       map[int]map[int]struct{}
	enemyYOffset         int // Easier to just store this instead of traversing through the map to find the min y value
	tickCount            int
	bulletPositions      []vector2d
	score                int
	gameOver             bool
	paused               bool
	enemyBulletPositions []vector2d
	livesRemaining       int
}

func newGameView() *gameView {
	return &gameView{
		playerPosition:       vector2d{x: gameViewSize.x / 2, y: gameViewSize.y - 1},
		enemyPositions:       generateEnemyPositions(),
		tickCount:            0,
		bulletPositions:      make([]vector2d, 0),
		score:                0,
		enemyBulletPositions: make([]vector2d, 0),
		livesRemaining:       3,
	}
}

func generateEnemyPositions() (enemyPositions map[int]map[int]struct{}) {
	const rowCount = 5
	enemyPositions = make(map[int]map[int]struct{}, rowCount)
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
		if gv.gameOver {
			if msg.String() == "enter" {
				m.view = newGameView()
			}
			return m, tea.Batch(bulletTickCmd(), enemyTickCmd())
		}

		if gv.paused {
			if msg.String() == "p" {
				gv.paused = false
			}

			return m, tea.Batch(bulletTickCmd(), enemyTickCmd())
		}

		switch msg.String() {
		case "left", "a":
			if gv.playerPosition.x > 1 {
				gv.playerPosition.x -= enemySpacing + 1
			}
		case "right", "d":
			if gv.playerPosition.x < gameViewSize.x-2 {
				gv.playerPosition.x += enemySpacing + 1
			}
		case " ":
			newBullet := vector2d{
				x: gv.playerPosition.x,
				y: gameViewSize.y - 2,
			}
			gv.bulletPositions = append(gv.bulletPositions, newBullet)

			return m, nil
		case "p":
			gv.paused = true
		}
	case bulletTickMsg:
		if !gv.gameOver && !gv.paused {
			gv.updateBullets()
			gv.updateEnemyBullets()
			gv.handleCollisions()
			gv.handleEnemyBulletCollisions()

			return m, bulletTickCmd()
		}
	case enemyTickMsg:
		if !gv.gameOver && !gv.paused {
			gv.updateEnemies()
			gv.handleCollisions()
			gv.createEnemyBulletPositions()

			return m, enemyTickCmd()
		}
	}
	return m, nil
}

func (gv *gameView) updateEnemies() {
	// If either end of the row is reached...
	if gv.tickCount >= gameViewSize.x-enemyColumnCount-(enemySpacing*(enemyColumnCount-1)) {
		gv.tickCount = 0

		if gv.enemyYOffset+len(gv.enemyPositions)-1 >= gameViewSize.y-2 {
			// If enemies have reached the bottom of the screen then it's game over
			gv.gameOver = true
		} else {
			// Else move enemies down
			updatedEnemyPositions := make(map[int]map[int]struct{}, len(gv.enemyPositions))
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
		updatedEnemyPositions := make(map[int]map[int]struct{}, len(gv.enemyPositions))
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

func (gv *gameView) updateBullets() {
	updatedPositions := make([]vector2d, 0, len(gv.bulletPositions))
	for _, position := range gv.bulletPositions {
		position.y--

		if isPositionValid(position) {
			updatedPositions = append(updatedPositions, position)
		}
	}
	gv.bulletPositions = updatedPositions
}

func (gv *gameView) updateEnemyBullets() {
	updatedPositions := make([]vector2d, 0, len(gv.enemyBulletPositions))
	for _, position := range gv.enemyBulletPositions {
		position.y++

		if isPositionValid(position) {
			updatedPositions = append(updatedPositions, position)
		}
	}
	gv.enemyBulletPositions = updatedPositions
}

func (gv *gameView) createEnemyBulletPositions() {
	// On every tick, randomly decide whether *any* enemy should shoot a bullet
	// If so, then randomly pick an enemy that will shoot
	// Probability of any enemy shooting a bullet is proportional to the number of enemies
	// Otherwise the enemies will appear more aggressive as more of them are killed
	if rand.IntN(3) == 0 {
		if bullet := gv.createEnemyBullet(); bullet != emptyVector2d {
			gv.enemyBulletPositions = append(gv.enemyBulletPositions, bullet)
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
	playerMoveIncrement := enemySpacing + 1
	return (x % playerMoveIncrement) == (gv.playerPosition.x % playerMoveIncrement)
}

func (gv *gameView) handleCollisions() {
	updatedBulletPositions := make([]vector2d, 0, len(gv.bulletPositions))
	for _, position := range gv.bulletPositions {
		xMap, yIsPresent := gv.enemyPositions[position.y]
		_, xIsPresent := xMap[position.x]
		collision := yIsPresent && xIsPresent
		if collision {
			delete(xMap, position.x)

			// If the inner map is empty then delete corresponding entry in outer map as no longer needed
			if len(xMap) == 0 {
				delete(gv.enemyPositions, position.y)
			}

			gv.score++
		} else {
			updatedBulletPositions = append(updatedBulletPositions, position)
		}
	}

	gv.bulletPositions = updatedBulletPositions
}

// todo: handle collision between enemy bullet and player bullet (so player can shoot enemy bullets)
func (gv *gameView) handleEnemyBulletCollisions() {
	updatedBulletPositions := make([]vector2d, 0, len(gv.enemyBulletPositions))
	for _, bulletPosition := range gv.enemyBulletPositions {
		if gv.playerPosition == bulletPosition {
			gv.livesRemaining--
			if gv.livesRemaining <= 0 {
				gv.gameOver = true
			}
		} else {
			updatedBulletPositions = append(updatedBulletPositions, bulletPosition)
		}
	}
	gv.enemyBulletPositions = updatedBulletPositions
}

func (gv *gameView) draw(model) string {
	border := lipgloss.RoundedBorder()
	style := lipgloss.NewStyle().
		BorderForeground(accentColor).
		BorderStyle(border).
		Padding(0, 1)

	outputMatrix := newOutputMatrix()
	gv.drawEnemies(&outputMatrix)
	gv.drawBullets(&outputMatrix)
	gv.drawEnemyBullets(&outputMatrix)
	gv.drawPlayer(&outputMatrix)

	mainString := outputMatrixToString(outputMatrix)
	scoreString := fmt.Sprintf("Score: %d", gv.score)
	livesString := fmt.Sprintf("Lives: %d", gv.livesRemaining)
	var statusString string
	if gv.gameOver {
		statusString = "Game over! Press Enter to restart..."
	} else if gv.paused {
		statusString = "Paused; press P to resume..."
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.Render(mainString),
		fmt.Sprintf("%s; %s", scoreString, livesString),
		lipgloss.NewStyle().PaddingTop(1).Render(statusString),
	)
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

func (gv *gameView) drawBullets(outputMatrix *[][]rune) {
	for _, position := range gv.bulletPositions {
		(*outputMatrix)[position.y][position.x] = '.'
	}
}

func (gv *gameView) drawEnemyBullets(outputMatrix *[][]rune) {
	for _, position := range gv.enemyBulletPositions {
		(*outputMatrix)[position.y][position.x] = '.'
	}
}

func (gv *gameView) drawPlayer(outputMatrix *[][]rune) {
	(*outputMatrix)[gv.playerPosition.y][gv.playerPosition.x] = '*'
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
