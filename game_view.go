package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// todo: decide on value properly later
var gameViewSize = vector2d{x: 50, y: 15}

const enemySpacing = 1
const enemyColumnCount = 10

type gameView struct {
	playerPosition  vector2d
	enemyPositions  map[int]map[int]struct{}
	enemyYOffset    int // Easier to just store this instead of traversing through the map to find the min y value
	tickCount       int
	bulletPositions []vector2d
	score           int
	gameOver        bool
}

func newGameView() *gameView {
	return &gameView{
		playerPosition:  vector2d{x: gameViewSize.x / 2, y: gameViewSize.y - 1},
		enemyPositions:  generateEnemyPositions(),
		tickCount:       0,
		bulletPositions: make([]vector2d, 0),
		score:           0,
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
			return m, nil
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
		}
	case bulletTickMsg:
		if !gv.gameOver {
			gv.updateBullets()
			gv.handleCollisions()
		}
		return m, bulletTickCmd()
	case enemyTickMsg:
		if !gv.gameOver {
			gv.updateEnemies()
			gv.handleCollisions()
		}
		return m, enemyTickCmd()
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

func (gv *gameView) handleCollisions() {
	updatedBulletPositions := make([]vector2d, 0, len(gv.bulletPositions))
	for _, position := range gv.bulletPositions {
		xMap, yIsPresent := gv.enemyPositions[position.y]
		_, xIsPresent := xMap[position.x]
		collision := yIsPresent && xIsPresent
		if collision {
			delete(xMap, position.x)
			gv.score++
		} else {
			updatedBulletPositions = append(updatedBulletPositions, position)
		}
	}

	gv.bulletPositions = updatedBulletPositions
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
	gv.drawPlayer(&outputMatrix)

	mainString := outputMatrixToString(outputMatrix)
	scoreString := fmt.Sprintf("Score: %d", gv.score)
	var gameOverString string
	if gv.gameOver {
		gameOverString = lipgloss.NewStyle().PaddingTop(1).Render("Game over! Press Enter to restart...")
	}

	return lipgloss.JoinVertical(lipgloss.Left, style.Render(mainString), scoreString, gameOverString)
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
