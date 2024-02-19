package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type FlipRemove int

const (
	Flip FlipRemove = iota
	Remove
)

var (
	WallIsLiberty = true
	FlipOrRemove  = Flip
)

func main() {

	boardSize := flag.Int("size", 9, "size of the board")
	doPlay := flag.Bool("play", false, "play the game")
	doFoo := flag.Bool("foo", false, "run foo")
	doRemove := flag.Bool("remove", false, "remove stones")
	doFlip := flag.Bool("flip", false, "flip stones")
	doWallIsLiberty := flag.Bool("wall-is-liberty", false, "wall is a liberty")
	doWallIsNotLiberty := flag.Bool("wall-is-not-liberty", false, "wall is not a liberty")
	flag.Parse()

	if *doRemove {
		FlipOrRemove = Remove
	}
	if *doFlip {
		FlipOrRemove = Flip
	}
	if *doWallIsLiberty {
		WallIsLiberty = true
	}
	if *doWallIsNotLiberty {
		WallIsLiberty = false
	}

	switch {
	case *doPlay:
		board := NewBoard(*boardSize)
		play(board)
	case *doFoo:
		foobar()
	default:
		log.Fatalf("usage: %s -foo|-play (-size N)", os.Args[0])
	}
}

func prompt() {
	fmt.Print("Enter move (color x y): ")
}

func play(board Board) {

	scanner := bufio.NewScanner(os.Stdin)
	var lastColor Stone
	for {
		board.Print()
		prompt()
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		var color string
		var xStr string
		var y int
		fmt.Sscanf(line, "%s %s %d", &color, &xStr, &y)

		switch color {
		case "//", "#":
			continue // comment
		case "g", "groups":
			board.PrintGroups()
			continue
		case "q", "quit":
			return
		case "e", "eval", "evaluate":
			board = board.FlipOrRemoveAfter(lastColor)
			continue
		}

		var x int
		if len(xStr) >= 1 {
			x = int(xStr[0] - 'a')
		}
		if !board.IsValid(Location{x, y}) {
			continue
		}

		switch color {
		case "b", "black", "blk":
			board = board.PutAndMerge(Black, Location{x, y})
			lastColor = Black
		case "w", "white", "wht":
			board = board.PutAndMerge(White, Location{x, y})
			lastColor = White
		default:
		}
	}

}

func foobar() {
	b := NewBoard(9)
	b = b.Put(Black, Location{0, 0})
	b = b.Put(Black, Location{0, 1})
	b = b.Put(White, Location{5, 5})
	b = b.Put(White, Location{6, 5})
	b = b.Put(White, Location{5, 6})
	b = b.Put(White, Location{7, 5})
	b = b.Put(White, Location{7, 7})

	b.Print()

	for _, loc := range b.AllLocations() {
		b = b.MergeNeighborsAt(loc)
	}

	b.PrintGroups()
}

func (b Board) PrintGroups() {
	for loc, group := range b.Groups {
		if len(group.Members) > 1 {
			fmt.Println("\ngroup at ", loc, ", has liberty? ", b.HasLiberty(group))
			b.PrintOnlyGroup(group)
		}
	}
}

type Set[T comparable] map[T]bool

type Stone int

const (
	Empty Stone = iota
	Black
	White
)

func (s Stone) String() string {
	switch s {
	case Black:
		return "⏺"
	case White:
		return "◯"
	default:
		return "·"
	}
}

type Location struct {
	X, Y int
}

func (l Location) String() string {
	return fmt.Sprintf("(%c, %d)", 'a'+l.X, l.Y)
}

func (l Location) IsBefore(other Location) bool {
	if l.X < other.X {
		return true
	}
	return l.Y < other.Y
}

type Board struct {
	Size    int
	Points  [][]Stone
	Groups  map[Location]Group    // TopLeft -> Group, i.e. all the groups of this board
	OfGroup map[Location]Location // Location -> TopLeft, i.e. which group this location belongs to
}

func NewBoard(size int) Board {
	points := make([][]Stone, size)
	for i := range points {
		points[i] = make([]Stone, size)
	}
	return Board{
		Size:    size,
		Points:  points,
		Groups:  make(map[Location]Group),
		OfGroup: make(map[Location]Location),
	}
}

func (b Board) Merge(g1, g2 Group) Board {
	if g1.TopLeft == g2.TopLeft {
		return b
	}

	if g2.TopLeft.IsBefore(g1.TopLeft) {
		g1, g2 = g2, g1
	}

	for loc := range g2.Members {
		g1.Members[loc] = true
		b.OfGroup[loc] = g1.TopLeft
	}

	delete(b.Groups, g2.TopLeft)

	return b
}

func (b Board) AllLocations() []Location {
	locations := make([]Location, 0, b.Size*b.Size)
	for y := 0; y < b.Size; y++ {
		for x := 0; x < b.Size; x++ {
			locations = append(locations, Location{x, y})
		}
	}
	return locations
}

func (b Board) IsValid(loc Location) bool {
	return loc.X >= 0 && loc.X < b.Size &&
		loc.Y >= 0 && loc.Y < b.Size
}

func (b Board) NeighborLocations(loc Location) []Location {
	neighbors := []Location{
		{loc.X - 1, loc.Y},
		{loc.X + 1, loc.Y},
		{loc.X, loc.Y - 1},
		{loc.X, loc.Y + 1},
	}

	validNeighbors := make([]Location, 0, 4)
	for _, n := range neighbors {
		if b.IsValid(n) {
			validNeighbors = append(validNeighbors, n)
		}
	}

	return validNeighbors
}

func (b Board) IsBesideWall(loc Location) bool {
	return loc.X == 0 || loc.Y == 0 || loc.X == b.Size-1 || loc.Y == b.Size-1
}

func (b Board) StoneAt(loc Location) Stone {
	return b.Points[loc.Y][loc.X]
}

func (b Board) MergeNeighborsAt(loc Location) Board {
	thisStone := b.StoneAt(loc)
	if thisStone == Empty {
		return b
	}

	neighbors := b.NeighborLocations(loc)

	for _, n := range neighbors {
		neighborStone := b.StoneAt(n)
		if thisStone == neighborStone { // both black, or both white
			g1 := b.Groups[b.OfGroup[loc]]
			g2 := b.Groups[b.OfGroup[n]]
			b = b.Merge(g1, g2)
		}
	}

	return b
}

func (b Board) PutAndMerge(stone Stone, loc Location) Board {
	b = b.Put(stone, loc)
	return b.MergeNeighborsAt(loc)
}

func (b Board) Put(stone Stone, loc Location) Board {
	b.Points[loc.Y][loc.X] = stone
	b.Groups[loc] = NewGroup(loc, stone)
	b.OfGroup[loc] = loc
	return b
}

func (b Board) Print() {
	fmt.Print("\n   ")
	for i := 0; i < b.Size; i++ {
		fmt.Printf("%c ", 'a'+i)
	}
	fmt.Println()

	for i, row := range b.Points {
		fmt.Printf("%2d ", i)
		for _, stone := range row {
			fmt.Print(stone, " ")
		}
		fmt.Println()
	}
	fmt.Println()
}

func (b Board) PrintOnlyGroup(g Group) {

	fmt.Print("\n   ")
	for i := 0; i < b.Size; i++ {
		fmt.Printf("%c ", 'a'+i)
	}
	fmt.Println()

	for y, row := range b.Points {
		fmt.Printf("%2d ", y)
		for x, stone := range row {
			if g.Members[Location{x, y}] {
				fmt.Print(stone, " ")
			} else {
				fmt.Print(Empty, " ")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func (b Board) HasLiberty(g Group) bool {

	for loc := range g.Members {
		if WallIsLiberty && b.IsBesideWall(loc) {
			return true
		}
		neighbors := b.NeighborLocations(loc)
		for _, n := range neighbors {
			if b.StoneAt(n) == Empty {
				return true
			}

		}
	}

	return false
}

type Group struct {
	Color   Stone
	TopLeft Location
	Members Set[Location]
}

func NewGroup(loc Location, color Stone) Group {
	return Group{
		Color:   color,
		TopLeft: loc,
		Members: Set[Location]{loc: true},
	}
}

func (b Board) FlipOrRemoveAfter(color Stone) Board {
	var otherColor Stone
	switch color {
	case Black:
		otherColor = White
	case White:
		otherColor = Black
	default:
		log.Fatalf("trying to flip/remove without valid color: %v", color)
	}

	otherGroups := []Group{}
	for _, group := range b.Groups {
		if group.Color == otherColor {
			otherGroups = append(otherGroups, group)
		}
	}

	log.Printf("looking for %s groups to flip/remove", otherColor)

	for _, group := range otherGroups {
		if !b.HasLiberty(group) {
			log.Printf("group at %v has no liberty", group.TopLeft)
			switch FlipOrRemove {
			case Flip:
				log.Printf("flipping group at %v", group.TopLeft)
				b = b.FlipGroup(group, color) // flip to the first color
			case Remove:
				log.Printf("removing group at %v", group.TopLeft)
				b = b.RemoveGroup(group)
			}
		}
	}

	log.Printf("done flipping/removing")

	return b
}

func (b Board) FlipGroup(g Group, otherColor Stone) Board {
	for loc := range g.Members {
		b.Points[loc.Y][loc.X] = otherColor
	}
	for loc := range g.Members {
		b = b.MergeNeighborsAt(loc)
	}
	return b
}

func (b Board) RemoveGroup(g Group) Board {
	for loc := range g.Members {
		b.Points[loc.Y][loc.X] = Empty
		b.OfGroup[loc] = Location{-1, -1}
	}
	delete(b.Groups, g.TopLeft)
	return b
}

// Input: board = [["X","X","X","X"],["X","O","O","X"],["X","X","O","X"],["X","O","X","X"]]
// Output: [["X","X","X","X"],["X","X","X","X"],["X","X","X","X"],["X","O","X","X"]]
// Explanation: Notice that an 'O' should not be flipped if:
// - It is on the border, or
// - It is adjacent to an 'O' that should not be flipped.
// The bottom 'O' is on the border, so it is not flipped.
// The other three 'O' form a surrounded region, so they are flipped.

// func solve(board [][]byte) {

// }
