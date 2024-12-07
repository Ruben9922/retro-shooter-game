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
	enemyPositions  [][]vector2d
	enemyYOffset    int
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

func generateEnemyPositions() (enemyPositions [][]vector2d) {
	const rowCount = 5
	enemyPositions = make([][]vector2d, 0, rowCount)
	for y := 0; y < rowCount; y++ {
		enemyPositions = append(enemyPositions, make([]vector2d, 0, enemyColumnCount))
		for columnIndex := 0; columnIndex < enemyColumnCount; columnIndex++ {
			var x int
			if y%2 == 0 {
				x = columnIndex * 2
			} else {
				x = gameViewSize.x - (columnIndex * 2) - 1
			}
			enemyPositions[y] = append(enemyPositions[y], vector2d{x: x, y: y})
		}
	}
	return
}

func (gv *gameView) update(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
			for i := range gv.enemyPositions {
				row := &gv.enemyPositions[i]
				for j := range *row {
					position := &(*row)[j]
					position.y++
					gv.enemyYOffset++
				}
			}
		}
	} else {
		gv.tickCount++

		// Move enemies left/right (move alternate rows in opposite directions, so 1st row left/right, then 2nd row right/left, etc.)
		for i := range gv.enemyPositions {
			row := &gv.enemyPositions[i]
			for j := range *row {

				position := &(*row)[j]
				if position.y%2 == 0 {
					position.x++
				} else {
					position.x--
				}
			}
		}
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
	// Instead of nested loops, create a map for the enemies, iterate through the bullet and do a lookup on the map
	enemyPositionMap := make(map[int]map[int]struct{})

	// Convert enemy slice into map
	for _, row := range gv.enemyPositions {
		for _, position := range row {
			if _, yIsPresent := enemyPositionMap[position.y]; !yIsPresent {
				enemyPositionMap[position.y] = make(map[int]struct{})
			}

			enemyPositionMap[position.y][position.x] = struct{}{}
		}
	}

	updatedBulletPositions := make([]vector2d, 0, len(gv.bulletPositions))
	for _, position := range gv.bulletPositions {
		xMap, yIsPresent := enemyPositionMap[position.y]
		_, xIsPresent := xMap[position.x]
		collision := yIsPresent && xIsPresent
		if collision {
			delete(xMap, position.x)
			gv.score++
		} else {
			updatedBulletPositions = append(updatedBulletPositions, position)
		}
	}

	// Convert enemy map back into a slice
	updatedEnemyPositions := make([][]vector2d, 5)
	for i := range updatedEnemyPositions {
		updatedEnemyPositions[i] = make([]vector2d, 0)
	}
	for y, xMap := range enemyPositionMap {
		for x := range xMap {
			enemyRowIndex := y - gv.enemyYOffset
			enemyRow := &updatedEnemyPositions[enemyRowIndex]
			position := vector2d{x: x, y: y}
			*enemyRow = append(*enemyRow, position)
		}
	}

	gv.enemyPositions = updatedEnemyPositions
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

	return lipgloss.JoinVertical(lipgloss.Right, style.Render(mainString), scoreString)
}

func newOutputMatrix() (outputMatrix [][]rune) {
	outputMatrix = make([][]rune, gameViewSize.y)
	for i := range outputMatrix {
		outputMatrix[i] = make([]rune, gameViewSize.x)
	}
	return
}

func (gv *gameView) drawEnemies(outputMatrix *[][]rune) {
	for _, row := range gv.enemyPositions {
		for _, position := range row {
			(*outputMatrix)[position.y][position.x] = '$'
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
