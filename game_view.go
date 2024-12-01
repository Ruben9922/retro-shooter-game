package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"slices"
	"strings"
)

// todo: decide on value properly later
var gameViewSize = vector2d{x: 50, y: 15}

const enemySpacing = 1
const enemyColumnCount = 10
const scorePerHit = 1

type gameView struct {
	playerPosition     vector2d
	enemyPositions     [][]vector2d
	tickCount          int
	enemyUpdateFlipped bool
	bulletPositions    []vector2d
	score              int
}

func newGameView() *gameView {
	return &gameView{
		playerPosition:     vector2d{x: gameViewSize.x / 2, y: gameViewSize.y - 1},
		enemyPositions:     generateEnemyPositions(),
		tickCount:          0,
		enemyUpdateFlipped: false,
		bulletPositions:    make([]vector2d, 0),
		score:              0,
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
		gv.updateBullets()
		gv.handleCollisions()
		return m, bulletTickCmd()
	case enemyTickMsg:
		gv.updateEnemies()
		gv.handleCollisions()
		return m, enemyTickCmd()
	}
	return m, nil
}

func (gv *gameView) updateEnemies() {
	for i := range gv.enemyPositions {
		row := &gv.enemyPositions[i]
		for j := range *row {
			position := &(*row)[j]
			// XOR `position.y%2 == 0` with `!gv.enemyUpdateFlipped`
			if (position.y%2 == 0) == gv.enemyUpdateFlipped {
				position.x--
			} else {
				position.x++
			}
		}
	}

	gv.tickCount++
	if gv.tickCount >= gameViewSize.x-enemyColumnCount-(enemySpacing*(enemyColumnCount-1)) {
		gv.tickCount = 0
		gv.enemyUpdateFlipped = !gv.enemyUpdateFlipped
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
	// todo: not needed
	// Copy `updatedBulletPositions`
	//updatedBulletPositions := make([]vector2d, len(gv.bulletPositions))
	//copy(updatedBulletPositions, gv.bulletPositions)

	// Copy `gv.enemyPositions`
	//updatedEnemyPositions := make([][]vector2d, 0, len(gv.enemyPositions))
	//for _, row := range updatedEnemyPositions {
	//	updatedRow := make([]vector2d, len(row))
	//	copy(updatedRow, row)
	//	updatedEnemyPositions = append(updatedEnemyPositions, updatedRow)
	//}

	// todo: can probably just remove this!
	// Don't want indices to be created in another thread until the current thread's indices have been deleted
	// E.g. 1st thread wants to delete index 1, 2nd thread wants to delete index 2, 1st thread deletes index 1, 2nd
	// thread might error if index 2 no longer exists as the elements has been shifted left by 1st thread
	//gv.mu.Lock()
	//defer gv.mu.Unlock()

	// todo: can simplify by setting positions to some "blank" value, then removing them (don't need to store deleted indices)
	// todo: kinda inefficient, probably want to improve this e.g. using maps
	// todo: potentially improve this - basically storing indices of positions to remove then removing them afterwards
	bulletPositionsDeleted := make([]int, 0, len(gv.bulletPositions))
	enemyPositionsDeleted := make([][]int, 0, len(gv.enemyPositions))
	for _, row := range gv.enemyPositions {
		enemyPositionsDeleted = append(enemyPositionsDeleted, make([]int, 0, len(row)))
	}
	for bulletPositionIndex, bulletPosition := range gv.bulletPositions {
		y := bulletPosition.y

		// Needed because bullet can be anywhere in the matrix, but `gv.enemyPositions` only has indexes for first few rows
		// i.e. because y can be greater than len(gv.enemyPositions)
		if y < 0 || y >= len(gv.enemyPositions) {
			continue
		}

		for enemyPositionIndex, enemyPosition := range gv.enemyPositions[y] {
			if bulletPosition.equals(enemyPosition) {
				enemyPositionsDeleted[y] = append(enemyPositionsDeleted[y], enemyPositionIndex)
				bulletPositionsDeleted = append(bulletPositionsDeleted, bulletPositionIndex)

				gv.score += scorePerHit
			}
		}
	}

	// Filter out deleted bullet positions
	// Don't want to call `slices.Delete` for each deleted index because that shifts subsequent elements left which
	// could make indices out-of-date
	updatedBulletPositions := make([]vector2d, 0, len(gv.bulletPositions))
	for i, position := range gv.bulletPositions {
		if !slices.Contains(bulletPositionsDeleted, i) {
			updatedBulletPositions = append(updatedBulletPositions, position)
		}
	}
	gv.bulletPositions = updatedBulletPositions

	// Filter out deleted enemy positions
	updatedEnemyPositions := make([][]vector2d, 0, len(gv.enemyPositions))
	for y, row := range gv.enemyPositions {
		updatedRow := make([]vector2d, 0, len(row))
		for i, position := range row {
			if !slices.Contains(enemyPositionsDeleted[y], i) {
				updatedRow = append(updatedRow, position)
			}
		}
		updatedEnemyPositions = append(updatedEnemyPositions, updatedRow)
	}
	gv.enemyPositions = updatedEnemyPositions
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
	for y, row := range gv.enemyPositions {
		for _, position := range row {
			(*outputMatrix)[y][position.x] = '$'
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
